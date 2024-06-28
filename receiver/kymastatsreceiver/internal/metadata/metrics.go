package metadata

import (
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MetricsBuilders struct {
	TelemetryMetricsBuilder *MetricsBuilder
}

type MetricsBuilder struct {
	StartTime     pcommon.Timestamp
	MetricsBuffer pmetric.Metrics
	Config        MetricsBuilderConfig
}

type MetricsBuilderConfig struct {
	KymaTelemetryModuleStat []MetricConfig
}

type MetricConfig struct {
	Name            string
	Description     string
	Unit            string
	ResourceGroup   string
	ResourceVersion string
	ResourceName    string
}

type Metric struct {
	data     pmetric.Metric
	config   MetricConfig
	capacity int
}

func (m *Metric) initGauge(cfg MetricConfig) {
	m.data.SetName(cfg.Name)
	m.data.SetDescription(cfg.Description)
	m.data.SetUnit(cfg.Unit)
	m.data.SetEmptyGauge()
}

func (m *Metric) recordGaugeDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val float64) {
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetDoubleValue(val)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *Metric) updateGaugeCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *Metric) emitGauge(metrics pmetric.MetricSlice) {
	if m.data.Gauge().DataPoints().Len() > 0 {
		m.updateGaugeCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.initGauge(m.config)
	}
}

func NewGaugeMetric(cfg MetricConfig) Metric {
	m := Metric{config: cfg}

	m.data = pmetric.NewMetric()
	m.initGauge(cfg)

	return m
}

func NewMetricsBuilder(cfg MetricsBuilderConfig) *MetricsBuilder {
	return &MetricsBuilder{
		StartTime:     pcommon.NewTimestampFromTime(time.Now()),
		MetricsBuffer: pmetric.NewMetrics(),
		Config:        cfg,
	}
}
