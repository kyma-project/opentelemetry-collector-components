package leaderreceivercreator

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/leaderreceivercreator/internal/k8sconfig"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/leaderreceivercreator/internal/metadata"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	tests := []struct {
		id       component.ID
		expected component.Config
	}{
		{
			id: component.NewIDWithName(metadata.Type, "check-default-values"),
			expected: &Config{
				leaderElectionConfig: leaderElectionConfig{
					APIConfig: k8sconfig.APIConfig{
						AuthType: k8sconfig.AuthTypeServiceAccount,
					},
					leaseName:            "my-lease",
					leaseNamespace:       "default",
					leaseDurationSeconds: defaultLeaseDuration,
					renewDeadlineSeconds: defaultRenewDeadline,
					retryPeriodSeconds:   defaultRetryPeriod,
				},
				subreceiverConfig: receiverConfig{
					id: component.MustNewID("otlp"),
					config: map[string]any{
						"protocols": map[string]any{
							"grpc": nil,
						},
					},
				},
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "check-all-values"),
			expected: &Config{
				leaderElectionConfig: leaderElectionConfig{
					APIConfig: k8sconfig.APIConfig{
						AuthType: k8sconfig.AuthTypeKubeConfig,
					},
					leaseName:            "foo",
					leaseNamespace:       "bar",
					leaseDurationSeconds: 15 * time.Second,
					renewDeadlineSeconds: 10 * time.Second,
					retryPeriodSeconds:   2 * time.Second,
				},
				subreceiverConfig: receiverConfig{
					id: component.MustNewID("k8s_cluster"),
					config: map[string]any{
						"auth_type":                   "serviceAccount",
						"node_conditions_to_report":   []interface{}{"Ready", "MemoryPressure"},
						"allocatable_types_to_report": []interface{}{"cpu", "memory"},
						"metrics": map[string]any{
							"k8s.container.cpu_limit": map[string]any{
								"enabled": false,
							},
						},
						"resource_attributes": map[string]any{
							"container.id": map[string]any{
								"enabled": false,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))

			assert.NoError(t, component.ValidateConfig(cfg))
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
