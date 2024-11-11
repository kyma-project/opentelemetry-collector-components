package leaderelector

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
)

// CreateDefaultConfig returns the default configuration for the extension.
func CreateDefaultConfig() component.Config {
	return &Config{
		LeaseDuration: 15 * time.Second,
		RenewDuration: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		// Set default values for your configuration
	}
}

// CreateExtension creates the extension instance based on the configuration.
func CreateExtension(
	ctx context.Context,
	set extension.Settings,
	cfg component.Config,
) (extension.Extension, error) {
	baseCfg, ok := cfg.(*Config)
	if !ok {
		return nil, errors.New("Invalid config, cannot create extension leaderelector")
	}
	fmt.Printf("Creating leaderelector extension with config: %+v\n", baseCfg)

	// Initialize k8s client in factory as doing it in extension.Start()
	// should cause race condition as http Proxy gets shared.
	client, err := k8sconfig.MakeClient(baseCfg.APIConfig)
	if err != nil {
		return nil, errors.New("Failed to create k8s client")
	}

	// Set leaseHolderID for local development

	leaseHolderID, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &leaderElectionExtension{
		config:        baseCfg,
		logger:        set.Logger,
		client:        client,
		leaseHolderID: leaseHolderID,
	}, nil
}

// NewFactory creates a new factory for your extension.
func NewFactory() extension.Factory {
	return extension.NewFactory(
		component.MustNewType("leaderelector"),
		CreateDefaultConfig,
		CreateExtension,
		component.StabilityLevelDevelopment,
	)
}
