package middleware

import (
	"github.com/Bofry/host"
	. "github.com/Bofry/worker-postgres/internal"
)

var _ host.Middleware = new(LoggingMiddleware)

type LoggingMiddleware struct {
	LoggingService LoggingService
}

// Init implements internal.Middleware
func (m *LoggingMiddleware) Init(app *host.AppModule) {
	var (
		worker    = asPostgresWorker(app.Host())
		registrar = NewPostgresWorkerRegistrar(worker)
	)

	m.LoggingService.ConfigureLogger(worker.Logger())

	loggingHandleModule := &LoggingHandleModule{
		loggingService: m.LoggingService,
	}
	registrar.RegisterMessageHandleModule(loggingHandleModule)
}
