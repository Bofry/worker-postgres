package internal

import postgres "github.com/Bofry/lib-postgres-stream"

var _ postgres.MessageDelegate = RestrictedMessageDelegate(0)

type RestrictedMessageDelegate int

// OnAck implements postgres.MessageDelegate.
func (r RestrictedMessageDelegate) OnAck(msg *postgres.Message) {
	panic(RestrictedOperationError("Message.Ack() cannot be called by restricted area"))
}
