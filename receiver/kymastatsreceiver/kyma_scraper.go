package kymastatsreceiver

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"errors"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type kymaScraper struct {
	client     dynamic.Interface
	logger     *zap.Logger
	mb         *metadata.MetricsBuilder
	moduleGVRs []ModuleResourceConfig
}

type moduleStats struct {
	name      string
	namespace string
	resource  string

	state      string
	conditions []condition
}

type condition struct {
	condType string
	status   string
	reason   string
}

type errFieldNotFound struct {
	field string
}

func (e *errFieldNotFound) Error() string {
	return fmt.Sprintf("field not found: %s", e.field)
}

func newKymaScraper(client dynamic.Interface, settings receiver.Settings, resources []ModuleResourceConfig, mbc metadata.MetricsBuilderConfig) (scraperhelper.Scraper, error) {
	ks := kymaScraper{
		client:     client,
		logger:     settings.Logger,
		mb:         metadata.NewMetricsBuilder(mbc, settings),
		moduleGVRs: resources,
	}

	return scraperhelper.NewScraper(metadata.Type.String(), ks.scrape)
}

func (ks *kymaScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	stats, err := ks.collectModuleStats(ctx)
	if err != nil {
		return pmetric.Metrics{}, err
	}

	now := pcommon.NewTimestampFromTime(time.Now())

	for _, s := range stats {
		ks.mb.RecordKymaModuleStatusStateDataPoint(now, int64(1), s.state, s.resource)
		rb := ks.mb.NewResourceBuilder()
		rb.SetK8sNamespaceName(s.namespace)
		rb.SetKymaModuleName(s.name)
		for _, c := range s.conditions {
			value := -1
			switch c.status {
			case string(metav1.ConditionTrue):
				value = 1
			case string(metav1.ConditionFalse):
				value = 0
			}
			ks.mb.RecordKymaModuleStatusConditionsDataPoint(now, int64(value), s.resource, c.reason, c.status, c.condType)
		}
		ks.mb.EmitForResource(metadata.WithResource(rb.Emit()))
	}

	return ks.mb.Emit(), nil
}

func (ks *kymaScraper) collectModuleStats(ctx context.Context) ([]moduleStats, error) {
	var res []moduleStats
	for _, gvr := range ks.moduleGVRs {
		moduleList, err := ks.client.Resource(schema.GroupVersionResource{
			Group:    gvr.ResourceGroup,
			Version:  gvr.ResourceVersion,
			Resource: gvr.ResourceName,
		}).List(ctx, metav1.ListOptions{})
		if err != nil {
			ks.logger.Error("Error fetching module list",
				zap.Error(err),
				zap.String("group", gvr.ResourceGroup),
				zap.String("version", gvr.ResourceVersion),
				zap.String("resource", gvr.ResourceName))
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
		return nil, &errFieldNotFound{"status"}
	}

	state, found, err := unstructured.NestedString(status, "state")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, &errFieldNotFound{"state"}
	}

	unstructuredConds, found, err := unstructured.NestedSlice(status, "conditions")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, &errFieldNotFound{"conditions"}
	}

	stats := &moduleStats{
		state:     state,
		name:      module.GetName(),
		namespace: module.GetNamespace(),
		resource:  module.GetKind(),
	}

	for _, unstructuredCond := range unstructuredConds {
		cond, err := ks.unstructuredToCondition(unstructuredCond)
		if err != nil {
			return nil, err
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
		return nil, &errFieldNotFound{"type"}
	}

	status, found, err := unstructured.NestedString(condMap, "status")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, &errFieldNotFound{"status"}
	}

	reason, found, err := unstructured.NestedString(condMap, "reason")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, &errFieldNotFound{"reason"}
	}

	return &condition{
		condType: condType,
		status:   status,
		reason:   reason,
	}, nil
}
