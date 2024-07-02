package internal

import (
	"time"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

func MetricsData(mbs *metadata.MetricsBuilders, s metadata.Stats) []pmetric.Metrics {

	acc := &metricDataAccumulator{
		time: time.Now(),
		mbs:  mbs,
	}
	for _, r := range s.Resources {
		acc.resourceStats(r)
		for _, c := range r.Conditions {
			acc.resourceConditionStats(r.Name, r.Module, r.Namespace, c)
		}
	}
	return acc.m
}
