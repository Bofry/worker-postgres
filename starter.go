package postgres

import (
	"github.com/Bofry/host"
	"github.com/Bofry/worker-postgres/internal"
)

func Startup(app interface{}) *host.Starter {
	var (
		starter = host.Startup(app)
	)

	host.RegisterHostModule(starter, internal.PostgresWorkerModuleInstance)

	return starter
}
