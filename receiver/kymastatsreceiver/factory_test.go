package kymastatsreceiver

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

func TestValidConfig(t *testing.T) {
	factory := NewFactory()
	err := componenttest.CheckConfigStruct(factory.CreateDefaultConfig())
	require.NoError(t, err)
}

func TestCreateMetricsReceiver(t *testing.T) {
	factory := NewFactory()
	metricsReceiver, err := factory.CreateMetricsReceiver(
		context.Background(),
		receivertest.NewNopSettings(),
		kubeConfig(),
		consumertest.NewNop(),
	)
	require.NoError(t, err)
	require.NotNil(t, metricsReceiver)
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

func kubeConfig() *Config {
	scheme := runtime.NewScheme()
	return &Config{
		ControllerConfig: scraperhelper.ControllerConfig{
			CollectionInterval: 10 * time.Second,
			InitialDelay:       time.Second,
		},

		APIConfig: k8sconfig.APIConfig{
			AuthType: "kubeConfig",
		},
		MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
		Resources:            internal.NewDefaultResourceConfiguration(),
		makeDynamicClient:    func() (dynamic.Interface, error) { return fake.NewSimpleDynamicClient(scheme), nil },
	}
}
