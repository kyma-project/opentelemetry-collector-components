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

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "expected_config.yaml"))
	require.NoError(t, err)

	duration := time.Minute
	delay := time.Second

	tests := []struct {
		id        component.ID
		expected  component.Config
		expectErr bool
	}{
		{
			id: component.NewIDWithName(metadata.Type, "default"),
			expected: &Config{
				ControllerConfig: scraperhelper.ControllerConfig{CollectionInterval: duration, InitialDelay: delay},
				APIConfig: k8sconfig.APIConfig{
					AuthType: "serviceAccount",
				},
				MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
			},
		},

		{
			id: component.NewIDWithName(metadata.Type, "k8s"),
			expected: &Config{
				ControllerConfig: scraperhelper.ControllerConfig{CollectionInterval: 30 * time.Second, InitialDelay: delay},
				APIConfig: k8sconfig.APIConfig{
					AuthType: "kubeConfig",
					Context:  "k8s-context",
				},
				MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "sa"),
			expected: &Config{
				ControllerConfig: scraperhelper.ControllerConfig{CollectionInterval: 10 * time.Second, InitialDelay: delay},
				APIConfig: k8sconfig.APIConfig{
					AuthType: "serviceAccount",
				},
				MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
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
				ControllerConfig: scraperhelper.ControllerConfig{CollectionInterval: duration, InitialDelay: delay},
				APIConfig: k8sconfig.APIConfig{
					AuthType: "none",
				},
				MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
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
