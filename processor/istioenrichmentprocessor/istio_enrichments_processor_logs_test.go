package istioenrichmentprocessor

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/kyma-project/opentelemetry-collector-components/processor/istioenrichmentprocessor/internal/metadata"
)

func TestProcessLogs(t *testing.T) {
	tt := []struct {
		name                               string
		logs                               plog.Logs
		expectedScopeName                  string
		expectedScopeVersion               string
		expectedNetworkProtocolName        string
		expectedNetworkProtocolVersion     string
		expectedServerAddress              string
		expectedClientAddress              string
		expectedClientPort                 string
		expectedSeverityText               string
		expectedSeverityNumber             plog.SeverityNumber
		expectedResourceAttributeCount     int
		expectShouldTestModifiedAttributes bool
	}{
		{
			name: "logs with kyma.module istio attribute",
			logs: NewPLogBuilder().WithScopeName("test_scope").
				WithScopeVersion("someVersion").
				WithResourceAttributes(map[string]string{
					resourceAttributeClusterName: "clusterName",
					resourceAttributeLogName:     "logName",
					resourceAttributeNodeName:    "nodeName",
					resourceAttributeZoneName:    "zoneName",
				}).
				WithLogAttributes(map[string]string{
					"kyma.module":           "istio",
					"network.protocol.name": "HTTP/1.0",
					"server.address":        "server.local:123",
					"client.address":        "client.local:456",
				}).
				Build(),
			expectedScopeName:                  istioScopeName,
			expectedScopeVersion:               "v1",
			expectedNetworkProtocolName:        "HTTP",
			expectedNetworkProtocolVersion:     "1.0",
			expectedServerAddress:              "server.local",
			expectedClientAddress:              "client.local",
			expectedClientPort:                 "456",
			expectedSeverityText:               "INFO",
			expectedSeverityNumber:             plog.SeverityNumberInfo,
			expectedResourceAttributeCount:     0,
			expectShouldTestModifiedAttributes: true,
		},
		{
			name: "logs without kyma.module istio attribute",
			logs: NewPLogBuilder().WithScopeName("test_scope").
				WithScopeVersion("someVersion").
				WithResourceAttributes(map[string]string{
					resourceAttributeClusterName: "clusterName",
					resourceAttributeLogName:     "logName",
					resourceAttributeNodeName:    "nodeName",
					resourceAttributeZoneName:    "zoneName",
				}).
				WithLogAttributes(map[string]string{
					"kyma.module":           "some_other_module",
					"network.protocol.name": "HTTP/1.0",
					"server.address":        "server.local:123",
					"client.address":        "client.local:456",
				}).
				Build(),
			expectedScopeName:                  "test_scope",
			expectedScopeVersion:               "someVersion",
			expectedNetworkProtocolName:        "HTTP/1.0",
			expectedServerAddress:              "server.local:123",
			expectedClientAddress:              "client.local:456",
			expectedSeverityText:               "",
			expectedSeverityNumber:             plog.SeverityNumberUnspecified,
			expectedResourceAttributeCount:     4,
			expectShouldTestModifiedAttributes: false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sink := new(consumertest.LogsSink)
			config := Config{
				ScopeVersion: "v1",
			}

			factory := NewFactory()
			cm, err := factory.CreateLogs(
				t.Context(),
				processortest.NewNopSettings(metadata.Type),
				config,
				sink,
			)
			require.NoError(t, err)
			require.NotNil(t, cm)

			err = cm.Start(t.Context(), componenttest.NewNopHost())
			require.NoError(t, err)

			cErr := cm.ConsumeLogs(t.Context(), tc.logs)
			require.NoError(t, cErr)

			allLogs := sink.AllLogs()
			require.Len(t, allLogs, 1)
			for _, l := range allLogs {
				for iResource := 0; iResource < l.ResourceLogs().Len(); iResource++ {

					require.Equal(t, tc.expectedResourceAttributeCount, l.ResourceLogs().At(iResource).Resource().Attributes().Len())

					for iScope := 0; iScope < l.ResourceLogs().At(iResource).ScopeLogs().Len(); iScope++ {

						require.Equal(t, tc.expectedScopeVersion, l.ResourceLogs().At(iResource).ScopeLogs().At(iScope).Scope().Version())
						require.Equal(t, tc.expectedScopeName, l.ResourceLogs().At(iResource).ScopeLogs().At(iScope).Scope().Name())

						for iLog := 0; iLog < l.ResourceLogs().At(iResource).ScopeLogs().At(iScope).LogRecords().Len(); iLog++ {
							logR := l.ResourceLogs().At(iResource).ScopeLogs().At(iScope).LogRecords().At(iLog)
							require.Equal(t, tc.expectedNetworkProtocolName, logR.Attributes().AsRaw()["network.protocol.name"])
							require.Equal(t, tc.expectedServerAddress, logR.Attributes().AsRaw()["server.address"])
							require.Equal(t, tc.expectedClientAddress, logR.Attributes().AsRaw()["client.address"])
							require.Equal(t, tc.expectedSeverityNumber, logR.SeverityNumber())
							require.Equal(t, tc.expectedSeverityText, logR.SeverityText())

							if tc.expectShouldTestModifiedAttributes {
								require.Equal(t, tc.expectedNetworkProtocolVersion, logR.Attributes().AsRaw()["network.protocol.version"])
								require.Equal(t, tc.expectedClientPort, logR.Attributes().AsRaw()["client.port"])
							}
						}
					}
				}
			}
		})
	}
}
