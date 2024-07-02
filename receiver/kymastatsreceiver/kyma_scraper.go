package kymastatsreceiver

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

type kymaScraper struct {
	client    dynamic.Interface
	logger    *zap.Logger
	mbs       *metadata.MetricsBuilders
	resources []internal.Resource
}

func newKymaScraper(client dynamic.Interface, set receiver.Settings, resources []internal.Resource, mbc metadata.MetricsBuilderConfig) (scraperhelper.Scraper, error) {
	ks := kymaScraper{
		client: client,
		logger: set.Logger,
		mbs: &metadata.MetricsBuilders{
			KymaTelemetryModuleMetricsBuilder: metadata.NewMetricsBuilder(mbc, set),
		},
		resources: resources,
	}

	return scraperhelper.NewScraper(metadata.Type.String(), ks.scrape)
}

func (scr *kymaScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	summary, err := scr.summary(ctx)
	if err != nil {
		scr.logger.Error("scraping module resources failed", zap.Error(err))
		return pmetric.Metrics{}, err
	}
	mds := internal.MetricsData(scr.mbs, *summary)

	md := pmetric.NewMetrics()
	for i := range mds {
		mds[i].ResourceMetrics().MoveAndAppendTo(md.ResourceMetrics())
	}
	return md, nil
}

func (scr *kymaScraper) summary(ctx context.Context) (*metadata.Stats, error) {
	s := &metadata.Stats{}
	for _, rc := range scr.resources {
		telemetryRes, err := scr.client.Resource(schema.GroupVersionResource{
			Group:    rc.ResourceGroup,
			Version:  rc.ResourceVersion,
			Resource: rc.ResourceName,
		}).List(ctx, metav1.ListOptions{})

		if err != nil {
			scr.logger.Error(fmt.Sprintf("Error fetching module resource %s %s %s", rc.ResourceGroup, rc.ResourceVersion, rc.ResourceName), zap.Error(err))
			return nil, err
		}

		for _, item := range telemetryRes.Items {

			_, ok := item.Object["status"]
			if !ok {
				scr.logger.Error(fmt.Sprintf("Error getting module status for %s %s %s", rc.ResourceGroup, rc.ResourceVersion, rc.ResourceName), zap.Error(err))
				continue
			}

			status := item.Object["status"].(map[string]interface{})

			state, sok := status["state"].(string)

			if !sok {
				scr.logger.Error(fmt.Sprintf("Error getting module status state for %s %s %s", rc.ResourceGroup, rc.ResourceVersion, rc.ResourceName), zap.Error(err))
				continue
			}
			var conditions []metadata.Condition
			r := metadata.ResourceStatusData{
				State:     state,
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
				Module:    rc.ResourceName,
			}

			if condList, cok := status["conditions"].([]interface{}); cok {
				conditions = buildConditions(condList)
			}
			r.Conditions = conditions
			s.Resources = append(s.Resources, r)
		}
	}

	return s, nil
}

func buildConditions(conditionsObj []interface{}) []metadata.Condition {
	var conditions []metadata.Condition
	for _, c := range conditionsObj {
		if cond, ok := c.(map[string]interface{}); ok {
			condition := metadata.Condition{}

			if t, tok := cond["type"].(string); tok {
				condition.Type = t
			}

			if s, sok := cond["status"].(string); sok {
				condition.Status = s
			}

			if r, rok := cond["reason"].(string); rok {
				condition.Reason = r
			}

			conditions = append(conditions, condition)
		}
	}
	return conditions
}
