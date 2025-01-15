package internal

import (
	"context"

	postgres "github.com/Bofry/lib-postgres-stream"
)

var _ MessageHandleModule = new(StdMessageHandleModule)

type StdMessageHandleModule struct {
	dispatcher *MessageDispatcher
}

func NewStdMessageHandleModule(dispatcher *MessageDispatcher) *StdMessageHandleModule {
	return &StdMessageHandleModule{
		dispatcher: dispatcher,
	}
}

// CanSetSuccessor implements MessageHandleModule.
func (*StdMessageHandleModule) CanSetSuccessor() bool {
	return false
}

// OnInitComplete implements MessageHandleModule.
func (*StdMessageHandleModule) OnInitComplete() {
	// ignored
}

// OnStart implements MessageHandleModule.
func (s *StdMessageHandleModule) OnStart(ctx context.Context) error {
	// do nothing
	return nil
}

// OnStop implements MessageHandleModule.
func (s *StdMessageHandleModule) OnStop(ctx context.Context) error {
	// do nothing
	return nil
}

// ProcessMessage implements MessageHandleModule.
func (m *StdMessageHandleModule) ProcessMessage(ctx *Context, message *postgres.Message, state ProcessingState, recover *Recover) {
	m.dispatcher.internalProcessMessage(ctx, message, state, recover)
}

// SetSuccessor implements MessageHandleModule.
func (*StdMessageHandleModule) SetSuccessor(successor MessageHandleModule) {
	panic("unsupported operation")
}
