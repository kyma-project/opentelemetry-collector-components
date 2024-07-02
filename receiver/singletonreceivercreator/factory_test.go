package singletonreceivercreator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/metadata"
)

func TestNewFactory(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "creates a new factory with correct type",
			testFunc: func(t *testing.T) {
				factory := NewFactory()
				ft := factory.Type()
				require.EqualValues(t, metadata.Type, ft)
			},
		}, {
			desc: "creates a new factory and CreateMetricsReceiver returns no error",
			testFunc: func(t *testing.T) {
				cfg := createDefaultConfig().(*Config)
				_, err := NewFactory().CreateMetricsReceiver(
					context.Background(),
					receivertest.NewNopSettings(),
					cfg,
					consumertest.NewNop(),
				)
				require.NoError(t, err)
			},
		}, {
			desc: "creates a new factory and CreateMetricsReceiver with default config",
			testFunc: func(t *testing.T) {
				factory := NewFactory()
				expectedCfg := &Config{
					APIConfig: k8sconfig.APIConfig{
						AuthType: "serviceAccount",
					},
					leaderElectionConfig: leaderElectionConfig{
						leaseName:      "singleton-receiver",
						leaseNamespace: "default",
						leaseDuration:  defaultLeaseDuration,
						renewDuration:  defaultRenewDeadline,
						retryPeriod:    defaultRetryPeriod,
					},
					subreceiverConfig: receiverConfig{},
				}

				require.Equal(t, expectedCfg, factory.CreateDefaultConfig())
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.desc, test.testFunc)
	}
}
