package internal

import (
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

type metricDataAccumulator struct {
	m    []pmetric.Metrics
	time time.Time
	mbs  *metadata.MetricsBuilders
}

func (acc *metricDataAccumulator) resourceStats(r metadata.ResourceStatusData) {
	currentTime := pcommon.NewTimestampFromTime(acc.time)

	addModuleStats(acc.mbs.KymaTelemetryModuleMetricsBuilder, metadata.KymaModuleMetrics.ModuleState, r, currentTime)
	rb := acc.mbs.KymaTelemetryModuleMetricsBuilder.NewResourceBuilder()
	rb.SetK8sNamespaceName(r.Namespace)
	acc.m = append(acc.m, acc.mbs.KymaTelemetryModuleMetricsBuilder.Emit(metadata.WithResource(rb.Emit())))
}

func (acc *metricDataAccumulator) resourceConditionStats(name string, namespace string, r v1.Condition) {
	currentTime := pcommon.NewTimestampFromTime(acc.time)

	addModuleConditionStats(acc.mbs.KymaTelemetryModuleMetricsBuilder, metadata.KymaModuleMetrics.ModuleCondition, name, r, currentTime)
	rb := acc.mbs.KymaTelemetryModuleMetricsBuilder.NewResourceBuilder()
	rb.SetK8sNamespaceName(namespace)
	acc.m = append(acc.m, acc.mbs.KymaTelemetryModuleMetricsBuilder.Emit(metadata.WithResource(rb.Emit())))
}

func addModuleStats(mb *metadata.MetricsBuilder, moduleMetrics metadata.RecordModuleStateDatapointFunc, r metadata.ResourceStatusData, currentTime pcommon.Timestamp) {
	value := 0
	if r.State == "Ready" {
		value = 1
	}
	moduleMetrics(mb, currentTime, int64(value), r.State, r.Name)
}

func addModuleConditionStats(mb *metadata.MetricsBuilder, moduleMetrics metadata.RecordModuleConditionDatapointFunc, name string, r v1.Condition, currentTime pcommon.Timestamp) {
	value := 0
	if r.Status == "True" {
		value = 1
	}
	moduleMetrics(mb, currentTime, int64(value), name, r.Reason, string(r.Status), r.Type)
}
