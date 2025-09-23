package istioenrichmentprocessor

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/kyma-project/opentelemetry-collector-components/processor/istioenrichmentprocessor/internal/metadata"
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
				require.Equal(t, metadata.Type, ft)
			},
		},
		{
			desc: "creates a new factory and CreatesLogs returns no error",
			testFunc: func(t *testing.T) {
				t.Helper()

				cfg := createDefaultConfig().(Config)
				_, err := NewFactory().CreateLogs(
					t.Context(),
					processortest.NewNopSettings(metadata.Type),
					cfg,
					consumertest.NewNop(),
				)
				require.NoError(t, err)
			},
		},
		{
			desc: "creates a new factory and CreatesLogs with wrong config and returns error",
			testFunc: func(t *testing.T) {
				t.Helper()

				cfg := []string{}
				_, err := NewFactory().CreateLogs(
					t.Context(),
					processortest.NewNopSettings(metadata.Type),
					cfg,
					consumertest.NewNop(),
				)
				require.EqualError(t, err, errInvalidConfig.Error())
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.desc, test.testFunc)
	}
}
