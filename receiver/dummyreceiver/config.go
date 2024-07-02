package dummyreceiver

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrParsingInterval  = errors.New("interval must be a valid duration string")
	ErrIntervalTooShort = errors.New("when defined, the interval has to be set to at least 1 minute (1m)")
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
		return fmt.Errorf("%w: %w", ErrParsingInterval, err)
	}

	if interval.Minutes() < 1 {
		return ErrIntervalTooShort
	}

	return nil
}
