package test

import (
	"github.com/Bofry/trace"
	postgres "github.com/Bofry/worker-postgres"
	"github.com/Bofry/worker-postgres/internal"
)

var (
	_ postgres.MessageHandler = new(InvalidMessageHandler)
)

type InvalidMessageHandler struct {
	ServiceProvider *ServiceProvider
}

// ProcessMessage implements internal.MessageHandler.
func (i *InvalidMessageHandler) ProcessMessage(ctx *internal.Context, message *postgres.Message) {
	ctx.Logger().Printf("InvalidMessageHandler.Message on %s (%s): [%s] %v\n", message.Slot, message.StartLSN(), message.StartLSN(), string(message.Body()))

	sp := trace.SpanFromContext(ctx)
	sp.Argv(message)
}
