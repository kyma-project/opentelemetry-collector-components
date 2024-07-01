package singletonreceivercreator

import (
	"context"
	"testing"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/dummyreceiver"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

// nopHost mocks a receiver.ReceiverHost for test purposes.
type mockHost struct {
	receivers *receiver.Builder
}

var mockReceiverConfig = receiverConfig{
	id: component.NewIDWithName(component.MustNewType("dummy"), "name"),
	config: map[string]any{
		"interval": "1m",
	},
}

var defaultCfg = &Config{
	leaderElectionConfig: leaderElectionConfig{
		leaseName:      "singleton-receiver",
		leaseNamespace: "default",
		leaseDuration:  defaultLeaseDuration,
		renewDuration:  defaultRenewDeadline,
		retryPeriod:    defaultRetryPeriod,
	},
	subreceiverConfig: mockReceiverConfig,
}

func createMockMetricsReceiver(_ context.Context, params receiver.Settings, cfg component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	return nil, nil //nolint:nilnil // required during testing
}

// NewNopHost returns a new instance of nopHost with proper defaults for most tests.
func NewMockHost() (component.Host, error) {

	var factories map[component.Type]receiver.Factory
	var err error

	factories, err = receiver.MakeFactoryMap([]receiver.Factory{
		dummyreceiver.NewFactory(),
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
	r := newReceiverRunner(receivertest.NewNopSettings(), mh)

	require.NoError(t, r.start(mockReceiverConfig, consumertest.NewNop()))
	require.NoError(t, r.shutdown(context.Background()))
}

func TestLoadReceiverConfig(t *testing.T) {
	mh, err := NewMockHost()
	require.NoError(t, err)
	r := newReceiverRunner(receivertest.NewNopSettings(), mh)
	factory := mh.GetFactory(component.KindReceiver, component.MustNewType("dummy"))
	recvrFact := factory.(receiver.Factory)

	cfg, err := r.loadReceiverConfig(recvrFact, mockReceiverConfig)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	expectedCfg := &dummyreceiver.Config{
		Interval: "1m",
	}
	require.Equal(t, expectedCfg, cfg)
}
