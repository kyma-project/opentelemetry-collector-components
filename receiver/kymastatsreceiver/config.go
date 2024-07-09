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

	ModuleGroups []string `mapstructure:"module_groups"`

	// Used for unit testing only
	makeDynamicClient func() (dynamic.Interface, error)
}

var errEmptyModuleGroups = errors.New("empty module groups")

func (cfg *Config) Validate() error {
	if err := cfg.ControllerConfig.Validate(); err != nil {
		return err
	}

	if err := cfg.APIConfig.Validate(); err != nil {
		return err
	}

	if len(cfg.ModuleGroups) == 0 {
		return errEmptyModuleGroups
	}

	return nil
}

func (cfg *Config) getK8sDynamicClient() (dynamic.Interface, error) {
	if cfg.makeDynamicClient != nil {
		return cfg.makeDynamicClient()
	}
	return k8sconfig.MakeDynamicClient(cfg.APIConfig)
}
