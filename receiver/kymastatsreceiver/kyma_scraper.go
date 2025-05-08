package kymastatsreceiver

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/k8sleaderelector"
)

type kymaScraper struct {
	config   Config
	dynamic  dynamic.Interface
	logger   *zap.Logger
	mb       *metadata.MetricsBuilder
	isLeader *atomic.Bool
}

type resourceStats struct {
	namespace string
	name      string

	group   string
	version string
	kind    string

	state      string
	conditions []condition

	hasState bool
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
	config Config,
	dynamic dynamic.Interface,
	settings receiver.Settings,
) (scraper.Metrics, error) {
	ks := kymaScraper{
		config:   config,
		dynamic:  dynamic,
		logger:   settings.Logger,
		mb:       metadata.NewMetricsBuilder(config.MetricsBuilderConfig, settings),
		isLeader: &atomic.Bool{},
	}

	return scraper.NewMetrics(ks.scrape, scraper.WithStart(ks.startFunc))
}

func (ks *kymaScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	if !ks.isLeader.Load() {
		return pmetric.Metrics{}, nil
	}
	stats, err := ks.collectResourceStats(ctx)
	if err != nil {
		return pmetric.Metrics{}, err
	}

	now := pcommon.NewTimestampFromTime(time.Now())

	for _, s := range stats {
		if s.hasState {
			ks.mb.RecordKymaResourceStatusStateDataPoint(now, int64(1), s.state)
		}
		rb := ks.mb.NewResourceBuilder()
		if s.namespace != "" {
			rb.SetK8sNamespaceName(s.namespace)
		}

		rb.SetK8sResourceName(s.name)

		rb.SetK8sResourceGroup(s.group)
		rb.SetK8sResourceVersion(s.version)
		rb.SetK8sResourceKind(s.kind)

		for _, c := range s.conditions {
			val := conditionStatusToValue(c.status)
			ks.mb.RecordKymaResourceStatusConditionsDataPoint(now, val, c.reason, c.status, c.condType)
		}
		ks.mb.EmitForResource(metadata.WithResource(rb.Emit()))
	}

	return ks.mb.Emit(), nil
}

func (ks *kymaScraper) startFunc(ctx context.Context, host component.Host) error {
	if ks.config.K8sLeaderElector != nil {
		ks.logger.Info("Starting kymaScraper with leader election")
		extList := host.GetExtensions()
		if extList == nil {
			return errors.New("extension list is empty")
		}

		ext := extList[*ks.config.K8sLeaderElector]
		if ext == nil {
			return errors.New("extension k8s leader elector not found")
		}

		leaderElectorExt, ok := ext.(k8sleaderelector.LeaderElection)
		if !ok {
			return errors.New("referenced extension is not k8s leader elector")
		}
		leaderElectorExt.SetCallBackFuncs(
			func(ctx context.Context) {
				ks.isLeader.Store(true)

			}, func() {
				ks.isLeader.Store(false)
			},
		)
	} else {
		ks.isLeader.Store(true)
		return nil
	}

	return nil
}

func (ks *kymaScraper) collectResourceStats(ctx context.Context) ([]resourceStats, error) {
	var res []resourceStats
	for _, resource := range ks.config.Resources {
		gvr := schema.GroupVersionResource(resource)
		resourceList, err := ks.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
		if err != nil {
			ks.logger.Error("Error fetching resource list",
				zap.Error(err),
				zap.String("group", gvr.Group),
				zap.String("version", gvr.Version),
				zap.String("resource", gvr.Resource))
			return nil, err
		}

		for _, r := range resourceList.Items {
			stats, err := ks.unstructuredToStats(r)
			if err != nil {
				ks.logger.Warn("Error converting unstructured resource to stats",
					zap.Error(err),
					zap.String("name", r.GetName()),
					zap.String("namespace", r.GetNamespace()),
					zap.String("kind", r.GetKind()),
				)
				continue
			}
			stats.group = gvr.Group
			stats.version = gvr.Version
			stats.kind = gvr.Resource

			res = append(res, *stats)
		}
	}

	return res, nil
}

func (ks *kymaScraper) unstructuredToStats(resource unstructured.Unstructured) (*resourceStats, error) {
	status, found, err := unstructured.NestedMap(resource.Object, "status")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, &fieldNotFoundError{"status"}
	}

	state, found, err := unstructured.NestedString(status, "state")
	if err != nil {
		ks.logger.Debug("Error retrieving state: state is not a string",
			zap.Error(err),
			zap.String("name", resource.GetName()),
			zap.String("namespace", resource.GetNamespace()),
			zap.String("kind", resource.GetKind()),
		)
	}
	if !found {
		ks.logger.Debug("Error retrieving state: state not found",
			zap.Error(err),
			zap.String("name", resource.GetName()),
			zap.String("namespace", resource.GetNamespace()),
			zap.String("kind", resource.GetKind()),
		)
	}

	stats := &resourceStats{
		state:     state,
		hasState:  found,
		namespace: resource.GetNamespace(),
		kind:      resource.GetKind(),
		name:      resource.GetName(),
	}

	unstructuredConds, found, err := unstructured.NestedSlice(status, "conditions")
	if err != nil {
		ks.logger.Debug("Error retrieving conditions: conditions are not a slice",
			zap.Error(err),
			zap.String("name", resource.GetName()),
			zap.String("namespace", resource.GetNamespace()),
			zap.String("kind", resource.GetKind()),
		)
		return stats, nil
	}
	if !found {
		ks.logger.Debug("Error retrieving conditions: conditions not found",
			zap.Error(err),
			zap.String("name", resource.GetName()),
			zap.String("namespace", resource.GetNamespace()),
			zap.String("kind", resource.GetKind()),
		)
		return stats, nil
	}

	for _, unstructuredCond := range unstructuredConds {
		cond, err := ks.unstructuredToCondition(unstructuredCond)
		if err != nil {
			ks.logger.Warn("Error converting unstructured resource to stats, condition not supported",
				zap.Error(err),
				zap.String("name", resource.GetName()),
				zap.String("namespace", resource.GetNamespace()),
				zap.String("kind", resource.GetKind()),
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
