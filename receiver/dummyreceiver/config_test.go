package dummyreceiver

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/dummyreceiver/internal/metadata"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		passed   component.Config
		expected component.Config
		err      error
	}{

		{
			name:   "check-passed-values",
			passed: &Config{Interval: "2m"},
			err:    nil,
		}, {
			name:   "check-invalid-interval",
			passed: &Config{Interval: "foo"},
			err:    ErrParsingInterval,
		}, {
			name:   "check-interval-less-than-1m",
			passed: &Config{Interval: "30s"},
			err:    ErrIntervalTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.passed.(*Config)
			err := cfg.Validate()
			require.ErrorIs(t, err, tt.err)
		})
	}
}

func TestCreateMetricsReceiverInvalidConfig(t *testing.T) {
	_, err := createMetricsReceiver(t.Context(), receivertest.NewNopSettings(metadata.Type), "not-a-config", consumertest.NewNop())
	require.Error(t, err)
}
