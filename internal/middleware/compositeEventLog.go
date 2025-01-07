package middleware

import (
	postgres "github.com/Bofry/lib-postgres-stream"
	"github.com/Bofry/worker-postgres/internal"
)

var _ EventLog = CompositeEventLog{}

type CompositeEventLog struct {
	eventLogs []EventLog
}

// Flush implements EventLog.
func (l CompositeEventLog) Flush() {
	for _, log := range l.eventLogs {
		log.Flush()
	}
}

// OnError implements EventLog.
func (l CompositeEventLog) OnError(message *postgres.Message, err interface{}, stackTrace []byte) {
	for _, log := range l.eventLogs {
		log.OnError(message, err, stackTrace)
	}
}

// OnProcessMessage implements EventLog.
func (l CompositeEventLog) OnProcessMessage(message *postgres.Message) {
	for _, log := range l.eventLogs {
		log.OnProcessMessage(message)
	}
}

// OnProcessMessageComplete implements EventLog.
func (l CompositeEventLog) OnProcessMessageComplete(message *postgres.Message, reply internal.ReplyCode) {
	for _, log := range l.eventLogs {
		log.OnProcessMessageComplete(message, reply)
	}
}
