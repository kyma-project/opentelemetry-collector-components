package istionoisefilter

import (
	"context"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"

	"github.com/kyma-project/opentelemetry-collector-components/processor/istionoisefilter/internal/rules"
)

type istioNoiseFilter struct {
	cfg    *Config
	logger *zap.Logger
}

func newProcessor(cfg *Config, set processor.Settings) *istioNoiseFilter {
	return &istioNoiseFilter{
		cfg:    cfg,
		logger: set.Logger,
	}
}

func (f *istioNoiseFilter) processTraces(_ context.Context, traces ptrace.Traces) (ptrace.Traces, error) {
	for _, rs := range traces.ResourceSpans().All() {
		for _, ss := range rs.ScopeSpans().All() {
			ss.Spans().RemoveIf(func(span ptrace.Span) bool {
				return rules.ShouldDropSpan(span, rs.Resource().Attributes())
			})
		}
	}

	return traces, nil
}

func (f *istioNoiseFilter) processLogs(_ context.Context, logs plog.Logs) (plog.Logs, error) {
	for _, rl := range logs.ResourceLogs().All() {
		for _, sl := range rl.ScopeLogs().All() {
			sl.LogRecords().RemoveIf(func(logRecord plog.LogRecord) bool {
				return rules.ShouldDropLogRecord(logRecord, rl.Resource().Attributes())
			})
		}
	}

	return logs, nil
}

func (f *istioNoiseFilter) processMetrics(_ context.Context, metrics pmetric.Metrics) (pmetric.Metrics, error) {
	for _, rm := range metrics.ResourceMetrics().All() {
		for _, sm := range rm.ScopeMetrics().All() {
			for _, m := range sm.Metrics().All() {
				switch m.Type() {
				case pmetric.MetricTypeGauge:
					m.Gauge().DataPoints().RemoveIf(func(ndp pmetric.NumberDataPoint) bool {
						return rules.ShouldDropMetricDataPoint(m.Name(), ndp.Attributes())
					})
				case pmetric.MetricTypeSum:
					m.Sum().DataPoints().RemoveIf(func(ndp pmetric.NumberDataPoint) bool {
						return rules.ShouldDropMetricDataPoint(m.Name(), ndp.Attributes())
					})
				case pmetric.MetricTypeHistogram:
					m.Histogram().DataPoints().RemoveIf(func(hdp pmetric.HistogramDataPoint) bool {
						return rules.ShouldDropMetricDataPoint(m.Name(), hdp.Attributes())
					})
				case pmetric.MetricTypeExponentialHistogram:
					m.ExponentialHistogram().DataPoints().RemoveIf(func(ehdp pmetric.ExponentialHistogramDataPoint) bool {
						return rules.ShouldDropMetricDataPoint(m.Name(), ehdp.Attributes())
					})
				case pmetric.MetricTypeSummary:
					m.Summary().DataPoints().RemoveIf(func(sdp pmetric.SummaryDataPoint) bool {
						return rules.ShouldDropMetricDataPoint(m.Name(), sdp.Attributes())
					})
				default:
					f.logger.Warn("Unknown metric type encountered in processMetrics",
						zap.String("metric_name", m.Name()),
						zap.Any("metric_type", m.Type()),
					)
				}
			}

		}
	}

	return metrics, nil
}
