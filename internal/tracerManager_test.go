package internal

import (
	"context"
	"strings"
	"testing"

	"github.com/Bofry/trace"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestManagedTracerAttributeLengthLimit(t *testing.T) {
	const limit = 5
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"short", "hello", "hello"},
		{"ascii", "hello world", "hello"},
		{"multibyte", "中文測試資料", "中文測試資"},
		{"replacement runes", strings.Repeat("\uFFFD", 32), strings.Repeat("\uFFFD", limit)},
		{"mixed replacement runes", "ab\uFFFDcdef", "ab\uFFFDcd"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := tracetest.NewSpanRecorder()
			limits := sdktrace.NewSpanLimits()
			limits.AttributeValueLengthLimit = limit
			provider := sdktrace.NewTracerProvider(
				sdktrace.WithSpanProcessor(recorder),
				sdktrace.WithSpanLimits(limits),
			)
			t.Cleanup(func() {
				if err := provider.Shutdown(context.Background()); err != nil {
					t.Error(err)
				}
			})
			manager := NewTraceManager()
			manager.TracerProvider = trace.CreateSeverityTracerProvider(provider)
			tracer := manager.GenerateManagedTracer(manager)
			span := tracer.Start(context.Background(), "attribute-limit")
			span.Tags(attribute.String("value", tc.input), attribute.StringSlice("values", []string{tc.input, "ok"}))
			if !span.HasTraceID() {
				t.Fatal("managed tracer did not create a recording span")
			}
			span.End()

			spans := recorder.Ended()
			if len(spans) != 1 {
				t.Fatalf("got %d spans, want 1", len(spans))
			}
			attrs := attribute.NewSet(spans[0].Attributes()...)
			value, ok := attrs.Value("value")
			if !ok || value.AsString() != tc.want {
				t.Errorf("string attribute = %q, want %q", value.AsString(), tc.want)
			}
			values, ok := attrs.Value("values")
			got := values.AsStringSlice()
			if !ok || len(got) != 2 || got[0] != tc.want || got[1] != "ok" {
				t.Errorf("slice attribute = %q, want [%q %q]", got, tc.want, "ok")
			}
		})
	}
}
