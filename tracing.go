package ottwirp

import (
	"context"

	ot "github.com/opentracing/opentracing-go"
	"github.com/twitchtv/twirp"
)

const (
	RequestReceivedEvent = "request.received"
)

// TODO: Add functional options for handling things such as logging status
// codes, etc.

// NewOpenTracingServerHook provides a twirp.ServerHooks struct which records
// OpenTracing spans.
func NewOpenTracingServerHook(tracer ot.Tracer) *twirp.ServerHooks {
	// TODO: Determine if setting this global tracer here is a good idea or should
	// be left up to the user.
	ot.SetGlobalTracer(tracer)

	hooks := &twirp.ServerHooks{}

	// RequestReceived: Create the initial span that we will use for the duration
	// of the request.
	hooks.RequestReceived = func(ctx context.Context) (context.Context, error) {
		// Create the initial span, it won't have a method name just yet.
		span, ctx := ot.StartSpanFromContext(ctx, RequestReceivedEvent)
		if span != nil {
			span.SetTag("component", "twirp")

			packageName, havePackageName := twirp.PackageName(ctx)
			if havePackageName {
				span.SetTag("package", packageName)
			}

			serviceName, haveServiceName := twirp.ServiceName(ctx)
			if haveServiceName {
				span.SetTag("service", serviceName)
			}

			span.SetTag("span.kind", "server")
		}

		return ctx, nil
	}

	// RequestRouted: Set the operation name based on the MethodName extracted
	// from span.
	hooks.RequestRouted = func(ctx context.Context) (context.Context, error) {
		span := ot.SpanFromContext(ctx)
		if span != nil {
			method, ok := twirp.MethodName(ctx)
			if !ok {
				return ctx, nil
			}

			span.SetOperationName(method)
		}

		return ctx, nil
	}

	// ResponseSent: Set the status code and mark the span as finished.
	hooks.ResponseSent = func(ctx context.Context) {
		span := ot.SpanFromContext(ctx)

		status, haveStatus := twirp.StatusCode(ctx)

		if span != nil {
			if haveStatus {
				// TODO: Check the status code, if it's a non-2xx/3xx status code, we
				// should probably mark it as an error of sorts.
				span.SetTag("http.status_code", status)
			}

			span.Finish()
		}
	}

	hooks.Error = func(ctx context.Context, err twirp.Error) context.Context {
		span := ot.SpanFromContext(ctx)
		if span != nil {
			span.SetTag("error", true)
		}
		// TODO: Set an error simply on the span.
		return nil
	}

	return hooks
}