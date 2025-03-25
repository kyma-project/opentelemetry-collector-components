package servicenameenrichmentprocessor

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/kyma-project/opentelemetry-collector-components/processor/serviceenrichmentprocessor/internal/metadata"
)

func TestNewFactory(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "creates a new factory with correct type",
			testFunc: func(t *testing.T) {
				t.Helper()
				factory := NewFactory()
				ft := factory.Type()
				require.EqualValues(t, metadata.Type, ft)
			},
		}, {
			desc: "creates a new factory and CreateMetricsReceiver returns no error",
			testFunc: func(t *testing.T) {
				t.Helper()
				cfg := createDefaultConfig().(Config)
				_, err := NewFactory().CreateMetrics(
					t.Context(),
					processortest.NewNopSettingsWithType(metadata.Type),
					cfg,
					consumertest.NewNop(),
				)
				require.NoError(t, err)
			},
		}, {
			desc: "creates a new factory and CreatesTracerReceiver returns no error",
			testFunc: func(t *testing.T) {
				t.Helper()
				cfg := createDefaultConfig().(Config)
				_, err := NewFactory().CreateTraces(
					t.Context(),
					processortest.NewNopSettingsWithType(metadata.Type),
					cfg,
					consumertest.NewNop(),
				)
				require.NoError(t, err)
			},
		},
		{
			desc: "creates a new factory and CreatesLogReceiver returns no error",
			testFunc: func(t *testing.T) {
				t.Helper()
				cfg := createDefaultConfig().(Config)
				_, err := NewFactory().CreateLogs(
					t.Context(),
					processortest.NewNopSettingsWithType(metadata.Type),
					cfg,
					consumertest.NewNop(),
				)
				require.NoError(t, err)
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.desc, test.testFunc)
	}
}
