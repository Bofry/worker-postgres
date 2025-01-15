package internal

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/Bofry/host"
	postgres "github.com/Bofry/lib-postgres-stream"
	"github.com/Bofry/structproto/reflecting"
	"github.com/Bofry/trace"
	"go.opentelemetry.io/otel/propagation"
)

var _ host.Host = new(PostgresWorker)

type PostgresWorker struct {
	Config                        *Config
	ReplicationSlotSourceProvider CreateReplicationSlotSourceProvider

	consumer *postgres.Consumer

	logger *log.Logger

	messageDispatcher *MessageDispatcher
	messageManager    interface{}

	messageHandleService   *MessageHandleService
	messageTracerService   *MessageTracerService
	messageObserverService *MessageObserverService

	tracerManager *TracerManager

	onErrorEventHandler host.HostOnErrorEventHandler

	wg          sync.WaitGroup
	mutex       sync.Mutex
	initialized bool
	running     bool
	disposed    bool
}

// Start implements internal.Host.
func (w *PostgresWorker) Start(ctx context.Context) {
	if w.disposed {
		PostgresWorkerLogger.Panic("the Worker has been disposed")
	}
	if !w.initialized {
		PostgresWorkerLogger.Panic("the Worker havn't be initialized yet")
	}
	if w.running {
		return
	}

	var err error
	w.mutex.Lock()
	defer func() {
		if err != nil {
			w.running = false
			w.disposed = true
		}
		w.mutex.Unlock()
	}()

	w.running = true
	w.messageDispatcher.start(ctx)

	var (
		slots = w.messageDispatcher.Slots()
	)

	err = w.registerSlots()
	if err != nil {
		PostgresWorkerLogger.Panic(err)
	}

	if len(slots) > 0 {
		err = w.messageDispatcher.subscribe(w.consumer)
		if err != nil {
			PostgresWorkerLogger.Panic(err)
		}
	}
}

// Stop implements internal.Host.
func (w *PostgresWorker) Stop(ctx context.Context) error {
	PostgresWorkerLogger.Printf("%% Stopping\n")

	w.mutex.Lock()
	defer func() {
		w.running = false
		w.disposed = true
		w.mutex.Unlock()

		w.messageDispatcher.stop(ctx)

		PostgresWorkerLogger.Printf("%% Stopped\n")
	}()

	w.consumer.Close()
	w.wg.Wait()
	return nil
}

func (w *PostgresWorker) Logger() *log.Logger {
	return w.logger
}

func (w *PostgresWorker) alloc() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.tracerManager = NewTraceManager()
	w.messageHandleService = NewMessageHandleService()

	w.messageTracerService = &MessageTracerService{
		TracerManager: w.tracerManager,
	}
	w.messageObserverService = &MessageObserverService{
		MessageObservers: make(map[reflect.Type]MessageObserver),
	}

	w.messageDispatcher = &MessageDispatcher{
		MessageHandleService:   w.messageHandleService,
		MessageTracerService:   w.messageTracerService,
		MessageObserverService: w.messageObserverService,
		Router:                 make(Router),
		SlotSet:                make(map[string]SlotOffset),
		OnHostErrorProc:        w.onHostError,
	}

	// register TracerManager
	GlobalTracerManager = w.tracerManager
}

func (w *PostgresWorker) init() {
	if w.initialized {
		return
	}

	w.mutex.Lock()
	defer func() {
		w.initialized = true
		w.mutex.Unlock()
	}()

	var invalidMessageHandler = w.messageDispatcher.InvalidMessageHandler
	if w.messageDispatcher.InvalidMessageHandler == nil {
		handler, err := w.findInvalidMessageHandler()
		if err != nil {
			panic(handler)
		}
		registrar := NewPostgresWorkerRegistrar(w)
		registrar.SetInvalidMessageHandler(handler)

		invalidMessageHandler = handler
	}

	w.messageTracerService.init(w.messageManager, invalidMessageHandler)
	w.messageObserverService.init(w.messageManager)
	w.messageDispatcher.init()
	w.configConsumer()
}

