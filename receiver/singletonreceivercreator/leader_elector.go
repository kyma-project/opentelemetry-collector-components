package singletonreceivercreator

import (
	"context"
	"os"
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

// NewResourceLock creates a new leases resource lock for use in a leader election loop
func newResourceLock(client kubernetes.Interface, leaderElectionNamespace, lockName string) (resourcelock.Interface, error) {
	// Leader id, needs to be unique, use pod name in kubernetes case.
	id, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return resourcelock.New(
		resourcelock.LeasesResourceLock,
		leaderElectionNamespace,
		lockName,
		client.CoreV1(),
		client.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: id,
		})
}

// newLeaderElector return  a leader elector object using client-go
func newLeaderElector(
	cfg leaderElectionConfig,
	client kubernetes.Interface,
	telemetryBuilder *metadata.TelemetryBuilder,
	onStartedLeading func(context.Context),
	onStoppedLeading func(),
) (*leaderelection.LeaderElector, error) {
	namespace := cfg.leaseNamespace
	lockName := cfg.leaseName

	resourceLock, err := newResourceLock(client, namespace, lockName)
	if err != nil {
		return &leaderelection.LeaderElector{}, err
	}

	leConfig := leaderelection.LeaderElectionConfig{
		Lock:          resourceLock,
		LeaseDuration: cfg.leaseDuration,
		RenewDeadline: cfg.renewDuration,
		RetryPeriod:   cfg.retryPeriod,
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
