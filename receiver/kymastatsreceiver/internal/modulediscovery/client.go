package modulediscovery

import (
	"fmt"
	"slices"
	"strings"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	groupsList, err := c.discovery.ServerGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to discover API groups")
	}

	var moduleGVRs []schema.GroupVersionResource
	for _, apiGroup := range groupsList.Groups {
		if !slices.Contains(c.moduleGroups, apiGroup.Name) {
			continue
		}

		c.logger.Debug("Discovered module group", zap.String("groupVersion", apiGroup.PreferredVersion.GroupVersion))

		gvrs, err := c.discoverGroupResources(apiGroup)
		if err != nil {
			return nil, err
		}

		moduleGVRs = append(moduleGVRs, gvrs...)
	}

	return moduleGVRs, nil
}

func (c *Client) discoverGroupResources(apiGroup metav1.APIGroup) ([]schema.GroupVersionResource, error) {
	rawGroupVersion := apiGroup.PreferredVersion.GroupVersion

	resources, err := c.discovery.ServerResourcesForGroupVersion(rawGroupVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to discover resources for groupVersion %s: %w", rawGroupVersion, err)
	}

	groupVersion, err := schema.ParseGroupVersion(rawGroupVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse groupVersion %s: %w", rawGroupVersion, err)
	}

	var gvrs []schema.GroupVersionResource
	for _, resource := range resources.APIResources {
		gvr := groupVersion.WithResource(resource.Name)
		if isSubresource(resource.Name) {
			c.logger.Debug("Skipping subresource", zap.Any("groupVersionResource", gvr))
			continue
		}

		gvrs = append(gvrs, gvr)

		c.logger.Debug("Discovered module resource", zap.Any("groupVersionResource", gvr))
	}

	return gvrs, nil
}

func isSubresource(resourceName string) bool {
	return strings.Contains(resourceName, "/")
}
