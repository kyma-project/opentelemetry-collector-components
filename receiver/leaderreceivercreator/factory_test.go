package leaderreceivercreator

import (
	"context"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/leaderreceivercreator/internal/k8sconfig"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/leaderreceivercreator/internal/metadata"
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
					receivertest.NewNopCreateSettings(),
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
					leaderElectionConfig: leaderElectionConfig{
						authType:             k8sconfig.AuthTypeServiceAccount,
						leaseName:            "my-lease",
						leaseNamespace:       "default",
						leaseDurationSeconds: defaultLeaseDuration,
						renewDeadlineSeconds: defaultRenewDeadline,
						retryPeriodSeconds:   defaultRetryPeriod,
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
