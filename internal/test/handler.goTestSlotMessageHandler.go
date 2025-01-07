package test

import (
	"fmt"
	"reflect"

	"github.com/Bofry/trace"
	postgres "github.com/Bofry/worker-postgres"
	"github.com/Bofry/worker-postgres/internal"
)

var (
	_ postgres.MessageHandler        = new(GoTestSlotMessageHandler)
	_ postgres.MessageObserverAffair = new(GoTestSlotMessageHandler)
)

type GoTestSlotMessageHandler struct {
	ServiceProvider *ServiceProvider

	// counter *GoTestSlotMessageCounter
}

func (h *GoTestSlotMessageHandler) Init() {
	fmt.Println("GoTestTopicMessageHandler.Init()")

	// h.counter = new(GoTestSlotMessageCounter)
}

// ProcessMessage implements internal.MessageHandler.
func (g *GoTestSlotMessageHandler) ProcessMessage(ctx *internal.Context, message *postgres.Message) error {
	ctx.Logger().Printf("Message on %s (%s): [%s] %v\n", message.Slot, message.StartLSN(), message.StartLSN(), string(message.Body()))

	sp := trace.SpanFromContext(ctx)
	sp.Argv(message)

	// message.Ack()

	return nil
}

// MessageObserverTypes implements internal.MessageObserverAffair.
func (g *GoTestSlotMessageHandler) MessageObserverTypes() []reflect.Type {
	return []reflect.Type{
		MessageObserverManager.GoTestSlotMessageObserver.Type(),
	}
}
