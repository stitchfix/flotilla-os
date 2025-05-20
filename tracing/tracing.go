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
	exitCode *int64) {

	if span == nil {
		return
	}

	span.SetTag("job.run_id", runID)
	span.SetTag("job.definition_id", definitionID)
	span.SetTag("job.alias", alias)
	span.SetTag("job.status", status)
	span.SetTag("job.cluster", clusterName)

	if queuedAt != nil {
		span.SetTag("job.queued_at", queuedAt.Unix())
		queuedDuration := time.Since(*queuedAt)
		span.SetTag("job.queue_duration_sec", queuedDuration.Seconds())

		if queuedDuration > 5*time.Minute {
			span.SetTag("job.long_queued", true)
		}
	}

	if startedAt != nil {
		span.SetTag("job.started_at", startedAt.Unix())
		if finishedAt != nil {
			span.SetTag("job.duration_sec", finishedAt.Sub(*startedAt).Seconds())
		} else if status == "RUNNING" {
			span.SetTag("job.running_duration_sec", time.Since(*startedAt).Seconds())
		}
	}

	if podName != nil {
		span.SetTag("job.pod_name", *podName)
	}

	if namespace != nil {
		span.SetTag("job.namespace", *namespace)
	}

	if exitCode != nil && *exitCode != 0 {
		span.SetTag("error", true)
		span.SetTag("job.exit_code", *exitCode)
	}

	if exitReason != nil {
		span.SetTag("job.exit_reason", *exitReason)
	}
}
