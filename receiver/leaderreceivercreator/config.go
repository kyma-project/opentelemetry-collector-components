// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package leaderreceivercreator

import (
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

const (
	// receiversConfigKey is the config key name used to specify the subreceivers.
	subreceiverConfigKey    = "receiver"
	leaderElectionConfigKey = "leader_election"
)

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

func newLeaderElectionConfig(lecConfig leaderElectionConfig, cfg map[string]any) (leaderElectionConfig, error) {
	if leaseName, ok := cfg["lease_name"].(string); ok {
		lecConfig.leaseName = leaseName
	}
	if leaseNamespace, ok := cfg["lease_namespace"].(string); ok {
		lecConfig.leaseNamespace = leaseNamespace
	}

	if leaseDuration, ok := cfg["lease_duration"].(string); ok {
		fmt.Printf("leaseDuration: %v\n", leaseDuration)
		leasedurationSec, err := time.ParseDuration(leaseDuration)
		if err != nil {
			return leaderElectionConfig{}, fmt.Errorf("failed to parse lease duration: %w", err)
		}
		lecConfig.leaseDurationSeconds = leasedurationSec
	}
	if renewDeadline, ok := cfg["renew_deadline"].(string); ok {
		renewDeadlineSec, err := time.ParseDuration(renewDeadline)
		if err != nil {
			return leaderElectionConfig{}, fmt.Errorf("failed to parse renew deadline: %w", err)
		}
		lecConfig.renewDeadlineSeconds = renewDeadlineSec
	}
	if retryPeriod, ok := cfg["retry_period"].(string); ok {
		retryPeriodSec, err := time.ParseDuration(retryPeriod)
		if err != nil {
			return leaderElectionConfig{}, fmt.Errorf("failed to parse retry period: %w", err)
		}
		lecConfig.retryPeriodSeconds = retryPeriodSec
	}

	return lecConfig, nil
}

var _ confmap.Unmarshaler = (*Config)(nil)

// Config defines configuration for receiver_creator.
type Config struct {
	leaderElectionConfig leaderElectionConfig `yaml:"leader_election"`
	subreceiverConfig    receiverConfig
}

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
