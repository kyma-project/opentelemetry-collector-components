package singletonreceivercreator

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/metadata"
)

const (
	leaseAttrKey = "lease"
)

type LeaderMetricProvider struct {
	telemetryBuilder *metadata.TelemetryBuilder
}

func NewLeaderMetricProvider(telemetryBuilder *metadata.TelemetryBuilder) LeaderMetricProvider {
	return LeaderMetricProvider{telemetryBuilder: telemetryBuilder}
}

func (p LeaderMetricProvider) On(name string) {
	if p.telemetryBuilder == nil {
		return
	}

	ctx := context.Background()
	p.telemetryBuilder.ReceiverSingletonLeaseAcquiredTotal.Add(ctx, 1, metric.WithAttributes(attribute.String(leaseAttrKey, name)))
	p.telemetryBuilder.ReceiverSingletonLeaderStatus.Record(ctx, 1, metric.WithAttributes(attribute.String(leaseAttrKey, name)))
}

func (p LeaderMetricProvider) Off(name string) {
	if p.telemetryBuilder == nil {
		return
	}

	ctx := context.Background()
	p.telemetryBuilder.ReceiverSingletonLeaseLostTotal.Add(ctx, 1, metric.WithAttributes(attribute.String(leaseAttrKey, name)))
	p.telemetryBuilder.ReceiverSingletonLeaderStatus.Record(ctx, 0, metric.WithAttributes(attribute.String(leaseAttrKey, name)))
}

func (p LeaderMetricProvider) SlowpathExercised(name string) {
	if p.telemetryBuilder == nil {
		return
	}

	ctx := context.Background()
	p.telemetryBuilder.ReceiverSingletonLeaseSlowpathTotal.Add(ctx, 1, metric.WithAttributes(attribute.String(leaseAttrKey, name)))
}
