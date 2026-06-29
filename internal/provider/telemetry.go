package provider

import (
	"context"
	"strconv"

	"github.com/dani-susanto/go-common/env"
	"github.com/dani-susanto/go-common/telemetry"
	"go.opentelemetry.io/otel/trace"
)

func SetupTelemetry(ctx context.Context, env env.Env) (trace.Tracer, error) {
	enabled, err := strconv.ParseBool(env.Get("TELEMETRY_COLLECTOR_ENABLED"))
	if err != nil {
		return nil, err
	}

	telemetry := telemetry.New(
		ctx,
		env.Get("TELEMETRY_COLLECTOR_HOST"),
		env.Get("TELEMETRY_COLLECTOR_PORT"),
		env.Get("APP_NAME"),
		enabled,
	)

	tracer, err := telemetry.NewTracer()
	if err != nil {
		return nil, err
	}

	return tracer, nil
}
