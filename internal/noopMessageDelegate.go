package internal

import postgres "github.com/Bofry/lib-postgres-stream"

var _ postgres.MessageDelegate = NoopMessageDelegate(0)

type NoopMessageDelegate int

// OnAck implements postgres.MessageDelegate.
func (n NoopMessageDelegate) OnAck(msg *postgres.Message) {}
