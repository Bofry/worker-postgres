package test

import (
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
func (i *InvalidMessageHandler) ProcessMessage(ctx *internal.Context, message *postgres.Message) error {
	panic("unimplemented")
}
