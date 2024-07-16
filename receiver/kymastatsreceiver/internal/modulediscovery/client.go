package modulediscovery

import (
	"fmt"
	"slices"
	"strings"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

type Client struct {
	discovery    discovery.DiscoveryInterface
	logger       *zap.Logger
	moduleGroups []string
}

func New(discovery discovery.DiscoveryInterface, logger *zap.Logger, moduleGroups []string) *Client {
	return &Client{
		discovery:    discovery,
		logger:       logger,
		moduleGroups: moduleGroups,
	}
}

func (c *Client) Discover() ([]schema.GroupVersionResource, error) {
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

		if !slices.Contains(c.moduleGroups, groupVersion.Group) {
			continue
		}

		c.logger.Debug("Discovered module group", zap.Any("groupVersion", groupVersion))

		for _, resource := range resourceList.APIResources {
			gvr := groupVersion.WithResource(resource.Name)
			if isSubresource(resource.Name) {
				c.logger.Debug("Skipping subresource", zap.Any("groupVersionResource", gvr))
				continue
			}

			gvrs = append(gvrs, gvr)

			c.logger.Debug("Discovered module resource", zap.Any("groupVersionResource", gvr))
		}
	}

	return gvrs, nil
}

func isSubresource(resourceName string) bool {
	return strings.Contains(resourceName, "/")
}
