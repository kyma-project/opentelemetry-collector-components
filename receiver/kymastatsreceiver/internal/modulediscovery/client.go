package modulediscovery

import (
	"fmt"
	"slices"
	"strings"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

type Config struct {
	ModuleGroups     []string `mapstructure:"module_groups"`
	ExludedResources []string `mapstructure:"excluded_resources"`
}

type Client struct {
	discovery discovery.DiscoveryInterface
	logger    *zap.Logger
	config    Config
}

func New(discovery discovery.DiscoveryInterface, logger *zap.Logger, config Config) *Client {
	return &Client{
		discovery: discovery,
		logger:    logger,
		config:    config,
	}
}

func (c *Client) Discover() ([]schema.GroupVersionResource, error) {
	// ServerPreferredResources returns API resources/groups of the preferred (usually, stored) API version.
	// It guarantees that only version per resource/group is returned.
	resourceLists, err := c.discovery.ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("failed to discover preferred resources: %w", err)
	}

	var gvrs []schema.GroupVersionResource
	for _, resourceList := range resourceLists {
		groupVersion, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to parse groupVersion %s: %w", resourceList.GroupVersion, err)
		}

		if c.shouldSkipGroup(groupVersion.Group) {
			continue
		}

		c.logger.Debug("Discovered module group", zap.Any("groupVersion", groupVersion))

		for _, resource := range resourceList.APIResources {
			gvr := groupVersion.WithResource(resource.Name)

			if c.shouldSkipResource(gvr) {
				continue
			}

			gvrs = append(gvrs, gvr)
			c.logger.Debug("Discovered module resource", zap.Any("groupVersionResource", gvr))
		}
	}

	return gvrs, nil
}

func (c *Client) shouldSkipGroup(group string) bool {
	return !slices.Contains(c.config.ModuleGroups, group)
}

func (c *Client) shouldSkipResource(gvr schema.GroupVersionResource) bool {
	if slices.Contains(c.config.ExludedResources, gvr.Resource) {
		c.logger.Debug("Skipping excluded resource", zap.Any("groupVersionResource", gvr))
		return true
	}

	if isSubresource(gvr.Resource) {
		c.logger.Debug("Skipping subresource", zap.Any("groupVersionResource", gvr))
		return true
	}

	return false
}

func isSubresource(resourceName string) bool {
	return strings.Contains(resourceName, "/")
}
