package kymastatsreceiver

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/receiver"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal"

	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
	. "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

type kymaScraper struct {
	client    dynamic.Interface
	logger    *zap.Logger
	mbs       *metadata.MetricsBuilders
	resources []internal.Resource
}

func newKymaScraper(client dynamic.Interface, set receiver.CreateSettings, resources []internal.Resource, mbc metadata.MetricsBuilderConfig) (scraperhelper.Scraper, error) {
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
		}).List(ctx, ListOptions{})

		if err != nil {
			scr.logger.Error(fmt.Sprintf("Error fetching module resource %s %s %s", rc.ResourceGroup, rc.ResourceVersion, rc.ResourceName), zap.Error(err))
			return nil, err
		}

		for _, item := range telemetryRes.Items {

			status := item.Object["status"].(map[string]interface{})
			var conditions []metadata.Condition

			r := metadata.ResourceStatusData{
				State:      status["state"].(string),
				Name:       telemetryRes.Items[0].GetName(),
				Namespace:  telemetryRes.Items[0].GetNamespace(),
				ModuleName: rc.ResourceName,
			}
			if status["conditions"] != nil {
				for _, c := range status["conditions"].([]interface{}) {
					condition := c.(map[string]interface{})
					conditions = append(conditions, metadata.Condition{
						Type:   condition["type"].(string),
						Status: condition["status"].(string),
						Reason: condition["reason"].(string),
					})
				}
			}
			r.Conditions = conditions
			s.Resources = append(s.Resources, r)
		}
	}

	return s, nil
}
