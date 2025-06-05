package istionoisefilter

import (
	"context"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/kyma-project/opentelemetry-collector-components/processor/istionoisefilter/internal/rules"
)

type istioNoiseFilter struct {
	cfg *Config
}

func newProcessor(cfg *Config) *istioNoiseFilter {
	return &istioNoiseFilter{
		cfg: cfg,
	}
}

//nolint:dupl //all 3 process methods are similar, but operate on different data types
func (f *istioNoiseFilter) processTraces(_ context.Context, traces ptrace.Traces) (ptrace.Traces, error) {
	for i := range traces.ResourceSpans().Len() {
		resourceSpans := traces.ResourceSpans().At(i)

		for j := range resourceSpans.ScopeSpans().Len() {
			scopeSpans := resourceSpans.ScopeSpans().At(j)

			scopeSpans.Spans().RemoveIf(func(span ptrace.Span) bool {
				return rules.ShouldDropSpan(span, resourceSpans.Resource().Attributes())
			})
		}
	}

	return traces, nil
}

//nolint:dupl //all 3 process methods are similar, but operate on different data types
func (f *istioNoiseFilter) processLogs(_ context.Context, logs plog.Logs) (plog.Logs, error) {
	for i := range logs.ResourceLogs().Len() {
		resourceLogs := logs.ResourceLogs().At(i)

		for j := range resourceLogs.ScopeLogs().Len() {
			scopeLogs := resourceLogs.ScopeLogs().At(j)

			scopeLogs.LogRecords().RemoveIf(func(logRecord plog.LogRecord) bool {
				return rules.ShouldDropLogRecord(logRecord, resourceLogs.Resource().Attributes())
			})
		}
	}

	return logs, nil
}

//nolint:dupl //all 3 process methods are similar, but operate on different data types
func (f *istioNoiseFilter) processMetrics(_ context.Context, metrics pmetric.Metrics) (pmetric.Metrics, error) {
	for i := range metrics.ResourceMetrics().Len() {
		resourceMetrics := metrics.ResourceMetrics().At(i)

		for j := range resourceMetrics.ScopeMetrics().Len() {
			scopeMetrics := resourceMetrics.ScopeMetrics().At(j)

			scopeMetrics.Metrics().RemoveIf(func(m pmetric.Metric) bool {
				return rules.ShouldDropMetric(m)
			})
		}
	}

	return metrics, nil
}
