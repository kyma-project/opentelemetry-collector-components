package istioenrichmentprocessor

import (
	"context"
	"regexp"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

const (
	kymaModuleAttributeName             = "kyma.module"
	kymaModuleAttributeValue            = "istio"
	istioScopeName                      = "io.kyma-project.telemetry/istio"
	clientAddressAttributeName          = "client.address"
	clientPortAttributeName             = "client.port"
	serverAddressAttributeName          = "server.address"
	networkProtocolNameAttributeName    = "network.protocol.name"
	networkProtocolVersionAttributeName = "network.protocol.version"
	resourceAttributeClusterName        = "cluster_name"
	resourceAttributeLogName            = "log_name"
	resourceAttributeZoneName           = "zone_name"
	resourceAttributeNodeName           = "node_name"
	defaultSeverityText                 = "INFO"
	defaultSeverityNumber               = plog.SeverityNumberInfo
)

var (
	networkProtocolRegex = regexp.MustCompile("^(.+)/(.+$)$")
	networkAddressRegex  = regexp.MustCompile("^(.+):(.+$)$")
)

type istioEnrichmentProcessor struct {
	logger *zap.Logger
	config Config
}

func newIstioEnrichmentProcessor(logger *zap.Logger, cfg Config) *istioEnrichmentProcessor {
	return &istioEnrichmentProcessor{
		logger: logger,
		config: cfg,
	}
}

func (iep *istioEnrichmentProcessor) processLogs(_ context.Context, logs plog.Logs) (plog.Logs, error) {
	resourceLogs := logs.ResourceLogs()
	for r := 0; r < resourceLogs.Len(); r++ {
		updateResourceAttributes := true

		for s := 0; s < resourceLogs.At(r).ScopeLogs().Len(); s++ {
			updateScopeAttributes := true
			scopeLogs := resourceLogs.At(r).ScopeLogs().At(s)

			for i := 0; i < scopeLogs.LogRecords().Len(); i++ {
				logR := scopeLogs.LogRecords().At(i)
				moduleName, exist := logR.Attributes().Get(kymaModuleAttributeName)
				if !exist || moduleName.Str() != kymaModuleAttributeValue {
					// If the log record does not have the kyma.module attribute set to "istio",
					// we skip the enrichment for this log record.
					continue
				}
				enrichSeverityAttributes(logR)
				setNetworkProtocolAttributes(logR)
				setNetworkAddressAttributes(logR)
				logR.Attributes().Remove(kymaModuleAttributeName)

				if updateScopeAttributes {
					iep.setScopeAttributes(scopeLogs)
					updateScopeAttributes = false
				}

				if updateResourceAttributes {
					removeIstioResourceAttributes(resourceLogs.At(r).Resource())
					updateResourceAttributes = false
				}

			}
		}
	}
	return logs, nil
}

func removeIstioResourceAttributes(resource pcommon.Resource) {
	// Remove Istio specific attributes
	resource.Attributes().Remove(resourceAttributeClusterName)
	resource.Attributes().Remove(resourceAttributeLogName)
	resource.Attributes().Remove(resourceAttributeZoneName)
	resource.Attributes().Remove(resourceAttributeNodeName)
}

func enrichSeverityAttributes(logR plog.LogRecord) {
	logR.SetSeverityText(defaultSeverityText)
	logR.SetSeverityNumber(defaultSeverityNumber)
}

func (iep *istioEnrichmentProcessor) setScopeAttributes(scopeLog plog.ScopeLogs) {
	scopeLog.Scope().SetName(istioScopeName)
	scopeLog.Scope().SetVersion(iep.config.ScopeVersion)
}

func setNetworkProtocolAttributes(logR plog.LogRecord) {

	networkProtocol, exist := logR.Attributes().Get(networkProtocolNameAttributeName)
	if exist && networkProtocolRegex.MatchString(networkProtocol.Str()) {
		matches := networkProtocolRegex.FindStringSubmatch(networkProtocol.Str())
		if len(matches) == 3 {
			logR.Attributes().PutStr(networkProtocolNameAttributeName, matches[1])

			logR.Attributes().PutStr(networkProtocolVersionAttributeName, matches[2])
		}
	}
}

func setNetworkAddressAttributes(logR plog.LogRecord) {
	networkProtocolName, existNN := logR.Attributes().Get(networkProtocolNameAttributeName)

	serverAddress, existSA := logR.Attributes().Get(serverAddressAttributeName)
	if existNN && networkProtocolName.Str() != "" && existSA && networkAddressRegex.MatchString(serverAddress.Str()) {
		matches := networkAddressRegex.FindStringSubmatch(serverAddress.Str())
		if len(matches) == 3 {
			logR.Attributes().PutStr(serverAddressAttributeName, matches[1])
		}
	}

	clientAddress, existCA := logR.Attributes().Get(clientAddressAttributeName)

	if existCA && networkAddressRegex.MatchString(clientAddress.Str()) {
		matches := networkAddressRegex.FindStringSubmatch(clientAddress.Str())
		if len(matches) == 3 {
			logR.Attributes().PutStr(clientAddressAttributeName, matches[1])
			logR.Attributes().PutStr(clientPortAttributeName, matches[2])
		}
	}
}
