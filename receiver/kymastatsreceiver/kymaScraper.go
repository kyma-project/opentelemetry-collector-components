package kymastatsreceiver

import (
	"context"

	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
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

	return nil, nil
}

func (r *kymaScraper) summary() {

}
