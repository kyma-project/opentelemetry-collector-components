package kymastatsreceiver

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

type kymaScraper struct {
	discovery    discovery.DiscoveryInterface
	dynamic      dynamic.Interface
	logger       *zap.Logger
	mb           *metadata.MetricsBuilder
	moduleGroups []string
}

type moduleStats struct {
	namespace string
	kind      string

	state      string
	conditions []condition
}

type condition struct {
	condType string
	status   string
	reason   string
}

type fieldNotFoundError struct {
	field string
}

func (e *fieldNotFoundError) Error() string {
	return fmt.Sprintf("field not found: %s", e.field)
}

func newKymaScraper(
	discovery discovery.DiscoveryInterface,
	dynamic dynamic.Interface,
	settings receiver.Settings,
	config *Config,
) (scraperhelper.Scraper, error) {
	ks := kymaScraper{
		discovery:    discovery,
		dynamic:      dynamic,
		logger:       settings.Logger,
		mb:           metadata.NewMetricsBuilder(config.MetricsBuilderConfig, settings),
		moduleGroups: config.ModuleGroups,
	}

	return scraperhelper.NewScraper(metadata.Type.String(), ks.scrape)
}

func (ks *kymaScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	moduleGVRs, err := ks.discoverModules()
	if err != nil {
		return pmetric.Metrics{}, err
	}

	stats, err := ks.collectModuleStats(ctx, moduleGVRs)
	if err != nil {
		return pmetric.Metrics{}, err
	}

	now := pcommon.NewTimestampFromTime(time.Now())

	for _, s := range stats {
		ks.mb.RecordKymaModuleStatusStateDataPoint(now, int64(1), s.state)
		rb := ks.mb.NewResourceBuilder()
		rb.SetK8sNamespaceName(s.namespace)
		rb.SetKymaModuleName(s.kind)
		for _, c := range s.conditions {
			val := conditionStatusToValue(c.status)
			ks.mb.RecordKymaModuleStatusConditionsDataPoint(now, val, c.reason, c.status, c.condType)
		}
		ks.mb.EmitForResource(metadata.WithResource(rb.Emit()))
	}

	return ks.mb.Emit(), nil
}

func (ks *kymaScraper) discoverModules() ([]schema.GroupVersionResource, error) {
	var moduleGVRs []schema.GroupVersionResource
	groupsList, err := ks.discovery.ServerGroups()
	if err != nil {
		return nil, err
	}

	for _, group := range groupsList.Groups {
		if !slices.Contains(ks.moduleGroups, group.Name) {
			continue
		}

		ks.logger.Debug("Discovered module group", zap.String("groupVersion", group.PreferredVersion.GroupVersion))

		resources, err := ks.discovery.ServerResourcesForGroupVersion(group.PreferredVersion.GroupVersion)
		if err != nil {
			return nil, err
		}

		for _, resource := range resources.APIResources {
			moduleGVRs = append(moduleGVRs, schema.GroupVersionResource{
				Group:    resource.Group,
				Version:  resource.Version,
				Resource: resource.Name,
			})

			ks.logger.Debug("Discovered module resource",
				zap.String("group", resource.Group),
				zap.String("version", resource.Version),
				zap.String("resource", resource.Name))
		}
	}

	return moduleGVRs, nil
}

func (ks *kymaScraper) collectModuleStats(ctx context.Context, moduleGVRs []schema.GroupVersionResource) ([]moduleStats, error) {
	var res []moduleStats
	for _, gvr := range moduleGVRs {
		moduleList, err := ks.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
		if err != nil {
			ks.logger.Error("Error fetching module list",
				zap.Error(err),
				zap.String("group", gvr.Group),
				zap.String("version", gvr.Version),
				zap.String("resource", gvr.Resource))
			return nil, err
		}

		for _, module := range moduleList.Items {
			stats, err := ks.unstructuredToStats(module)
			if err != nil {
				ks.logger.Error("Error converting unstructured module to stats",
					zap.Error(err),
					zap.String("name", module.GetName()),
					zap.String("namespace", module.GetNamespace()),
					zap.String("kind", module.GetKind()),
				)
				continue
			}

			res = append(res, *stats)
			// Take only the first valid module custom resource
			break
		}
	}

	return res, nil
}

func (ks *kymaScraper) unstructuredToStats(module unstructured.Unstructured) (*moduleStats, error) {
	status, found, err := unstructured.NestedMap(module.Object, "status")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, &fieldNotFoundError{"status"}
	}

	state, found, err := unstructured.NestedString(status, "state")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, &fieldNotFoundError{"state"}
	}

	stats := &moduleStats{
		state:     state,
		namespace: module.GetNamespace(),
		kind:      module.GetKind(),
	}

	unstructuredConds, found, err := unstructured.NestedSlice(status, "conditions")
	if err != nil {
		ks.logger.Warn("Error retrieving conditions: conditions are not a slice",
			zap.Error(err),
			zap.String("name", module.GetName()),
			zap.String("namespace", module.GetNamespace()),
			zap.String("kind", module.GetKind()),
		)
		return stats, nil
	}
	if !found {
		ks.logger.Warn("Error retrieving conditions: conditions not found",
			zap.Error(err),
			zap.String("name", module.GetName()),
			zap.String("namespace", module.GetNamespace()),
			zap.String("kind", module.GetKind()),
		)
		return stats, nil
	}

	for _, unstructuredCond := range unstructuredConds {
		cond, err := ks.unstructuredToCondition(unstructuredCond)
		if err != nil {
			ks.logger.Error("Error converting unstructured module to stats, condition not supported",
				zap.Error(err),
				zap.String("name", module.GetName()),
				zap.String("namespace", module.GetNamespace()),
				zap.String("kind", module.GetKind()),
			)
			continue
		}
		stats.conditions = append(stats.conditions, *cond)
	}

	return stats, nil
}

func (ks *kymaScraper) unstructuredToCondition(cond any) (*condition, error) {
	condMap, ok := cond.(map[string]any)
	if !ok {
		return nil, errors.New("condition is not a map")
	}

	condType, found, err := unstructured.NestedString(condMap, "type")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, &fieldNotFoundError{"type"}
	}

	status, found, err := unstructured.NestedString(condMap, "status")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, &fieldNotFoundError{"status"}
	}

	reason, found, err := unstructured.NestedString(condMap, "reason")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, &fieldNotFoundError{"reason"}
	}

	return &condition{
		condType: condType,
		status:   status,
		reason:   reason,
	}, nil
}

func conditionStatusToValue(status string) int64 {
	switch status {
	case string(metav1.ConditionTrue):
		return 1
	case string(metav1.ConditionFalse):
		return 0
	default:
		return -1
	}
}
