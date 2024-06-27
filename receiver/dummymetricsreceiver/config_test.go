package dummymetricsreceiver

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"testing"
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
			err:    fmt.Errorf("interval must be a valid duration string: time: invalid duration \"foo\""),
		}, {
			name:   "check-interval-less-than-1m",
			passed: &Config{Interval: "30s"},
			err:    fmt.Errorf("when defined, the interval has to be set to at least 1 minute (1m)"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.passed.(*Config)
			err := cfg.Validate()
			require.Equal(t, tt.err, err)
		})
	}
}
