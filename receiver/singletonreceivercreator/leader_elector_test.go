package singletonreceivercreator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLeaderElector(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	onStartedLeading := func(ctx context.Context) {}
	onStoppedLeading := func() {}
	lec := leaderElectionConfig{
		leaseName:            "foo",
		leaseNamespace:       "bar",
		leaseDurationSeconds: 10,
		renewDeadlineSeconds: 5,
		retryPeriodSeconds:   2,
	}
	leaderElector, err := newLeaderElector(fakeClient, onStartedLeading, onStoppedLeading, lec, nil)
	require.NoError(t, err)
	require.NotNil(t, leaderElector)
}
