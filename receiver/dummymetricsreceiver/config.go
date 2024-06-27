package dummymetricsreceiver

import (
	"fmt"
	"time"
)

// Config represents the receiver config settings within the collector's config.yaml
type Config struct {
	// The number of metrics to generate per interval.
	Interval string `mapstructure:"interval"`
}

// Validate checks if the receiver configuration is valid
func (cfg *Config) Validate() error {
	interval, err := time.ParseDuration(cfg.Interval)
	if err != nil {
		return fmt.Errorf("interval must be a valid duration string: %v", err)
	}

	if interval.Minutes() < 1 {
		return fmt.Errorf("when defined, the interval has to be set to at least 1 minute (1m)")
	}

	return nil
}
