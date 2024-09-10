package kymastatsreceiver

import (
	"errors"

	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

// Config represents the receiver config settings within the collector's config.yaml
type Config struct {
	k8sconfig.APIConfig            `mapstructure:",squash"`
	scraperhelper.ControllerConfig `mapstructure:",squash"`
	metadata.MetricsBuilderConfig  `mapstructure:",squash"`
	Modules                        []ModuleConfig `mapstructure:"modules"`

	// Used for unit testing only
	makeDynamicClient func() (dynamic.Interface, error)
}

type ModuleConfig struct {
	Group    string `mapstructure:"group"`
	Version  string `mapstructure:"version"`
	Resource string `mapstructure:"resource"`
}

var errEmptyModules = errors.New("empty modules")

func (cfg *Config) Validate() error {
	if err := cfg.APIConfig.Validate(); err != nil {
		return err
	}

	if err := cfg.ControllerConfig.Validate(); err != nil {
		return err
	}

	if len(cfg.Modules) == 0 {
		return errEmptyModules
	}

	return nil
}

func (cfg *Config) getDynamicClient() (dynamic.Interface, error) {
	if cfg.makeDynamicClient != nil {
		return cfg.makeDynamicClient()
	}
	return k8sconfig.MakeDynamicClient(cfg.APIConfig)
}
