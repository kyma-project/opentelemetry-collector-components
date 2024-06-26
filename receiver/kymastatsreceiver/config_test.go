package kymastatsreceiver

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	duration := 30 * time.Second

	tests := []struct {
		id          component.ID
		expected    component.Config
		expectedErr error
	}{
		{
			id: component.NewIDWithName(metadata.Type, "default"),
			expected: &Config{

				CollectionInterval: duration,
				APIConfig: k8sconfig.APIConfig{
					AuthType: "serviceAccount",
				},
			},
		},

		{
			id: component.NewIDWithName(metadata.Type, "k8s"),
			expected: &Config{

				CollectionInterval: duration,
				APIConfig: k8sconfig.APIConfig{
					AuthType: "kubeConfig",
				},
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "custom"),
			expected: &Config{

				CollectionInterval: 10 * time.Second,
				APIConfig: k8sconfig.APIConfig{
					AuthType: "serviceAccount",
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
			require.NoError(t, component.UnmarshalConfig(sub, cfg))

			assert.NoError(t, component.ValidateConfig(cfg))
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
