package leaderelector

import (
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
)

// Config is the configuration struct for your extension.
type Config struct {
	k8sconfig.APIConfig `mapstructure:",squash"`
	LeaseName           string        `mapstructure:"lease_name"`
	LeaseNamespace      string        `mapstructure:"lease_namespace"`
	LeaseDuration       time.Duration `mapstructure:"lease_duration"`
	RenewDuration       time.Duration `mapstructure:"renew_deadline"`
	RetryPeriod         time.Duration `yaml:"retry_period"`
	mu                  sync.Mutex
	makeClient          func(apiConf k8sconfig.APIConfig) (kubernetes.Interface, error)
	// Define any custom fields for your extension's configuration
}

type LeaderElector struct {
}

func (cfg *Config) getK8sClient() (kubernetes.Interface, error) {
	if cfg.makeClient == nil {
		cfg.makeClient = k8sconfig.MakeClient
	}
	return cfg.makeClient(cfg.APIConfig)
}

var _ component.Config = (*Config)(nil)

// Validate checks if the extension configuration is valid
//func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
//	if componentParser == nil {
//		// Nothing to do if there is no config given.
//		return nil
//	}
//
//	if err := componentParser.Unmarshal(cfg, confmap.WithIgnoreUnused()); err != nil {
//		return err
//	}
//
//	lecConfig, err := componentParser.Sub("leaderelector")
//	if err != nil {
//		return fmt.Errorf("unable to extract key %v: %w", subreceiverConfigKey, err)
//	}
//
//	return nil
//}
