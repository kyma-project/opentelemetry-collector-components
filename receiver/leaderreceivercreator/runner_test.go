package leaderreceivercreator

import (
	"context"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/leaderreceivercreator/internal/metadata"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"testing"
)

// nopHost mocks a receiver.ReceiverHost for test purposes.
type mockHost struct {
	receivers *receiver.Builder
}

var mockReceiverConfig = receiverConfig{
	id: component.NewIDWithName(component.MustNewType("foo"), "name"),
	config: map[string]any{
		"protocols": map[string]any{
			"grpc": nil,
		},
	},
}

var defaultCfg = &Config{
	leaderElectionConfig: leaderElectionConfig{
		leaseName:            "my-lease",
		leaseNamespace:       "default",
		leaseDurationSeconds: defaultLeaseDuration,
		renewDeadlineSeconds: defaultRenewDeadline,
		retryPeriodSeconds:   defaultRetryPeriod,
	},
	subreceiverConfig: mockReceiverConfig,
}

func createMockMetricsReceiver(_ context.Context, params receiver.CreateSettings, cfg component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	return nil, nil
}

// NewNopHost returns a new instance of nopHost with proper defaults for most tests.
func NewMockHost() (component.Host, error) {

	factories, err := receiver.MakeFactoryMap([]receiver.Factory{
		receiver.NewFactory(component.MustNewType("foo"), func() component.Config { return &defaultCfg }, receiver.WithMetrics(createMockMetricsReceiver, metadata.MetricsStability)),
	}...)
	if err != nil {
		return nil, err
	}
	cfg := map[component.ID]component.Config{component.MustNewID("foo"): struct{}{}}
	return &mockHost{
		receivers: receiver.NewBuilder(cfg, factories),
	}, nil
}

func (nh *mockHost) GetFactory(kind component.Kind, t component.Type) component.Factory {
	return nh.receivers.Factory(t)
}

func (nh *mockHost) GetExtensions() map[component.ID]component.Component {
	return nil
}

func (nh *mockHost) GetExporters() map[component.DataType]map[component.ID]component.Component {
	return nil
}

func TestRunnerStart(t *testing.T) {
	mh, err := NewMockHost()
	require.NoError(t, err)
	r := newReceiverRunner(receivertest.NewNopCreateSettings(), mh)

	require.NoError(t, r.start(mockReceiverConfig, consumertest.NewNop()))
	require.NoError(t, r.shutdown(context.Background()))
}

func TestLoadReceiverConfig(t *testing.T) {
	mh, err := NewMockHost()
	require.NoError(t, err)
	r := newReceiverRunner(receivertest.NewNopCreateSettings(), mh)
	factory := mh.GetFactory(component.KindReceiver, component.MustNewType("foo"))
	recvrFact := factory.(receiver.Factory)

	cfg, _, err := r.loadReceiverConfig(recvrFact, mockReceiverConfig)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	rcvrCfg := cfg.(**Config)
	require.NotNil(t, mockReceiverConfig, (*rcvrCfg).subreceiverConfig.config)
}
