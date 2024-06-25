package kymastatsreceiver

import (
	k8s "k8s.io/client-go/kubernetes"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/k8sconfig"
)

// Config represents the receiver config settings within the collector's config.yaml
type Config struct {
	Interval            string `mapstructure:"collection_interval"`
	k8sconfig.APIConfig `mapstructure:",squash"`

	makeClient func(apiConf k8sconfig.APIConfig) (k8s.Interface, error)
}

func (cfg *Config) Validate() error {
	return cfg.APIConfig.Validate()
}

func (cfg *Config) getK8sClient() (k8s.Interface, error) {
	if cfg.makeClient == nil {
		cfg.makeClient = k8sconfig.GetK8sClient
	}
	return cfg.makeClient(cfg.APIConfig)
}
