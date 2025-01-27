// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"errors"

	"go.opentelemetry.io/otel/metric"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
)

func Meter(settings component.TelemetrySettings) metric.Meter {
	return settings.MeterProvider.Meter("github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator")
}

func Tracer(settings component.TelemetrySettings) trace.Tracer {
	return settings.TracerProvider.Tracer("github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator")
}

// TelemetryBuilder provides an interface for components to report telemetry
// as defined in metadata and user config.
type TelemetryBuilder struct {
	meter                               metric.Meter
	ReceiverSingletonLeaderStatus       metric.Int64Gauge
	ReceiverSingletonLeaseAcquiredTotal metric.Int64Counter
	ReceiverSingletonLeaseLostTotal     metric.Int64Counter
	ReceiverSingletonLeaseSlowpathTotal metric.Int64Counter
}

// TelemetryBuilderOption applies changes to default builder.
type TelemetryBuilderOption interface {
	apply(*TelemetryBuilder)
}

type telemetryBuilderOptionFunc func(mb *TelemetryBuilder)

func (tbof telemetryBuilderOptionFunc) apply(mb *TelemetryBuilder) {
	tbof(mb)
}

// NewTelemetryBuilder provides a struct with methods to update all internal telemetry
// for a component
func NewTelemetryBuilder(settings component.TelemetrySettings, options ...TelemetryBuilderOption) (*TelemetryBuilder, error) {
	builder := TelemetryBuilder{}
	for _, op := range options {
		op.apply(&builder)
	}
	builder.meter = Meter(settings)
	var err, errs error
	builder.ReceiverSingletonLeaderStatus, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Gauge(
		"otelcol_receiver_singleton_leader_status",
		metric.WithDescription("A gauge of if the reporting system is the leader of the relevant lease, 0 indicates backup, and 1 indicates leader."),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.ReceiverSingletonLeaseAcquiredTotal, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_receiver_singleton_lease_acquired_total",
		metric.WithDescription("The total number of successful lease acquisitions."),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.ReceiverSingletonLeaseLostTotal, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_receiver_singleton_lease_lost_total",
		metric.WithDescription("The total number of lease losses."),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.ReceiverSingletonLeaseSlowpathTotal, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_receiver_singleton_lease_slowpath_total",
		metric.WithDescription("The total number of slow paths exercised in renewing leader leases."),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	return &builder, errs
}

func getLeveledMeter(meter metric.Meter, cfgLevel, srvLevel configtelemetry.Level) metric.Meter {
	if cfgLevel <= srvLevel {
		return meter
	}
	return noopmetric.Meter{}
}
