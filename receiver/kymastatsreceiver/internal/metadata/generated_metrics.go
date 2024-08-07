// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/filter"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
)

type metricKymaModuleStatusConditions struct {
	data     pmetric.Metric // data buffer for generated metric.
	config   MetricConfig   // metric config provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills kyma.module.status.conditions metric with initial data.
func (m *metricKymaModuleStatusConditions) init() {
	m.data.SetName("kyma.module.status.conditions")
	m.data.SetDescription("The module status conditions. Possible metric values for condition status are 'True' => 1, 'False' => 0, and -1 for other status values.")
	m.data.SetUnit("1")
	m.data.SetEmptyGauge()
	m.data.Gauge().DataPoints().EnsureCapacity(m.capacity)
}

func (m *metricKymaModuleStatusConditions) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val int64, reasonAttributeValue string, statusAttributeValue string, typeAttributeValue string) {
	if !m.config.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntValue(val)
	dp.Attributes().PutStr("reason", reasonAttributeValue)
	dp.Attributes().PutStr("status", statusAttributeValue)
	dp.Attributes().PutStr("type", typeAttributeValue)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricKymaModuleStatusConditions) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricKymaModuleStatusConditions) emit(metrics pmetric.MetricSlice) {
	if m.config.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricKymaModuleStatusConditions(cfg MetricConfig) metricKymaModuleStatusConditions {
	m := metricKymaModuleStatusConditions{config: cfg}
	if cfg.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

type metricKymaModuleStatusState struct {
	data     pmetric.Metric // data buffer for generated metric.
	config   MetricConfig   // metric config provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills kyma.module.status.state metric with initial data.
func (m *metricKymaModuleStatusState) init() {
	m.data.SetName("kyma.module.status.state")
	m.data.SetDescription("The module status state, metric value is 1 for last scraped module status state, including state as metric attribute.")
	m.data.SetUnit("1")
	m.data.SetEmptyGauge()
	m.data.Gauge().DataPoints().EnsureCapacity(m.capacity)
}

func (m *metricKymaModuleStatusState) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val int64, stateAttributeValue string) {
	if !m.config.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntValue(val)
	dp.Attributes().PutStr("state", stateAttributeValue)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricKymaModuleStatusState) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricKymaModuleStatusState) emit(metrics pmetric.MetricSlice) {
	if m.config.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricKymaModuleStatusState(cfg MetricConfig) metricKymaModuleStatusState {
	m := metricKymaModuleStatusState{config: cfg}
	if cfg.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

// MetricsBuilder provides an interface for scrapers to report metrics while taking care of all the transformations
// required to produce metric representation defined in metadata and user config.
type MetricsBuilder struct {
	config                           MetricsBuilderConfig // config of the metrics builder.
	startTime                        pcommon.Timestamp    // start time that will be applied to all recorded data points.
	metricsCapacity                  int                  // maximum observed number of metrics per resource.
	metricsBuffer                    pmetric.Metrics      // accumulates metrics data before emitting.
	buildInfo                        component.BuildInfo  // contains version information.
	resourceAttributeIncludeFilter   map[string]filter.Filter
	resourceAttributeExcludeFilter   map[string]filter.Filter
	metricKymaModuleStatusConditions metricKymaModuleStatusConditions
	metricKymaModuleStatusState      metricKymaModuleStatusState
}

// metricBuilderOption applies changes to default metrics builder.
type metricBuilderOption func(*MetricsBuilder)

// WithStartTime sets startTime on the metrics builder.
func WithStartTime(startTime pcommon.Timestamp) metricBuilderOption {
	return func(mb *MetricsBuilder) {
		mb.startTime = startTime
	}
}

func NewMetricsBuilder(mbc MetricsBuilderConfig, settings receiver.Settings, options ...metricBuilderOption) *MetricsBuilder {
	mb := &MetricsBuilder{
		config:                           mbc,
		startTime:                        pcommon.NewTimestampFromTime(time.Now()),
		metricsBuffer:                    pmetric.NewMetrics(),
		buildInfo:                        settings.BuildInfo,
		metricKymaModuleStatusConditions: newMetricKymaModuleStatusConditions(mbc.Metrics.KymaModuleStatusConditions),
		metricKymaModuleStatusState:      newMetricKymaModuleStatusState(mbc.Metrics.KymaModuleStatusState),
		resourceAttributeIncludeFilter:   make(map[string]filter.Filter),
		resourceAttributeExcludeFilter:   make(map[string]filter.Filter),
	}
	if mbc.ResourceAttributes.K8sNamespaceName.MetricsInclude != nil {
		mb.resourceAttributeIncludeFilter["k8s.namespace.name"] = filter.CreateFilter(mbc.ResourceAttributes.K8sNamespaceName.MetricsInclude)
	}
	if mbc.ResourceAttributes.K8sNamespaceName.MetricsExclude != nil {
		mb.resourceAttributeExcludeFilter["k8s.namespace.name"] = filter.CreateFilter(mbc.ResourceAttributes.K8sNamespaceName.MetricsExclude)
	}
	if mbc.ResourceAttributes.KymaModuleName.MetricsInclude != nil {
		mb.resourceAttributeIncludeFilter["kyma.module.name"] = filter.CreateFilter(mbc.ResourceAttributes.KymaModuleName.MetricsInclude)
	}
	if mbc.ResourceAttributes.KymaModuleName.MetricsExclude != nil {
		mb.resourceAttributeExcludeFilter["kyma.module.name"] = filter.CreateFilter(mbc.ResourceAttributes.KymaModuleName.MetricsExclude)
	}

	for _, op := range options {
		op(mb)
	}
	return mb
}

// NewResourceBuilder returns a new resource builder that should be used to build a resource associated with for the emitted metrics.
func (mb *MetricsBuilder) NewResourceBuilder() *ResourceBuilder {
	return NewResourceBuilder(mb.config.ResourceAttributes)
}

// updateCapacity updates max length of metrics and resource attributes that will be used for the slice capacity.
func (mb *MetricsBuilder) updateCapacity(rm pmetric.ResourceMetrics) {
	if mb.metricsCapacity < rm.ScopeMetrics().At(0).Metrics().Len() {
		mb.metricsCapacity = rm.ScopeMetrics().At(0).Metrics().Len()
	}
}

// ResourceMetricsOption applies changes to provided resource metrics.
type ResourceMetricsOption func(pmetric.ResourceMetrics)

// WithResource sets the provided resource on the emitted ResourceMetrics.
// It's recommended to use ResourceBuilder to create the resource.
func WithResource(res pcommon.Resource) ResourceMetricsOption {
	return func(rm pmetric.ResourceMetrics) {
		res.CopyTo(rm.Resource())
	}
}

// WithStartTimeOverride overrides start time for all the resource metrics data points.
// This option should be only used if different start time has to be set on metrics coming from different resources.
func WithStartTimeOverride(start pcommon.Timestamp) ResourceMetricsOption {
	return func(rm pmetric.ResourceMetrics) {
		var dps pmetric.NumberDataPointSlice
		metrics := rm.ScopeMetrics().At(0).Metrics()
		for i := 0; i < metrics.Len(); i++ {
			switch metrics.At(i).Type() {
			case pmetric.MetricTypeGauge:
				dps = metrics.At(i).Gauge().DataPoints()
			case pmetric.MetricTypeSum:
				dps = metrics.At(i).Sum().DataPoints()
			}
			for j := 0; j < dps.Len(); j++ {
				dps.At(j).SetStartTimestamp(start)
			}
		}
	}
}

// EmitForResource saves all the generated metrics under a new resource and updates the internal state to be ready for
// recording another set of data points as part of another resource. This function can be helpful when one scraper
// needs to emit metrics from several resources. Otherwise calling this function is not required,
// just `Emit` function can be called instead.
// Resource attributes should be provided as ResourceMetricsOption arguments.
func (mb *MetricsBuilder) EmitForResource(rmo ...ResourceMetricsOption) {
	rm := pmetric.NewResourceMetrics()
	ils := rm.ScopeMetrics().AppendEmpty()
	ils.Scope().SetName("otelcol/kymastats")
	ils.Scope().SetVersion(mb.buildInfo.Version)
	ils.Metrics().EnsureCapacity(mb.metricsCapacity)
	mb.metricKymaModuleStatusConditions.emit(ils.Metrics())
	mb.metricKymaModuleStatusState.emit(ils.Metrics())

	for _, op := range rmo {
		op(rm)
	}
	for attr, filter := range mb.resourceAttributeIncludeFilter {
		if val, ok := rm.Resource().Attributes().Get(attr); ok && !filter.Matches(val.AsString()) {
			return
		}
	}
	for attr, filter := range mb.resourceAttributeExcludeFilter {
		if val, ok := rm.Resource().Attributes().Get(attr); ok && filter.Matches(val.AsString()) {
			return
		}
	}

	if ils.Metrics().Len() > 0 {
		mb.updateCapacity(rm)
		rm.MoveTo(mb.metricsBuffer.ResourceMetrics().AppendEmpty())
	}
}

// Emit returns all the metrics accumulated by the metrics builder and updates the internal state to be ready for
// recording another set of metrics. This function will be responsible for applying all the transformations required to
// produce metric representation defined in metadata and user config, e.g. delta or cumulative.
func (mb *MetricsBuilder) Emit(rmo ...ResourceMetricsOption) pmetric.Metrics {
	mb.EmitForResource(rmo...)
	metrics := mb.metricsBuffer
	mb.metricsBuffer = pmetric.NewMetrics()
	return metrics
}

// RecordKymaModuleStatusConditionsDataPoint adds a data point to kyma.module.status.conditions metric.
func (mb *MetricsBuilder) RecordKymaModuleStatusConditionsDataPoint(ts pcommon.Timestamp, val int64, reasonAttributeValue string, statusAttributeValue string, typeAttributeValue string) {
	mb.metricKymaModuleStatusConditions.recordDataPoint(mb.startTime, ts, val, reasonAttributeValue, statusAttributeValue, typeAttributeValue)
}

// RecordKymaModuleStatusStateDataPoint adds a data point to kyma.module.status.state metric.
func (mb *MetricsBuilder) RecordKymaModuleStatusStateDataPoint(ts pcommon.Timestamp, val int64, stateAttributeValue string) {
	mb.metricKymaModuleStatusState.recordDataPoint(mb.startTime, ts, val, stateAttributeValue)
}

// Reset resets metrics builder to its initial state. It should be used when external metrics source is restarted,
// and metrics builder should update its startTime and reset it's internal state accordingly.
func (mb *MetricsBuilder) Reset(options ...metricBuilderOption) {
	mb.startTime = pcommon.NewTimestampFromTime(time.Now())
	for _, op := range options {
		op(mb)
	}
}
