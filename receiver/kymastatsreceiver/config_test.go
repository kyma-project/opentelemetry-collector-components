package kymastatsreceiver

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	duration := time.Minute
	delay := time.Second

	tests := []struct {
		id        component.ID
		expected  component.Config
		expectErr bool
	}{
		{
			id: component.NewIDWithName(metadata.Type, ""),
			expected: &Config{
				APIConfig: k8sconfig.APIConfig{
					AuthType: "serviceAccount",
				},
				ControllerConfig:     scraperhelper.ControllerConfig{CollectionInterval: duration, InitialDelay: delay},
				MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
				Modules: []ModuleConfig{
					{
						Group:    "operator.kyma-project.io",
						Version:  "v1alpha1",
						Resource: "telemetries",
					},
				},
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "kubeconfig"),
			expected: &Config{
				APIConfig: k8sconfig.APIConfig{
					AuthType: "kubeConfig",
					Context:  "k8s-context",
				},
				ControllerConfig:     scraperhelper.ControllerConfig{CollectionInterval: 30 * time.Second, InitialDelay: delay},
				MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
				Modules: []ModuleConfig{
					{
						Group:    "operator.kyma-project.io",
						Version:  "v1alpha1",
						Resource: "telemetries",
					},
				},
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "sa"),
			expected: &Config{
				APIConfig: k8sconfig.APIConfig{
					AuthType: "serviceAccount",
				},
				ControllerConfig:     scraperhelper.ControllerConfig{CollectionInterval: 10 * time.Second, InitialDelay: delay},
				MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
				Modules: []ModuleConfig{
					{
						Group:    "operator.kyma-project.io",
						Version:  "v1alpha1",
						Resource: "telemetries",
					},
				},
			},
		},
		{
			id:        component.NewIDWithName(metadata.Type, "invalidauth"),
			expectErr: true,
		},
		{
			id:        component.NewIDWithName(metadata.Type, "invalidinterval"),
			expectErr: true,
		},
		{
			id: component.NewIDWithName(metadata.Type, "none"),
			expected: &Config{
				APIConfig: k8sconfig.APIConfig{
					AuthType: "none",
				},
				ControllerConfig:     scraperhelper.ControllerConfig{CollectionInterval: duration, InitialDelay: delay},
				MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
				Modules: []ModuleConfig{
					{
						Group:    "operator.kyma-project.io",
						Version:  "v1alpha1",
						Resource: "telemetries",
					},
				},
			},
		},
		{
			id:        component.NewIDWithName(metadata.Type, "nomodulegroups"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			t.Parallel()
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(&cfg))
			err = component.ValidateConfig(cfg)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
