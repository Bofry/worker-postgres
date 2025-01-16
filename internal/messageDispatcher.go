package internal

import (
	"context"
	"errors"
	"fmt"

	postgres "github.com/Bofry/lib-postgres-stream"
	"github.com/Bofry/trace"
)

type MessageDispatcher struct {
	MessageHandleService   *MessageHandleService
	MessageTracerService   *MessageTracerService
	MessageObserverService *MessageObserverService
	Router                 Router

	OnHostErrorProc OnHostErrorHandler

	ErrorHandler          ErrorHandler
	InvalidMessageHandler MessageHandler

	SlotSet map[string]SlotOffset
}

func (d *MessageDispatcher) SlotOffsets() []SlotOffset {
	var (
		slots = d.SlotSet
	)

	offsets := make([]SlotOffset, 0, len(slots))
	for _, v := range slots {
		offsets = append(offsets, v)
	}
	return offsets
}

func (d *MessageDispatcher) Slots() []string {
	var (
		router = d.Router
	)

	if router != nil {
		keys := make([]string, 0, len(router))
		for k := range router {
			keys = append(keys, k)
		}
		return keys
	}
	return nil
}

func (d *MessageDispatcher) ProcessMessage(ctx *Context, message *Message) {
	// start tracing
	var (
		handlerID = d.Router.FindHandlerComponentID(message.Slot)

		spanName string = message.Slot
		tr       *trace.SeverityTracer
		sp       *trace.SeveritySpan
	)

	tr = d.MessageTracerService.Tracer(handlerID)
	sp = tr.Start(ctx, spanName)
	defer sp.End()

	processingState := ProcessingState{
		Slot:   message.Slot,
		Tracer: tr,
		Span:   sp,
	}

	// set invalidMessageHandler
	ctx.invalidMessageHandler = d.InvalidMessageHandler

	// register observer into message
	d.MessageObserverService.RegisterMessageObservers(message, handlerID)

	d.MessageHandleService.ProcessMessage(ctx, message, processingState, new(Recover))
}

func (d *MessageDispatcher) subscribe(consumer *postgres.Consumer) error {
	var (
		slots   = d.SlotOffsets()
		offsets = make([]SlotOffsetInfo, 0, len(slots))
	)

	for _, slot := range slots {
		offsets = append(offsets, slot)
	}

	return consumer.Subscribe(offsets...)
}

func (d *MessageDispatcher) internalProcessMessage(ctx *Context, message *Message, state ProcessingState, recover *Recover) {
	var handler MessageHandler

	recover.
		Defer(func(err interface{}) {
			if err != nil {
				if handler != nil {
					if h, ok := handler.(MessageErrorHandler); ok {
						h.ProcessMessageError(ctx, message, err)
					}
				}
				if !ctx.aborted {
					d.processError(ctx, message, err)
				}
			}
		}).
		Do(func(finalizer Finalizer) {
			var (
				tr   *trace.SeverityTracer = state.Tracer
				sp   *trace.SeveritySpan   = state.Span
				slot string                = state.Slot
			)
			_ = tr

			// set Span
			trace.SpanToContext(ctx, sp)

			finalizer.Add(func(err interface{}) {
				var (
					reply = GlobalContextHelper.ExtractReplyCode(ctx)
				)

				if err != nil {
					if e, ok := err.(error); ok {
						sp.Err(e)
					} else if e, ok := err.(string); ok {
						sp.Err(errors.New(e))
					} else if e, ok := err.(fmt.Stringer); ok {
						sp.Err(errors.New(e.String()))
					} else {
						sp.Err(fmt.Errorf("%+v", err))
					}

					GlobalContextHelper.InjectReplyCode(ctx, FAIL)
				}

				if reply == UNSET {
					GlobalContextHelper.InjectReplyCode(ctx, PASS)
				}

				switch reply {
				case PASS:
					sp.Reply(trace.PASS, reply)
				case FAIL, ABORT:
					sp.Reply(trace.FAIL, reply)
				}
			})

			sp.Tags(
				// TODO: add redis server version
				trace.Stream(slot),
				trace.MessageID(message.StartLSN().String()),
			)

			handler = d.Router.Get(message.Slot)
			if handler != nil {
				handler.ProcessMessage(ctx, message)
				return
			}
			ctx.InvalidMessage(message)
		})
}

func (d *MessageDispatcher) init() {
	// register the default MessageHandleModule
	stdMessageHandleModule := NewStdMessageHandleModule(d)
	d.MessageHandleService.Register(stdMessageHandleModule)
}

func (d *MessageDispatcher) processError(ctx *Context, message *Message, err interface{}) {
	if d.ErrorHandler != nil {
		d.ErrorHandler(ctx, message, err)
	}
}

func (d *MessageDispatcher) start(ctx context.Context) {
	err := d.MessageHandleService.triggerStart(ctx)
	if err != nil {
		var disposed bool = false
		if d.OnHostErrorProc != nil {
			disposed = d.OnHostErrorProc(err)
		}
		if !disposed {
			PostgresWorkerLogger.Fatalf("%+v", err)
		}
	}
}

func (d *MessageDispatcher) stop(ctx context.Context) {
	for err := range d.MessageHandleService.triggerStop(ctx) {
		if err != nil {
			PostgresWorkerLogger.Printf("%+v", err)
		}
	}
}
