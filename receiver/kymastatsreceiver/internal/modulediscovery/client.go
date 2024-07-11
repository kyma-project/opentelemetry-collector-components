package modulediscovery

import (
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
	var moduleGVRs []schema.GroupVersionResource
	groupsList, err := c.discovery.ServerGroups()
	if err != nil {
		return nil, err
	}

	for _, apiGroup := range groupsList.Groups {
		if !slices.Contains(c.moduleGroups, apiGroup.Name) {
			continue
		}

		groupVersion := apiGroup.PreferredVersion.GroupVersion

		c.logger.Debug("Discovered module group", zap.String("groupVersion", groupVersion))

		resources, err := c.discovery.ServerResourcesForGroupVersion(groupVersion)
		if err != nil {
			return nil, err
		}

		split := strings.Split(groupVersion, "/")
		if len(split) != 2 {
			c.logger.Error("Error splitting groupVersion",
				zap.String("groupVersion", groupVersion))
			continue
		}
		group, version := split[0], split[1]

		for _, resource := range resources.APIResources {
			if strings.Contains(resource.Name, "/") {
				c.logger.Debug("Skipping subresource",
					zap.String("group", group),
					zap.String("version", version),
					zap.String("resource", resource.Name))
				continue
			}

			moduleGVRs = append(moduleGVRs, schema.GroupVersionResource{
				Group:    split[0],
				Version:  split[1],
				Resource: resource.Name,
			})

			c.logger.Debug("Discovered module resource",
				zap.String("group", group),
				zap.String("version", version),
				zap.String("resource", resource.Name))
		}
	}

	return moduleGVRs, nil
}
