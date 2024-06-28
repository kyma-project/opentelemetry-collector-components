package metadata

import "go.opentelemetry.io/collector/pdata/pcommon"

type RecordModuleStateDatapointFunc func(*MetricsBuilder, pcommon.Timestamp, int64, string, string)
type RecordModuleConditionDatapointFunc func(*MetricsBuilder, pcommon.Timestamp, int64, string, string, string, string)

type MetricsBuilders struct {
	KymaTelemetryModuleMetricsBuilder *MetricsBuilder
}

type ModuleMetrics struct {
	ModuleState     RecordModuleStateDatapointFunc
	ModuleCondition RecordModuleConditionDatapointFunc
}

var KymaModuleMetrics = ModuleMetrics{
	ModuleState:     (*MetricsBuilder).RecordKymaModuleStatusStatDataPoint,
	ModuleCondition: (*MetricsBuilder).RecordKymaModuleStatusConditionDataPoint,
}
