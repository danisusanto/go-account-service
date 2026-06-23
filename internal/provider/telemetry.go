package provider

import (
	"context"

	"github.com/dani-susanto/go-common/env"
	"github.com/dani-susanto/go-common/telemetry"
	"go.opentelemetry.io/otel/trace"
)

func SetupTelemetry(ctx context.Context, env env.Env) (trace.Tracer, error) {
	telemetry := telemetry.New(
		ctx,
		env.Get("TELEMETRY_COLLECTOR_HOST"),
		env.Get("TELEMETRY_COLLECTOR_PORT"),
		env.Get("APP_NAME"),
	)

	tracer, err := telemetry.NewTracer()
	if err != nil {
		return nil, err
	}

	return tracer, nil
}
