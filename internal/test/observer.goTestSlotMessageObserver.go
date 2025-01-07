package test

import (
	"fmt"
	"reflect"

	postgres "github.com/Bofry/worker-postgres"
	"github.com/Bofry/worker-postgres/internal"
	"github.com/Bofry/worker-postgres/tracing"
)

var _ postgres.MessageObserver = new(GoTestSlotMessageObserver)

type GoTestSlotMessageObserver struct {
	ServiceProvider *ServiceProvider
}

func (*GoTestSlotMessageObserver) Init() {
	fmt.Println("GoTestStreamMessageObserver.Init()")
}

// OnAck implements internal.MessageObserver.
func (o *GoTestSlotMessageObserver) OnAck(ctx *internal.Context, message *postgres.Message) {
	tr := tracing.GetTracer(o)
	sp := tr.Start(ctx, "OnAck()")
	defer sp.End()

	o.ServiceProvider.Logger().Println("GoTestSlotMessageObserver.OnAck()")
}

// Type implements internal.MessageObserver.
func (o *GoTestSlotMessageObserver) Type() reflect.Type {
	return reflect.TypeOf(o)
}
