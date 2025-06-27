package istionoisefilter



import (
	"context"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
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

//nolint:dupl // trace and log processing has similar shape, but different logic
func (f *istioNoiseFilter) processTraces(_ context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	td.ResourceSpans().RemoveIf(func(rs ptrace.ResourceSpans) bool {
		rs.ScopeSpans().RemoveIf(func(ss ptrace.ScopeSpans) bool {
			ss.Spans().RemoveIf(func(span ptrace.Span) bool {
				return rules.ShouldDropSpan(span, rs.Resource().Attributes())
			})

			return ss.Spans().Len() == 0
		})

		return rs.ScopeSpans().Len() == 0
	})

	if td.ResourceSpans().Len() == 0 {
		return td, processorhelper.ErrSkipProcessingData
	}

	return td, nil
}

//nolint:dupl // trace and log processing has similar shape, but different logic
func (f *istioNoiseFilter) processLogs(_ context.Context, ld plog.Logs) (plog.Logs, error) {
	ld.ResourceLogs().RemoveIf(func(rl plog.ResourceLogs) bool {
		rl.ScopeLogs().RemoveIf(func(sl plog.ScopeLogs) bool {
			sl.LogRecords().RemoveIf(func(logRecord plog.LogRecord) bool {
				return rules.ShouldDropLogRecord(logRecord, rl.Resource().Attributes())
			})

			return sl.LogRecords().Len() == 0
		})

		return rl.ScopeLogs().Len() == 0
	})

	if ld.ResourceLogs().Len() == 0 {
		return ld, processorhelper.ErrSkipProcessingData
	}

	return ld, nil
}

func (f *istioNoiseFilter) processMetrics(_ context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	md.ResourceMetrics().RemoveIf(func(rm pmetric.ResourceMetrics) bool {
		rm.ScopeMetrics().RemoveIf(func(sm pmetric.ScopeMetrics) bool {
			sm.Metrics().RemoveIf(func(m pmetric.Metric) bool {
				dataPointsLen := f.removeMetricDataPointsIfMatch(m)
				return dataPointsLen == 0
			})

			return sm.Metrics().Len() == 0
		})

		return rm.ScopeMetrics().Len() == 0
	})

	if md.ResourceMetrics().Len() == 0 {
		return md, processorhelper.ErrSkipProcessingData
	}

	return md, nil
}

func (f *istioNoiseFilter) removeMetricDataPointsIfMatch(m pmetric.Metric) int {
	switch m.Type() {
	case pmetric.MetricTypeGauge:
		m.Gauge().DataPoints().RemoveIf(func(ndp pmetric.NumberDataPoint) bool {
			return rules.ShouldDropMetricDataPoint(m.Name(), ndp.Attributes())
		})
		return m.Gauge().DataPoints().Len()
	case pmetric.MetricTypeSum:
		m.Sum().DataPoints().RemoveIf(func(ndp pmetric.NumberDataPoint) bool {
			return rules.ShouldDropMetricDataPoint(m.Name(), ndp.Attributes())
		})
		return m.Sum().DataPoints().Len()
	case pmetric.MetricTypeHistogram:
		m.Histogram().DataPoints().RemoveIf(func(hdp pmetric.HistogramDataPoint) bool {
			return rules.ShouldDropMetricDataPoint(m.Name(), hdp.Attributes())
		})
		return m.Histogram().DataPoints().Len()
	case pmetric.MetricTypeExponentialHistogram:
		m.ExponentialHistogram().DataPoints().RemoveIf(func(ehdp pmetric.ExponentialHistogramDataPoint) bool {
			return rules.ShouldDropMetricDataPoint(m.Name(), ehdp.Attributes())
		})
		return m.ExponentialHistogram().DataPoints().Len()
	case pmetric.MetricTypeSummary:
		m.Summary().DataPoints().RemoveIf(func(sdp pmetric.SummaryDataPoint) bool {
			return rules.ShouldDropMetricDataPoint(m.Name(), sdp.Attributes())
		})
		return m.Summary().DataPoints().Len()
	default:
		f.logger.Warn("Unknown metric type encountered in processMetrics",
			zap.String("metric_name", m.Name()),
			zap.Any("metric_type", m.Type()),
		)
		return -1
	}
}
