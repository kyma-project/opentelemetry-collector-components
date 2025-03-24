package serviceenrichmentprocessor

import (
	"context"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
	"testing"
)

func TestProcessLogs(t *testing.T) {
	tt := []struct {
		name                string
		logs                plog.Logs
		expectedServiceName string
	}{
		{
			name: "logs with service name not set",
			logs: setLogs(map[string]string{
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "logs with service name set to unknown_service",
			logs: setLogs(map[string]string{
				"service.name":                "unknown_service",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "logs with service name not set and deployment name set",
			logs: setLogs(map[string]string{
				"k8s.deployment.name": "foo-deployment-name",
			}),
			expectedServiceName: "foo-deployment-name",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			logger := zap.NewNop()
			config := &Config{
				CustomLabels: []string{
					"kyma.kubernetes_io_app_name",
					"kyma.app_name",
				},
			}
			sep := newServiceEnrichmentProcessor(logger, config)
			res, err := sep.processLogs(context.TODO(), tc.logs)
			require.NoError(t, err)
			for i := 0; i < res.ResourceLogs().Len(); i++ {
				attr := res.ResourceLogs().At(i).Resource().Attributes()
				svcName, ok := attr.Get("service.name")
				require.True(t, ok)
				require.Equal(t, tc.expectedServiceName, svcName.AsString())
			}
		})
	}
}

func setLogs(attrs ...map[string]string) plog.Logs {
	logs := plog.NewLogs()
	for _, attr := range attrs {
		resLogs := logs.ResourceLogs().AppendEmpty()
		for k, v := range attr {
			resLogs.Resource().Attributes().PutStr(k, v)
		}
	}
	return logs
}
