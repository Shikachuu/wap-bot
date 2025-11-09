// Package telemetry provides observability infrastructure including OpenTelemetry
// tracing, metrics, and structured logging for the wap-bot application.
//
// Use SetupOTel to initialize tracing and metrics, and SetupLogger to configure
// structured logging. The package exports global Tracer and Meter instances for
// instrumentation throughout the application.
//
//	shutdown, err := telemetry.SetupOTel(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer shutdown(context.Background())
//
// OpenTelemetry exporters are configured via standard OTEL_* environment variables.
package telemetry