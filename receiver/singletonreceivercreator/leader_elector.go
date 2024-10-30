package singletonreceivercreator

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/metadata"
)

const (
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
	leaseAttrKey         = "lease"
)

// newLeaderElector return  a leader elector object using client-go
func newLeaderElector(
	cfg leaderElectionConfig,
	client kubernetes.Interface,
	telemetryBuilder *metadata.TelemetryBuilder,
	onStartedLeading func(context.Context),
	onStoppedLeading func(),
	identity string,
) (*leaderelection.LeaderElector, error) {
	resourceLock, err := resourcelock.New(
		resourcelock.LeasesResourceLock,
		cfg.leaseNamespace,
		cfg.leaseName,
		client.CoreV1(),
		client.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: identity,
		})
	if err != nil {
		return &leaderelection.LeaderElector{}, err
	}

	leConfig := leaderelection.LeaderElectionConfig{
		// The lock resource name is used as a lease label in leader election metrics.
		Name:            cfg.leaseName,
		Lock:            resourceLock,
		LeaseDuration:   cfg.leaseDuration,
		RenewDeadline:   cfg.renewDuration,
		RetryPeriod:     cfg.retryPeriod,
		ReleaseOnCancel: true,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: onStartedLeading,
			OnStoppedLeading: onStoppedLeading,
		},
	}

	leaderelection.SetProvider(leaderMetricProvider{
		telemetryBuilder: telemetryBuilder,
	})

	return leaderelection.NewLeaderElector(leConfig)
}

type leaderMetricProvider struct {
	telemetryBuilder *metadata.TelemetryBuilder
}

func (l leaderMetricProvider) NewLeaderMetric() leaderelection.LeaderMetric {
	return leaderMetric(l)
}

type leaderMetric struct {
	telemetryBuilder *metadata.TelemetryBuilder
}

func (lm leaderMetric) On(name string) {
	ctx := context.Background()
	lm.telemetryBuilder.ReceiverSingletonLeaseAcquiredTotal.Add(ctx, 1, metric.WithAttributes(attribute.String(leaseAttrKey, name)))
	lm.telemetryBuilder.ReceiverSingletonLeaderStatus.Record(ctx, 1, metric.WithAttributes(attribute.String(leaseAttrKey, name)))
}

func (lm leaderMetric) Off(name string) {
	ctx := context.Background()
	lm.telemetryBuilder.ReceiverSingletonLeaseLostTotal.Add(ctx, 1, metric.WithAttributes(attribute.String(leaseAttrKey, name)))
	lm.telemetryBuilder.ReceiverSingletonLeaderStatus.Record(ctx, 0, metric.WithAttributes(attribute.String(leaseAttrKey, name)))
}

func (lm leaderMetric) SlowpathExercised(name string) {
	ctx := context.Background()
	lm.telemetryBuilder.ReceiverSingletonLeaseSlowpathTotal.Add(ctx, 1, metric.WithAttributes(attribute.String(leaseAttrKey, name)))
}
