package singletonreceivercreator

import (
	"context"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

const (
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
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
func newLeaderElector(client kubernetes.Interface, onStartedLeading func(context.Context), onStoppedLeading func(), cfg leaderElectionConfig) (*leaderelection.LeaderElector, error) {
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

	return leaderelection.NewLeaderElector(leConfig)
}
