package middleware

import (
	"github.com/Bofry/host"
	. "github.com/Bofry/worker-postgres/internal"
)

var _ host.Middleware = new(TracingMiddleware)

type TracingMiddleware struct {
	Enabled bool
}

// Init implements internal.Middleware.
func (m *TracingMiddleware) Init(app *host.AppModule) {
	var (
		worker    = asPostgresWorker(app.Host())
		registrar = NewPostgresWorkerRegistrar(worker)
	)

	registrar.EnableTracer(m.Enabled)
}
