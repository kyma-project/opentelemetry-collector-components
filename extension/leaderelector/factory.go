package leaderelector

import (
	"context"
	"github.com/kyma-project/opentelemetry-collector-components/extension/leaderelector/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"time"
)

// CreateDefaultConfig returns the default configuration for the extension.
func CreateDefaultConfig() component.Config {
	return &Config{
		leaseName:      "my-lease",
		leaseNamespace: "default",
		leaseDuration:  15 * time.Second,
		renewDuration:  10 * time.Second,
		retryPeriod:    2 * time.Second,
		// Set default values for your configuration
	}
}

// CreateExtension creates the extension instance based on the configuration.
func CreateExtension(
	ctx context.Context,
	set extension.Settings,
	cfg component.Config,
) (extension.Extension, error) {
	return &leaderElectionExtension{
		config: cfg.(*Config),
		logger: set.Logger,
	}, nil
}

// NewFactory creates a new factory for your extension.
func NewFactory() extension.Factory {
	return extension.NewFactory(
		metadata.Type,
		CreateDefaultConfig,
		CreateExtension,
		metadata.ExtensionStability,
	)
}
