package utils

import (
	"context"

	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/tracing"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// TraceJob starts or continues a trace for a job operation
func TraceJob(ctx context.Context, operationName string, runID string) (context.Context, tracer.Span) {
	return tracing.TraceJob(ctx, operationName, runID)
}

// TagJobRun adds standardized job metadata to a span
func TagJobRun(span tracer.Span, run state.Run) {
	tracing.TagRunInfo(span,
		run.RunID, run.DefinitionID, run.Alias, run.Status, run.ClusterName,
		run.QueuedAt, run.StartedAt, run.FinishedAt,
		run.PodName, run.Namespace, run.ExitReason, run.ExitCode, string(run.Tier))
}
