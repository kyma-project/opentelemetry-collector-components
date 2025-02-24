package singletonreceivercreator

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/confmap/xconfmap"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/metadata"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	tests := []struct {
		id          component.ID
		expected    component.Config
		expectedErr error
	}{
		{
			id: component.NewIDWithName(metadata.Type, "default"),
			expected: &Config{
				APIConfig: k8sconfig.APIConfig{
					AuthType: "serviceAccount",
				},
				leaderElectionConfig: leaderElectionConfig{
					leaseName:      "foo",
					leaseNamespace: "bar",
					leaseDuration:  defaultLeaseDuration,
					renewDuration:  defaultRenewDeadline,
					retryPeriod:    defaultRetryPeriod,
				},
				subreceiverConfig: receiverConfig{
					id:     component.MustNewID("dummy"),
					config: make(map[string]any),
				},
			},
		},
		{
			id:          component.NewIDWithName(metadata.Type, "missing_name"),
			expectedErr: errMissingLeaseName,
		},
		{
			id:          component.NewIDWithName(metadata.Type, "missing_namespace"),
			expectedErr: errMissingLeaseNamespace,
		},
		{
			id:          component.NewIDWithName(metadata.Type, "zero_lease_duration"),
			expectedErr: errNonPositiveInterval,
		},
		{
			id:          component.NewIDWithName(metadata.Type, "zero_renew_deadline"),
			expectedErr: errNonPositiveInterval,
		},
		{
			id:          component.NewIDWithName(metadata.Type, "zero_retry_period"),
			expectedErr: errNonPositiveInterval,
		},
		{
			id: component.NewIDWithName(metadata.Type, "complex_subreceiver"),
			expected: &Config{
				APIConfig: k8sconfig.APIConfig{
					AuthType: "serviceAccount",
				},
				leaderElectionConfig: leaderElectionConfig{
					leaseName:      "foo",
					leaseNamespace: "bar",
					leaseDuration:  15 * time.Second,
					renewDuration:  10 * time.Second,
					retryPeriod:    2 * time.Second,
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
		{
			id: component.NewIDWithName(metadata.Type, "auth_type_kubeconfig"),
			expected: &Config{
				APIConfig: k8sconfig.APIConfig{
					AuthType: "kubeConfig",
				},
				leaderElectionConfig: leaderElectionConfig{
					leaseName:      "foo",
					leaseNamespace: "bar",
					leaseDuration:  15 * time.Second,
					renewDuration:  10 * time.Second,
					retryPeriod:    2 * time.Second,
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
			t.Parallel()
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))

			err = xconfmap.Validate(cfg)
			if tt.expectedErr == nil {
				require.NoError(t, err)
				require.Equal(t, tt.expected, cfg)
				return
			}

			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
