package tracing

import (
	"github.com/Bofry/trace"
	"github.com/Bofry/worker-postgres/internal"
)

func GetTracer(v interface{}) *trace.SeverityTracer {
	return internal.GlobalTracerManager.GenerateManagedTracer(v)
}
