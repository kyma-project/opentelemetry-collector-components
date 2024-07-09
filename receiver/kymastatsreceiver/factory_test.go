package kymastatsreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

func TestValidConfig(t *testing.T) {
	factory := NewFactory()
	err := componenttest.CheckConfigStruct(factory.CreateDefaultConfig())
	require.NoError(t, err)
}

func TestCreateMetricsReceiver(t *testing.T) {
	tests := []struct {
		name        string
		cfg         component.Config
		expectedErr bool
	}{
		{
			name: "valid",
			cfg: &Config{
				APIConfig: k8sconfig.APIConfig{
					AuthType: "kubeConfig",
				},
				ControllerConfig: scraperhelper.ControllerConfig{
					CollectionInterval: 10 * time.Second,
					InitialDelay:       time.Second,
				},
				MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
				makeDynamicClient:    func() (dynamic.Interface, error) { return fake.NewSimpleDynamicClient(runtime.NewScheme()), nil },
			},
		},
		{
			name:        "invalid",
			cfg:         component.Config([]byte{1, 2, 3}),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewFactory()
			metricsReceiver, err := factory.CreateMetricsReceiver(
				context.Background(),
				receivertest.NewNopSettings(),
				tt.cfg,
				consumertest.NewNop(),
			)
			if tt.expectedErr {
				require.Error(t, err)
				require.Nil(t, metricsReceiver)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, metricsReceiver)
		})
	}
}

func TestCreateTraceReceiver(t *testing.T) {
	factory := NewFactory()
	traceReceiver, err := factory.CreateTracesReceiver(
		context.Background(),
		receivertest.NewNopSettings(),
		&Config{
			APIConfig: k8sconfig.APIConfig{
				AuthType: "kubeConfig",
			},
		},
		nil,
	)
	require.ErrorIs(t, err, component.ErrDataTypeIsNotSupported)
	require.Nil(t, traceReceiver)
}

func TestCreateLogsReceiver(t *testing.T) {
	factory := NewFactory()
	logsReceiver, err := factory.CreateLogsReceiver(
		context.Background(),
		receivertest.NewNopSettings(),
		&Config{
			APIConfig: k8sconfig.APIConfig{
				AuthType: "kubeConfig",
			},
		},
		nil,
	)
	require.ErrorIs(t, err, component.ErrDataTypeIsNotSupported)
	require.Nil(t, logsReceiver)
}

func TestFactoryBadAuthType(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{
		APIConfig: k8sconfig.APIConfig{
			AuthType: "none",
		},
	}
	_, err := factory.CreateMetricsReceiver(
		context.Background(),
		receivertest.NewNopSettings(),
		cfg,
		consumertest.NewNop(),
	)
	require.Error(t, err)
}

func TestFactoryNoneAuthType(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "somehost")
	t.Setenv("KUBERNETES_SERVICE_PORT", "443")
	factory := NewFactory()
	cfg := &Config{
		APIConfig: k8sconfig.APIConfig{
			AuthType: "none",
		},
		ControllerConfig: scraperhelper.ControllerConfig{
			CollectionInterval: 10 * time.Second,
		},
	}
	_, err := factory.CreateMetricsReceiver(
		context.Background(),
		receivertest.NewNopSettings(),
		cfg,
		consumertest.NewNop(),
	)
	require.NoError(t, err)
}
