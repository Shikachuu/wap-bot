package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// Name used to define the application name that is being instrumented.
const name = "github.com/Shikachuu/wap-bot"

var (
	// Tracer contains the global tracer implementation that uses the correct package name and params from env vars.
	Tracer = otel.Tracer(name)
	// Meter contains the global meter implementation that uses the correct package name and params from env vars.
	Meter = otel.Meter(name)
)

// SetupOTel creates a new open telemetry trace and metric provider and sets them on the global context.
//
// ctx is the current context that we use to set these metrics up.
//
// Returns a shutdown function and error if any.
func SetupOTel(ctx context.Context) (func(context.Context) error, error) {
	res := resource.Default()

	se, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("span exporter creation: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(se),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	mr, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("metric reader creation: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(mr),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	return func(sCtx context.Context) error {
		if sErr := tp.Shutdown(sCtx); sErr != nil {
			return fmt.Errorf("trace provider shutdown: %w", sErr)
		}

		if sErr := mp.Shutdown(sCtx); sErr != nil {
			return fmt.Errorf("metric provider shutdown: %w", sErr)
		}

		return nil
	}, nil
}
