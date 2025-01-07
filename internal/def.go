package internal

import (
	"context"
	"log"
	"os"
	"reflect"

	postgres "github.com/Bofry/lib-postgres-stream"
	"github.com/Bofry/trace"
	"go.opentelemetry.io/otel/propagation"
)

const (
	LOGGER_PREFIX string = "[worker-postgres] "

	__CONTEXT_REPLY_KEY ctxReplyKeyType = 0

	__UNDEFINED_TRACER_NAME     = "undefined"
	__INVALID_MESSAGE_SPAN_NAME = "InvalidMessage"

	__INVALID_MESSAGE_HANDLER_NAME = "InvalidMessageHandler"
)

const (
	UNSET ReplyCode = iota
	PASS
	FAIL
	ABORT

	__reply_code_minimum__ = UNSET
	__reply_code_maximum__ = ABORT

	INVALID ReplyCode = -1

	__reply_code_invalid_text__ = "invalid"
)

const (
	RestrictedForwardMessage_InvalidOperation int = 0
	RestrictedForwardMessage_Recursive        int = 1
)

var (
	typeOfHost                  = reflect.TypeOf(PostgresWorker{})
	typeOfMessageHandler        = reflect.TypeOf((*MessageHandler)(nil)).Elem()
	typeOfMessageObserverAffair = reflect.TypeOf((*MessageObserverAffair)(nil)).Elem()
	defaultTracerProvider       = createNoopTracerProvider()
	defaultTextMapPropagator    = createNoopTextMapPropagator()
	defaultMessageDelegate      = NoopMessageDelegate(0)

	GlobalTracerManager             *TracerManager           // be register from PostgresWorker
	GlobalContextHelper             ContextHelper            = ContextHelper{}
	GlobalRestrictedMessageDelegate postgres.MessageDelegate = RestrictedMessageDelegate(0)
	GlobalNoopMessageDelegate       postgres.MessageDelegate = NoopMessageDelegate(0)
	GlobalMessageDelegateHelper     MessageDelegateHelper    = MessageDelegateHelper{}

	PostgresWorkerModuleInstance = PostgresWorkerModule{}

	PostgresWorkerLogger *log.Logger = log.New(os.Stdout, LOGGER_PREFIX, log.LstdFlags|log.Lmsgprefix)
)

type (
	ctxReplyKeyType int

	StatusCode = ReplyCode

	Config                              = postgres.Config
	Event                               = postgres.Event
	Message                             = postgres.Message
	SlotOffset                          = postgres.SlotOffset
	SlotOffsetInfo                      = postgres.SlotOffsetInfo
	CreateReplicationSlotSource         = postgres.CreateReplicationSlotSource
	CreateReplicationSlotSourceProvider = postgres.CreateReplicationSlotSourceProvider

	MessageHandleModule interface {
		CanSetSuccessor() bool
		SetSuccessor(successor MessageHandleModule)
		ProcessMessage(ctx *Context, message *Message, state ProcessingState, recover *Recover) error
		OnInitComplete()
		OnStart(ctx context.Context) error
		OnStop(ctx context.Context) error
	}

	MessageObserver interface {
		OnAck(ctx *Context, message *Message)
		Type() reflect.Type
	}

	MessageObserverAffair interface {
		MessageObserverTypes() []reflect.Type
	}

	MessageHandler interface {
		ProcessMessage(ctx *Context, message *Message) error
	}

	ErrorHandler func(ctx *Context, message *Message, err interface{})

	OnHostErrorHandler func(err error) (disposed bool)
)

func createNoopTracerProvider() *trace.SeverityTracerProvider {
	tp, err := trace.NoopProvider()
	if err != nil {
		PostgresWorkerLogger.Fatalf("cannot create NoopProvider: %v", err)
	}
	return tp
}

func createNoopTextMapPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator()
}
