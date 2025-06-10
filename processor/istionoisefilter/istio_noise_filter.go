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

			for k := range scopeMetrics.Metrics().Len() {
				metric := scopeMetrics.Metrics().At(k)
				metricName := metric.Name()

				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					metric.Gauge().DataPoints().RemoveIf(func(ndp pmetric.NumberDataPoint) bool {
						return rules.ShouldDropMetricDataPoint(metricName, ndp.Attributes())
					})
				case pmetric.MetricTypeSum:
					metric.Sum().DataPoints().RemoveIf(func(ndp pmetric.NumberDataPoint) bool {
						return rules.ShouldDropMetricDataPoint(metricName, ndp.Attributes())
					})
				case pmetric.MetricTypeHistogram:
					metric.Histogram().DataPoints().RemoveIf(func(hdp pmetric.HistogramDataPoint) bool {
						return rules.ShouldDropMetricDataPoint(metricName, hdp.Attributes())
					})
				case pmetric.MetricTypeExponentialHistogram:
					metric.ExponentialHistogram().DataPoints().RemoveIf(func(ehdp pmetric.ExponentialHistogramDataPoint) bool {
						return rules.ShouldDropMetricDataPoint(metricName, ehdp.Attributes())
					})
				case pmetric.MetricTypeSummary:
					metric.Summary().DataPoints().RemoveIf(func(sdp pmetric.SummaryDataPoint) bool {
						return rules.ShouldDropMetricDataPoint(metricName, sdp.Attributes())
					})
				default:
					f.logger.Warn("Unknown metric type encountered in processMetrics",
						zap.String("metric_name", metricName),
						zap.Any("metric_type", metric.Type()),
					)
				}
			}

		}
	}

	return metrics, nil
}
