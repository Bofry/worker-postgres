package internal

import (
	"reflect"

	postgres "github.com/Bofry/lib-postgres-stream"
)

type PostgresWorkerRegistrar struct {
	worker *PostgresWorker
}

func NewPostgresWorkerRegistrar(worker *PostgresWorker) *PostgresWorkerRegistrar {
	return &PostgresWorkerRegistrar{
		worker: worker,
	}
}

func (r *PostgresWorkerRegistrar) RegisterMessageHandleModule(module MessageHandleModule) {
	r.worker.messageHandleService.Register(module)
}

func (r *PostgresWorkerRegistrar) EnableTracer(enabled bool) {
	r.worker.messageTracerService.Enabled = enabled
}

func (r *PostgresWorkerRegistrar) SetErrorHandler(handler ErrorHandler) {
	r.worker.messageDispatcher.ErrorHandler = handler
}

func (r *PostgresWorkerRegistrar) SetInvalidMessageHandler(handler MessageHandler) {
	r.worker.messageDispatcher.InvalidMessageHandler = handler
}

func (r *PostgresWorkerRegistrar) SetMessageManager(messageManager interface{}) {
	r.worker.messageManager = messageManager
}

func (r *PostgresWorkerRegistrar) RegisterSlot(slotOffset postgres.SlotOffset) {
	r.worker.messageDispatcher.SlotSet[slotOffset.Slot] = slotOffset
}

func (r *PostgresWorkerRegistrar) AddRouter(slot string, handler MessageHandler, handlerComponentID string) {
	r.worker.messageDispatcher.Router.Add(slot, handler, handlerComponentID)
}

func (r *PostgresWorkerRegistrar) RegisterMessageObserver(v MessageObserver) {
	t := reflect.TypeOf(v)
	r.worker.messageObserverService.MessageObservers[t] = v
}
