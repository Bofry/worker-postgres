package middleware

import (
	"fmt"
	"os"
	"reflect"

	"github.com/Bofry/host"
	"github.com/Bofry/structproto"
	"github.com/Bofry/structproto/reflecting"
	"github.com/Bofry/structproto/tagresolver"
	"github.com/Bofry/worker-postgres/internal"
)

var _ structproto.StructBinder = new(MessageManagerBinder)

type MessageManagerBinder struct {
	registrar *internal.PostgresWorkerRegistrar
	app       *host.AppModule
}

func (b *MessageManagerBinder) Init(context *structproto.StructProtoContext) error {
	return nil
}

func (b *MessageManagerBinder) Bind(field structproto.FieldInfo, rv reflect.Value) error {
	if !rv.IsValid() {
		return fmt.Errorf("specifiec argument 'rv' is invalid")
	}

	// assign zero if rv is nil
	rvMessageHandler := reflecting.AssignZero(rv)
	binder := &MessageHandlerBinder{
		messageHandlerType: rv.Type().Name(),
		components: map[string]reflect.Value{
			host.APP_CONFIG_FIELD:           b.app.Config(),
			host.APP_SERVICE_PROVIDER_FIELD: b.app.ServiceProvider(),
		},
	}
	err := b.bindMessageHandler(rvMessageHandler, binder)
	if err != nil {
		return err
	}

	// register MessageHandlers
	var (
		moduleID = field.IDName()
		slot     = field.Name()
		offset   = field.Tag().Get(TAG_OFFSET)
	)

	if !b.isKnownStream(slot) {
		optExpandEnv := field.Tag().Get(TAG_OPT_EXPAND_ENV)
		if optExpandEnv != OPT_OFF || len(optExpandEnv) == 0 || optExpandEnv == OPT_ON {
			slot = os.ExpandEnv(slot)
		}
	}

	return b.registerRoute(moduleID, slot, offset, rvMessageHandler)
}

func (b *MessageManagerBinder) Deinit(context *structproto.StructProtoContext) error {
	return nil
}

func (b *MessageManagerBinder) bindMessageHandler(target reflect.Value, binder *MessageHandlerBinder) error {
	prototype, err := structproto.Prototypify(target,
		&structproto.StructProtoResolveOption{
			TagResolver: tagresolver.NoneTagResolver,
		})
	if err != nil {
		return err
	}

	return prototype.Bind(binder)
}

func (b *MessageManagerBinder) registerRoute(moduleID, slot, offset string, rv reflect.Value) error {
	// register MessageHandlers
	if isMessageHandler(rv) {
		handler := asMessageHandler(rv)
		if handler != nil {
			if slot == INVALID_MESSAGE_HANDLER_SLOT_SYMBOL {
				b.registrar.SetInvalidMessageHandler(handler)
			} else {
				if offset != "-" {
					// FIXME:....
					b.registrar.RegisterSlot(internal.SlotOffset{
						Slot: slot,
						LSN:  offset,
					})
				}
				b.registrar.AddRouter(slot, handler, moduleID)
			}
		}
	}
	return nil
}

func (b *MessageManagerBinder) isKnownStream(stream string) bool {
	switch stream {
	case INVALID_MESSAGE_HANDLER_SLOT_SYMBOL:
		return true
	}
	return false
}
