package kymastatsreceiver

import (
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

	// Used for unit testing only
	makeDynamicClient func() (dynamic.Interface, error)
}

func (cfg *Config) Validate() error {
	err := cfg.ControllerConfig.Validate()
	if err != nil {
		return err
	}
	return cfg.APIConfig.Validate()
}

func (cfg *Config) getK8sDynamicClient() (dynamic.Interface, error) {
	if cfg.makeDynamicClient != nil {
		return cfg.makeDynamicClient()
	}
	return k8sconfig.MakeDynamicClient(cfg.APIConfig)
}
