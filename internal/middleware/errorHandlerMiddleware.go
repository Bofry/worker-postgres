package middleware

import (
	"github.com/Bofry/host"
	. "github.com/Bofry/worker-postgres/internal"
)

var _ host.Middleware = new(ErrorHandlerMiddleware)

type ErrorHandlerMiddleware struct {
	Handler ErrorHandler
}

// Init implements internal.Middleware.
func (m *ErrorHandlerMiddleware) Init(app *host.AppModule) {
	var (
		worker    = asPostgresWorker(app.Host())
		registrar = NewPostgresWorkerRegistrar(worker)
	)

	registrar.SetErrorHandler(m.Handler)
}
