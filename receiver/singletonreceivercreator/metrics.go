package singletonreceivercreator

import (
	"context"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/metadata"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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
	p.telemetryBuilder.SingletonreceivercreatorLeaseAcquireTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("lease", name)))
}

func (p LeaderMetricProvider) Off(name string) {
	if p.telemetryBuilder == nil {
		return
	}

	ctx := context.Background()
	p.telemetryBuilder.SingletonreceivercreatorLeaseLostTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("lease", name)))
}

func (p LeaderMetricProvider) SlowpathExercised(name string) {
	if p.telemetryBuilder == nil {
		return
	}

	ctx := context.Background()
	p.telemetryBuilder.SingletonreceivercreatorLeaseSlowpathExcerciseTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("lease", name)))
}
