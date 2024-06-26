package singletonreceivercreator

import (
	"fmt"
	"time"

	k8s "k8s.io/client-go/kubernetes"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

const (
	// receiversConfigKey is the config key name used to specify the subreceivers.
	subreceiverConfigKey    = "receiver"
	leaderElectionConfigKey = "leader_election"
)

// Config defines configuration for leader receiver creator.
type Config struct {
	k8sconfig.APIConfig  `mapstructure:",squash"`
	leaderElectionConfig leaderElectionConfig `yaml:"leader_election"`
	subreceiverConfig    receiverConfig

	makeClient func() (k8s.Interface, error)
}

type leaderElectionConfig struct {
	leaseName            string        `mapstructure:"lease_name"`
	leaseNamespace       string        `mapstructure:"lease_namespace"`
	leaseDurationSeconds time.Duration `mapstructure:"lease_duration"`
	renewDeadlineSeconds time.Duration `mapstructure:"renew_deadline"`
	retryPeriodSeconds   time.Duration `mapstructure:"retry_period"`
}

// receiverConfig describes a receiver instance with a default config.
type receiverConfig struct {
	// id is the id of the subreceiver (ie <receiver type>/<id>).
	id component.ID
	// config is the map configured by the user in the config file. It is the contents of the map from
	// the "config" section. The keys and values are arbitrarily configured by the user.
	config map[string]any
}

// and its arbitrary config map values.
func newReceiverConfig(name string, cfg map[string]any) (receiverConfig, error) {
	id := component.ID{}
	if err := id.UnmarshalText([]byte(name)); err != nil {
		return receiverConfig{}, fmt.Errorf("failed to parse subreceiver id %v: %w", name, err)
	}

	return receiverConfig{
		id:     id,
		config: cfg,
	}, nil
}

func newLeaderElectionConfig(lec leaderElectionConfig, cfg map[string]any) (leaderElectionConfig, error) {
	if leaseName, ok := cfg["lease_name"].(string); ok {
		lec.leaseName = leaseName
	}

	if leaseNamespace, ok := cfg["lease_namespace"].(string); ok {
		lec.leaseNamespace = leaseNamespace
	}

	if leaseDuration, ok := cfg["lease_duration"].(string); ok {
		leasedurationSec, err := time.ParseDuration(leaseDuration)
		if err != nil {
			return leaderElectionConfig{}, fmt.Errorf("failed to parse lease duration: %w", err)
		}
		lec.leaseDurationSeconds = leasedurationSec
	}

	if renewDeadline, ok := cfg["renew_deadline"].(string); ok {
		renewDeadlineSec, err := time.ParseDuration(renewDeadline)
		if err != nil {
			return leaderElectionConfig{}, fmt.Errorf("failed to parse renew deadline: %w", err)
		}
		lec.renewDeadlineSeconds = renewDeadlineSec
	}

	if retryPeriod, ok := cfg["retry_period"].(string); ok {
		retryPeriodSec, err := time.ParseDuration(retryPeriod)
		if err != nil {
			return leaderElectionConfig{}, fmt.Errorf("failed to parse retry period: %w", err)
		}
		lec.retryPeriodSeconds = retryPeriodSec
	}

	return lec, nil
}

var _ confmap.Unmarshaler = (*Config)(nil)

func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	if componentParser == nil {
		// Nothing to do if there is no config given.
		return nil
	}

	if err := componentParser.Unmarshal(cfg, confmap.WithIgnoreUnused()); err != nil {
		return err
	}

	subreceiverConfig, err := componentParser.Sub(subreceiverConfigKey)
	if err != nil {
		return fmt.Errorf("unable to extract key %v: %w", subreceiverConfigKey, err)
	}

	lec, err := componentParser.Sub(leaderElectionConfigKey)
	if err != nil {
		return fmt.Errorf("unable to extract key %v: %w", leaderElectionConfigKey, err)
	}

	cfg.leaderElectionConfig, err = newLeaderElectionConfig(cfg.leaderElectionConfig, lec.ToStringMap())
	if err != nil {
		return fmt.Errorf("failed to create leader election config: %w", err)
	}

	for subreceiverKey := range subreceiverConfig.ToStringMap() {
		receiverConfig, err := subreceiverConfig.Sub(subreceiverKey)
		if err != nil {
			return fmt.Errorf("unable to extract subreceiver key %v: %w", subreceiverKey, err)
		}

		cfg.subreceiverConfig, err = newReceiverConfig(subreceiverKey, receiverConfig.ToStringMap())
		if err != nil {
			return fmt.Errorf("failed to create subreceiver config: %w", err)
		}

		return nil
	}

	return nil
}

func (cfg *Config) Validate() error {
	return cfg.APIConfig.Validate()
}

func (cfg *Config) getK8sClient() (k8s.Interface, error) {
	if cfg.makeClient != nil {
		return cfg.makeClient()
	}
	fmt.Printf("cfg.APIConfig: %v\n", cfg.APIConfig.AuthType)
	return k8sconfig.MakeClient(cfg.APIConfig)
}
