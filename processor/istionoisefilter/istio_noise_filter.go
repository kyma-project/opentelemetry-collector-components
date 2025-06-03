package istionoisefilter

import (
	"context"

	"github.com/kyma-project/opentelemetry-collector-components/processor/istionoisefilter/internal/filter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type istioNoiseFilter struct {
	cfg *Config
}

func newProcessor(cfg *Config) *istioNoiseFilter {
	return &istioNoiseFilter{
		cfg: cfg,
	}
}

func (f *istioNoiseFilter) processTraces(_ context.Context, traces ptrace.Traces) (ptrace.Traces, error) {
	for i := range traces.ResourceSpans().Len() {
		resourceSpans := traces.ResourceSpans().At(i)

		for j := range resourceSpans.ScopeSpans().Len() {
			scopeSpans := resourceSpans.ScopeSpans().At(j)

			spans := scopeSpans.Spans()
			spans.RemoveIf(func(span ptrace.Span) bool {
				return filter.ShouldDropSpan(span, resourceSpans.Resource().Attributes())
			})
		}
	}

	return traces, nil
}

func (f *istioNoiseFilter) processLogs(_ context.Context, logs plog.Logs) (plog.Logs, error) {
	for i := range logs.ResourceLogs().Len() {
		resourceLogs := logs.ResourceLogs().At(i)

		for j := range resourceLogs.ScopeLogs().Len() {
			scopeLogs := resourceLogs.ScopeLogs().At(j)

			logRecords := scopeLogs.LogRecords()
			logRecords.RemoveIf(func(logRecord plog.LogRecord) bool {
				return filter.ShouldDropLogRecord(logRecord, resourceLogs.Resource().Attributes())
			})
		}
	}

	return logs, nil
}

func (f *istioNoiseFilter) processMetrics(_ context.Context, metrics pmetric.Metrics) (pmetric.Metrics, error) {
	return pmetric.Metrics{}, nil
}
