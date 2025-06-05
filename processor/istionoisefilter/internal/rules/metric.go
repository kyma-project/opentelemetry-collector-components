package rules

import (
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func ShouldDropMetric(metric pmetric.Metric) bool {
	if !strings.HasPrefix(metric.Name(), "istio.") {
		return false
	}

	dataPointAttrs := extractDataPointAttrs(metric)

	if sourceWorkload, found := dataPointAttrs.Get("source_workload"); found && sourceWorkload.Str() == "telemetry-metric-agent" {
		return true
	}

	if destinationWorkload, found := dataPointAttrs.Get("destination_workload"); found {
		switch destinationWorkload.Str() {
		case "telemetry-log-gateway", "telemetry-metric-gateway", "telemetry-trace-gateway":
			return true
		}
	}

	return false
}

func extractDataPointAttrs(metric pmetric.Metric) pcommon.Map {
	attrs := pcommon.NewMap()

	//exhaustive:enforce
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		for i := range metric.Gauge().DataPoints().Len() {
			dp := metric.Gauge().DataPoints().At(i)
			dp.Attributes().CopyTo(attrs)
		}
	case pmetric.MetricTypeSum:
		for i := range metric.Sum().DataPoints().Len() {
			dp := metric.Sum().DataPoints().At(i)
			dp.Attributes().CopyTo(attrs)
		}
	case pmetric.MetricTypeHistogram:
		for i := range metric.Histogram().DataPoints().Len() {
			dp := metric.Histogram().DataPoints().At(i)
			dp.Attributes().CopyTo(attrs)
		}
	case pmetric.MetricTypeExponentialHistogram:
		for i := range metric.ExponentialHistogram().DataPoints().Len() {
			dp := metric.ExponentialHistogram().DataPoints().At(i)
			dp.Attributes().CopyTo(attrs)
		}
	case pmetric.MetricTypeSummary:
		for i := range metric.Summary().DataPoints().Len() {
			dp := metric.Summary().DataPoints().At(i)
			dp.Attributes().CopyTo(attrs)
		}
	default:
		return attrs
	}

	return attrs
}
