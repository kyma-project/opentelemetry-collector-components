package kymastatsreceiver

import (
	"context"

	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
	. "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

type kymaScraper struct {
	client dynamic.Interface
	logger *zap.Logger
	mbs    *metadata.MetricsBuilders
}

func newKymaScraper(client dynamic.Interface, mbc metadata.MetricsBuilderConfig) (scraperhelper.Scraper, error) {
	ks := kymaScraper{
		client: client,
		mbs: &metadata.MetricsBuilders{
			TelemetryMetricsBuilder: metadata.NewMetricsBuilder(mbc),
		},
	}

	return scraperhelper.NewScraper(metadata.Type.String(), ks.scrape)
}

func (r *kymaScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {

	mds := MetricsData()

	md := pmetric.NewMetrics()
	for i := range mds {
		mds[i].ResourceMetrics().MoveAndAppendTo(md.ResourceMetrics())
	}
	return md, nil
}

func (r *kymaScraper) summary(ctx context.Context) {

	for _, mbc := range r.mbs.TelemetryMetricsBuilder.Config.KymaTelemetryModuleStat {
		telemetryRes, err := r.client.Resource(schema.GroupVersionResource{
			Group:    mbc.ResourceGroup,
			Version:  mbc.ResourceVersion,
			Resource: mbc.ResourceName,
		}).List(ctx, ListOptions{})

		if err != nil {
			r.logger.Error("Error fetching telemetry resource", zap.Error(err))
			continue
		}
		metadata.NewGaugeMetric(mbc)
		status := telemetryRes.Items[0].Object["status"].(map[string]interface{})

		telemetryModuleData := metadata.ResourceStatusData{
			State:      status["state"].(string),
			Conditions: status["conditions"].([]Condition),
			Name:       telemetryRes.Items[0].GetName(),
			Namespace:  telemetryRes.Items[0].GetNamespace(),
		}
		r.logger.Info(telemetryModuleData.Namespace)
	}
}
