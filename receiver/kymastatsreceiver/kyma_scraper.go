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

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
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
		ks.logger.Error("scraping module resources failed", zap.Error(err))
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
	for _, rc := range ks.moduleGVRs {
		telemetryRes, err := ks.client.Resource(schema.GroupVersionResource{
			Group:    rc.ResourceGroup,
			Version:  rc.ResourceVersion,
			Resource: rc.ResourceName,
		}).List(ctx, metav1.ListOptions{})

		if err != nil {
			ks.logger.Error(fmt.Sprintf("Error fetching module resource %s %s %s", rc.ResourceGroup, rc.ResourceVersion, rc.ResourceName), zap.Error(err))
			return nil, err
		}

		for _, item := range telemetryRes.Items {
			_, ok := item.Object["status"]
			if !ok {
				ks.logger.Error(fmt.Sprintf("Error getting module status for %s %s %s", rc.ResourceGroup, rc.ResourceVersion, rc.ResourceName), zap.Error(err))
				continue
			}

			status, sok := item.Object["status"].(map[string]interface{})
			if !sok {
				ks.logger.Error(fmt.Sprintf("Error getting module status type for %s %s %s", rc.ResourceGroup, rc.ResourceVersion, rc.ResourceName), zap.Error(err))
				continue
			}
			state, sok := status["state"].(string)

			if !sok {
				ks.logger.Error(fmt.Sprintf("Error getting module status state for %s %s %s", rc.ResourceGroup, rc.ResourceVersion, rc.ResourceName), zap.Error(err))
				continue
			}
			var conditions []condition
			stats := moduleStats{
				state:     state,
				name:      item.GetName(),
				namespace: item.GetNamespace(),
				resource:  item.GetKind(),
			}

			if condList, cok := status["conditions"].([]interface{}); cok {
				conditions = buildConditions(condList)
			}
			stats.conditions = conditions
			res = append(res, stats)
		}
	}

	return res, nil
}

func buildConditions(conditionsObj []interface{}) []condition {
	var conditions []condition
	for _, c := range conditionsObj {
		if cond, ok := c.(map[string]interface{}); ok {
			condItem := condition{}

			if t, tok := cond["type"].(string); tok {
				condItem.condType = t
			}

			if s, sok := cond["status"].(string); sok {
				condItem.status = s
			}

			if r, rok := cond["reason"].(string); rok {
				condItem.reason = r
			}

			conditions = append(conditions, condItem)
		}
	}
	return conditions
}
