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
		leaseName:      "foo",
		leaseNamespace: "bar",
		leaseDuration:  10,
		renewDuration:  5,
		retryPeriod:    2,
	}
	leaderElector, err := newLeaderElector(fakeClient, onStartedLeading, onStoppedLeading, lec, "hos1")
	require.NoError(t, err)
	require.NotNil(t, leaderElector)
}
