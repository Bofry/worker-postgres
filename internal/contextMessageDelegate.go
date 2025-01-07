package internal

import (
	"sync"
	"sync/atomic"

	postgres "github.com/Bofry/lib-postgres-stream"
)

var _ postgres.MessageDelegate = new(ContextMessageDelegate)

type ContextMessageDelegate struct {
	parent postgres.MessageDelegate

	ctx *Context

	messageObserver MessageObserver

	restricted int32
	mu         sync.Mutex
}

func NewContextMessageDelegate(ctx *Context) *ContextMessageDelegate {
	return &ContextMessageDelegate{
		ctx: ctx,
	}
}

// OnAck implements postgres.MessageDelegate.
func (d *ContextMessageDelegate) OnAck(msg *postgres.Message) {
	if d.isRestricted() {
		GlobalNoopMessageDelegate.OnAck(nil)
		return
	}

	d.parent.OnAck(msg)
	GlobalContextHelper.InjectReplyCodeSafe(d.ctx, PASS)

	// observer
	if d.messageObserver != nil {
		d.messageObserver.OnAck(d.ctx, msg)
	}
}

func (d *ContextMessageDelegate) configure(msg *Message) {
	if d.parent == nil {
		d.mu.Lock()
		defer d.mu.Unlock()
		if d.parent == nil {
			d.parent = msg.Delegate
			msg.Delegate = d
		}
	}
}

func (d *ContextMessageDelegate) isRestricted() bool {
	return atomic.LoadInt32(&d.restricted) == 1
}

func (d *ContextMessageDelegate) restrict() {
	atomic.StoreInt32(&d.restricted, 1)
}

func (d *ContextMessageDelegate) unrestrict() {
	atomic.StoreInt32(&d.restricted, 0)
}

func (d *ContextMessageDelegate) registerMessageObservers(observers []MessageObserver) {
	d.messageObserver = CompositeMessageObserver(observers)
}

func (d *ContextMessageDelegate) unregisterAllMessageObservers() {
	d.messageObserver = nil
}
