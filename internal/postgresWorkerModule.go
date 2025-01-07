package internal

import (
	"io"
	"reflect"

	"github.com/Bofry/host"
	"github.com/Bofry/host/helper"
)

var _ host.HostModule = PostgresWorkerModule{}

type PostgresWorkerModule struct{}

// ConfigureLogger implements internal.HostModule.
func (p PostgresWorkerModule) ConfigureLogger(logflags int, w io.Writer) {
	PostgresWorkerLogger.SetFlags(logflags)
	PostgresWorkerLogger.SetOutput(w)
}

// DescribeHostType implements internal.HostModule.
func (p PostgresWorkerModule) DescribeHostType() reflect.Type {
	return typeOfHost
}

// Init implements internal.HostModule.
func (p PostgresWorkerModule) Init(h host.Host, app *host.AppModule) {
	if v, ok := h.(*PostgresWorker); ok {
		v.alloc()
		v.setTracerProvider(app.TracerProvider())
		v.setTextMapPropagator(app.TextMapPropagator())
		v.setLogger(app.Logger())

		{
			host := helper.HostHelper(app)
			v.onErrorEventHandler = host.OnErrorEventHandler()
		}
	}
}

// InitComplete implements internal.HostModule.
func (p PostgresWorkerModule) InitComplete(h host.Host, app *host.AppModule) {
	if v, ok := h.(*PostgresWorker); ok {
		v.init()
	}
}
