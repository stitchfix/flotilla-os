package tracing

import (
	"context"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// TraceJob starts or continues a trace for a job operation
func TraceJob(ctx context.Context, operationName string, runID string) (context.Context, tracer.Span) {
	span, ctx := tracer.StartSpanFromContext(
		ctx,
		operationName,
		tracer.ResourceName(runID),
		tracer.Tag("job.run_id", runID),
	)

	return ctx, span
}

// TagRunInfo adds standardized job metadata to a span
func TagRunInfo(span tracer.Span,
	runID, definitionID, alias, status, clusterName string,
	queuedAt, startedAt, finishedAt *time.Time,
	podName, namespace, exitReason *string,
	exitCode *int64, tier string) {

	if span == nil {
		return
	}

	span.SetTag("job.run_id", runID)

	if exitReason != nil {
		span.SetTag("job.exit_reason", *exitReason)
	}
}

type TextMapCarrier map[string]string

// ForeachKey implements the TextMapReader interface for Extract
func (c TextMapCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, v := range c {
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Set implements the TextMapWriter interface for Inject
func (c TextMapCarrier) Set(key, val string) {
	c[key] = val
}
