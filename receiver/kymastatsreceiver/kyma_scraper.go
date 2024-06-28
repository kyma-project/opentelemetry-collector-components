package kymastatsreceiver

import (
	"context"

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
			scr.logger.Error("Error fetching telemetry resource", zap.Error(err))
			return nil, err
		}

		status := telemetryRes.Items[0].Object["status"].(map[string]interface{})

		r := metadata.ResourceStatusData{
			State:      status["state"].(string),
			Conditions: status["conditions"].([]Condition),
			Name:       telemetryRes.Items[0].GetName(),
			Namespace:  telemetryRes.Items[0].GetNamespace(),
		}

		s.Resources = append(s.Resources, r)
	}

	return s, nil
}
