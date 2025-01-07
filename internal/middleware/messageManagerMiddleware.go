package middleware

import (
	"github.com/Bofry/host"
	"github.com/Bofry/structproto"
	. "github.com/Bofry/worker-postgres/internal"
)

var _ host.Middleware = new(MessageManagerMiddleware)

type MessageManagerMiddleware struct {
	MessageManager interface{}
}

func (m *MessageManagerMiddleware) Init(app *host.AppModule) {
	var (
		worker    = asPostgresWorker(app.Host())
		registrar = NewPostgresWorkerRegistrar(worker)
	)

	// register MessageManager offer NsqWorker processing later.
	registrar.SetMessageManager(m.MessageManager)

	// binding MessageManager
	binder := &MessageManagerBinder{
		registrar: registrar,
		app:       app,
	}

	err := m.bindMessageManager(m.MessageManager, binder)
	if err != nil {
		panic(err)
	}
}

func (m *MessageManagerMiddleware) bindMessageManager(target interface{}, binder *MessageManagerBinder) error {
	prototype, err := structproto.Prototypify(target,
		&structproto.StructProtoResolveOption{
			TagName:             TAG_SLOT,
			TagResolver:         SlotTagResolver,
			CheckDuplicateNames: true,
		},
	)
	if err != nil {
		return err
	}

	return prototype.Bind(binder)
}
