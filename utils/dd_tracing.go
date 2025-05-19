package utils

import (
	"context"
	"time"

	"github.com/stitchfix/flotilla-os/state"
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

// TagJobRun adds standardized job metadata to a span
func TagJobRun(span tracer.Span, run state.Run) {
	if span == nil {
		return
	}

	span.SetTag("job.run_id", run.RunID)
	span.SetTag("job.definition_id", run.DefinitionID)
	span.SetTag("job.alias", run.Alias)
	span.SetTag("job.status", run.Status)
	span.SetTag("job.cluster", run.ClusterName)

	if run.QueuedAt != nil {
		span.SetTag("job.queued_at", run.QueuedAt.Unix())
		queuedDuration := time.Since(*run.QueuedAt)
		span.SetTag("job.queue_duration_sec", queuedDuration.Seconds())

		if queuedDuration > 5*time.Minute {
			span.SetTag("job.long_queued", true)
		}
	}

	if run.StartedAt != nil {
		span.SetTag("job.started_at", run.StartedAt.Unix())
		if run.FinishedAt != nil {
			span.SetTag("job.duration_sec", run.FinishedAt.Sub(*run.StartedAt).Seconds())
		} else if run.Status == state.StatusRunning {
			span.SetTag("job.running_duration_sec", time.Since(*run.StartedAt).Seconds())
		}
	}

	if run.PodName != nil {
		span.SetTag("job.pod_name", *run.PodName)
	}

	if run.Namespace != nil {
		span.SetTag("job.namespace", *run.Namespace)
	}

	if run.ExitCode != nil && *run.ExitCode != 0 {
		span.SetTag("error", true)
		span.SetTag("job.exit_code", *run.ExitCode)
	}

	if run.ExitReason != nil {
		span.SetTag("job.exit_reason", *run.ExitReason)
	}
}
