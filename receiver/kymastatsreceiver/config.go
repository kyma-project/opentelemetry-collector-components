package kymastatsreceiver

import (
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

// Config represents the receiver config settings within the collector's config.yaml
type Config struct {
	scraperhelper.ControllerConfig `mapstructure:",squash"`
	k8sconfig.APIConfig            `mapstructure:",squash"`
	makeDynamicClient              func() (dynamic.Interface, error)
	metadata.MetricsBuilderConfig  `mapstructure:",squash"`
	Resources                      []Resource `mapstructure:"kyma_module_resources"`
}

type Resource struct {
	ResourceGroup   string `mapstructure:"resource_group"`
	ResourceName    string `mapstructure:"resource_name"`
	ResourceVersion string `mapstructure:"resource_version"`
}

func (cfg *Config) Validate() error {
	err := cfg.ControllerConfig.Validate()
	if err != nil {
		return err
	}
	return cfg.APIConfig.Validate()
}

func (cfg *Config) GetK8sDynamicClient() (dynamic.Interface, error) {
	if cfg.makeDynamicClient != nil {
		return cfg.makeDynamicClient()
	}
	return k8sconfig.MakeDynamicClient(cfg.APIConfig)
}
