package utils

import (
	"context"
	
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/tracing"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// TraceJob starts or continues a trace for a job operation
// This is a wrapper around the tracing package for backward compatibility
func TraceJob(ctx context.Context, operationName string, runID string) (context.Context, tracer.Span) {
	return tracing.TraceJob(ctx, operationName, runID)
}

// TagJobRun adds standardized job metadata to a span
// This function correctly passes the Run struct fields to tracing.TagRunInfo
func TagJobRun(span tracer.Span, run state.Run) {
	tracing.TagRunInfo(span,
		run.RunID, run.DefinitionID, run.Alias, run.Status, run.ClusterName,
		run.QueuedAt, run.StartedAt, run.FinishedAt,
		run.PodName, run.Namespace, run.ExitReason, run.ExitCode)
}
