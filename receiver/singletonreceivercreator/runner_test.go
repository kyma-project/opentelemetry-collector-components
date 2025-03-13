package singletonreceivercreator

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/pipeline"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/dummyreceiver"
)

// mockSettings mocks a ReceiverConfigs and ReceiverFactories for test purposes.
type mockSettings struct {
	ReceiversConfigs   map[component.ID]component.Config
	ReceiversFactories map[component.Type]receiver.Factory
}

var mockReceiverConfig = receiverConfig{
	id: component.NewIDWithName(component.MustNewType("dummy"), "name"),
	config: map[string]any{
		"interval": "1m",
	},
}

// NewMockSettings returns a new instance of mockSettings with proper defaults for most tests.
func NewMockSettings() (*mockSettings, error) {

	var factories map[component.Type]receiver.Factory
	var err error
	factories, err = otelcol.MakeFactoryMap[receiver.Factory]([]receiver.Factory{
		dummyreceiver.NewFactory(),
	}...)

	if err != nil {
		return nil, err
	}

	cfg := map[component.ID]component.Config{component.MustNewID("foo"): struct{}{}}
	return &mockSettings{
		ReceiversConfigs:   cfg,
		ReceiversFactories: factories,
	}, nil
}

func (ms *mockSettings) GetFactory(kind component.Kind, t component.Type) component.Factory {
	return ms.ReceiversFactories[t]
}

func (ms *mockSettings) GetExtensions() map[component.ID]component.Component {
	return nil
}

func (ms *mockSettings) GetExporters() map[pipeline.Signal]map[component.ID]component.Component {
	return nil
}

func TestRunnerStart(t *testing.T) {
	ms, err := NewMockSettings()
	require.NoError(t, err)
	r := newReceiverRunner(receivertest.NewNopSettings(receivertest.NopType), ms)

	require.NoError(t, r.start(mockReceiverConfig, consumertest.NewNop()))
	require.NoError(t, r.shutdown(t.Context()))
}

func TestLoadReceiverConfig(t *testing.T) {
	ms, err := NewMockSettings()
	require.NoError(t, err)
	r := newReceiverRunner(receivertest.NewNopSettings(receivertest.NopType), ms)
	factory := ms.GetFactory(component.KindReceiver, component.MustNewType("dummy"))
	recvrFact := factory.(receiver.Factory)

	cfg, err := r.loadReceiverConfig(recvrFact, mockReceiverConfig)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	expectedCfg := &dummyreceiver.Config{
		Interval: "1m",
	}
	require.Equal(t, expectedCfg, cfg)
}

func TestLoadReceiverConfigError(t *testing.T) {
	var factories map[component.Type]receiver.Factory
	var err error

	factories, err = otelcol.MakeFactoryMap([]receiver.Factory{
		receiver.NewFactory(component.MustNewType("foo"), func() component.Config { return &struct{}{} }),
	}...)

	require.NoError(t, err)
	ms := &mockSettings{
		ReceiversFactories: factories,
	}
	require.NoError(t, err)
	r := newReceiverRunner(receivertest.NewNopSettings(receivertest.NopType), ms)
	err = r.start(mockReceiverConfig, consumertest.NewNop())
	require.EqualError(t, err, "unable to lookup factory for wrapped receiver \"dummy/name\"")
}