func (w *PostgresWorker) registerSlots() error {
	w.logger.Printf("ReplicationSlotSourceProvider:: %+v\n", w.ReplicationSlotSourceProvider.Sources())

	conn, err := postgres.NewConn(w.Config)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	w.logger.Printf("CreateReplicationSlot")
	err = postgres.CreateReplicationSlot(context.Background(), conn, w.ReplicationSlotSourceProvider)
	if err != nil {
		if !postgres.IsDuplicateObjectError(err) {
			return err
		}
	}
	return nil
}

func (w *PostgresWorker) configConsumer() {
	instance := &postgres.Consumer{
		Config:         w.Config,
		Logger:         w.logger,
		MessageHandler: w.receiveMessage,
		EventHandler:   nil,
		ErrorHandler:   nil,
	}

	w.consumer = instance
}

func (w *PostgresWorker) receiveMessage(message *Message) {
	ctx := &Context{
		consumer:              w.consumer,
		logger:                w.logger,
		invalidMessageHandler: nil, // be determined by MessageDispatcher
	}

	// configure nsq.MessageDelegate
	if message.Delegate == nil {
		message.Delegate = defaultMessageDelegate
	}
	delegate := NewContextMessageDelegate(ctx)
	delegate.configure(message)

	w.messageDispatcher.ProcessMessage(ctx, message)
}

func (w *PostgresWorker) onHostError(err error) (disposed bool) {
	if w.onErrorEventHandler != nil {
		return w.onErrorEventHandler.OnError(err)
	}
	return false
}

func (w *PostgresWorker) setTextMapPropagator(propagator propagation.TextMapPropagator) {
	w.messageTracerService.textMapPropagator = propagator
}

func (w *PostgresWorker) setTracerProvider(provider *trace.SeverityTracerProvider) {
	w.messageTracerService.tracerProvider = provider
}

func (w *PostgresWorker) setLogger(l *log.Logger) {
	w.logger = l
}

func (w *PostgresWorker) findInvalidMessageHandler() (MessageHandler, error) {
	var (
		handler MessageHandler

		rvManager reflect.Value = reflect.ValueOf(w.messageManager)
	)
	if rvManager.Kind() != reflect.Pointer || rvManager.IsNil() {
		return nil, nil
	}

	rvManager = reflect.Indirect(rvManager)
	numOfHandles := rvManager.NumField()
	for i := 0; i < numOfHandles; i++ {
		rvHandler := rvManager.Field(i)

		// is pointer ?
		if rvHandler.Kind() != reflect.Pointer {
			continue
		}
		// is MessageHandler ?
		if !IsMessageHandlerType(rvHandler.Type()) {
			continue
		}

		if rvHandler.Type().Elem().Name() == __INVALID_MESSAGE_HANDLER_NAME {
			if rvHandler.IsNil() {
				rvHandler = reflecting.AssignZero(rvHandler)

				// initialize
				rv := reflect.Indirect(rvHandler)
				if rv.CanAddr() {
					rv = rv.Addr()
					// call MessageHandler.Init()
					fn := rv.MethodByName(host.APP_COMPONENT_INIT_METHOD)
					if fn.IsValid() {
						if fn.Kind() != reflect.Func {
							return nil, fmt.Errorf("fail to Init() resource. cannot find func %s() within type %s\n", host.APP_COMPONENT_INIT_METHOD, rv.Type().String())
						}
						if fn.Type().NumIn() != 0 || fn.Type().NumOut() != 0 {
							return nil, fmt.Errorf("fail to Init() resource. %s.%s() type should be func()\n", rv.Type().String(), host.APP_COMPONENT_INIT_METHOD)
						}
						fn.Call([]reflect.Value(nil))
					}
				}
			}

			handler = AsMessageHandler(rvHandler)
			break
		}
	}
	return handler, nil
}
