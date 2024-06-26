package kymastatsreceiver

import (
	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	k8s "k8s.io/client-go/kubernetes"
	"time"
)

// Config represents the receiver config settings within the collector's config.yaml
type Config struct {
	CollectionInterval  time.Duration `mapstructure:"collection_interval"`
	k8sconfig.APIConfig `mapstructure:",squash"`

	makeClient func() (k8s.Interface, error)
}

func (cfg *Config) Validate() error {
	return cfg.APIConfig.Validate()
}

func (cfg *Config) getK8sClient() (k8s.Interface, error) {
	if cfg.makeClient != nil {
		return cfg.makeClient()
	}
	return k8sconfig.MakeClient(cfg.APIConfig)
}
