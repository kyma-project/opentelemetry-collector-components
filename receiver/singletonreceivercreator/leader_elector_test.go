package singletonreceivercreator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/metadata"
)

func TestLeaderElector(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	onStartedLeading := func(ctx context.Context) {}
	onStoppedLeading := func() {}
	leConfig := leaderElectionConfig{
		leaseName:      "foo",
		leaseNamespace: "bar",
		leaseDuration:  10,
		renewDuration:  5,
		retryPeriod:    2,
	}

	telemetryBuilder, err := metadata.NewTelemetryBuilder(componenttest.NewNopTelemetrySettings())
	require.NoError(t, err)

	leaderElector, err := newLeaderElector(leConfig, fakeClient, telemetryBuilder, onStartedLeading, onStoppedLeading)
	require.NoError(t, err)
	require.NotNil(t, leaderElector)
}
