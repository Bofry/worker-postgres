package middleware

import (
	"reflect"
	"unsafe"

	"github.com/Bofry/worker-postgres/internal"
)

func isMessageHandler(rv reflect.Value) bool {
	return internal.IsMessageHandler(rv)
}

func isMessageObserver(rv reflect.Value) bool {
	if rv.IsValid() {
		return rv.Type().AssignableTo(typeOfMessageObserver)
	}
	return false
}

func asMessageHandler(rv reflect.Value) internal.MessageHandler {
	return internal.AsMessageHandler(rv)
}

func asMessageObserver(rv reflect.Value) internal.MessageObserver {
	if rv.IsValid() {
		if v, ok := rv.Convert(typeOfMessageObserver).Interface().(internal.MessageObserver); ok {
			return v
		}
	}
	return nil
}

func asPostgresWorker(rv reflect.Value) *internal.PostgresWorker {
	return reflect.NewAt(typeOfHost, unsafe.Pointer(rv.Pointer())).
		Interface().(*internal.PostgresWorker)
}
