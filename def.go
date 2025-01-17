package postgres

import (
	postgres "github.com/Bofry/lib-postgres-stream"
	"github.com/Bofry/worker-postgres/internal"
	"github.com/Bofry/worker-postgres/internal/middleware"
)

const (
	StatusInvalid = internal.INVALID
	StatusUnset   = internal.UNSET
	StatusPass    = internal.PASS
	StatusFail    = internal.FAIL
	StatusAbort   = internal.ABORT
)

const (
	LogicalReplication  = postgres.LogicalReplication
	PhysicalReplication = postgres.PhysicalReplication
)

type (
	Message           = postgres.Message
	Config            = postgres.Config
	ReplicationOption = postgres.ReplicationOption

	EventEvidence  = middleware.EventEvidence
	LoggingService = middleware.LoggingService
	EventLog       = middleware.EventLog

	MessageObserver       = internal.MessageObserver
	MessageObserverAffair = internal.MessageObserverAffair

	MessageHandler      = internal.MessageHandler
	MessageErrorHandler = internal.MessageErrorHandler
	Worker              = internal.PostgresWorker
	Context             = internal.Context
	ReplyCode           = internal.ReplyCode

	ErrorHandler = internal.ErrorHandler
)

func ConfigureReplicationOptions() postgres.ReplicationOptions {
	return postgres.ConfigureReplicationOptions()
}
