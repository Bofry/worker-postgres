package middleware

import (
	postgres "github.com/Bofry/lib-postgres-stream"
	"github.com/Bofry/worker-postgres/internal"
)

var _ EventLog = NoopEventLog(0)

type NoopEventLog int

// Flush implements EventLog.
func (n NoopEventLog) Flush() {}

// OnError implements EventLog.
func (n NoopEventLog) OnError(message *postgres.Message, err interface{}, stackTrace []byte) {}

// OnProcessMessage implements EventLog.
func (n NoopEventLog) OnProcessMessage(message *postgres.Message) {}

// OnProcessMessageComplete implements EventLog.
func (n NoopEventLog) OnProcessMessageComplete(message *postgres.Message, reply internal.ReplyCode) {}
