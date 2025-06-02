package istioenrichmentprocessor

import (
	"context"
	"regexp"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

const (
	kymaModuleAttributeName             = "kyma.module"
	kymaModuleAttributeValue            = "istio"
	istioScopeName                      = "io.kyma-project.telemetry/istio"
	clientAddressAttributeName          = "client.address"
	clientPortAttributeName             = "client.port"
	networkProtocolNameAttributeName    = "network.protocol.name"
	networkProtocolVersionAttributeName = "network.protocol.version"
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
			}
		}
	}
	return logs, nil
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

	clientAddress, existCA := logR.Attributes().Get(clientAddressAttributeName)

	if existCA && networkAddressRegex.MatchString(clientAddress.Str()) {
		matches := networkAddressRegex.FindStringSubmatch(clientAddress.Str())
		if len(matches) == 3 {
			logR.Attributes().PutStr(clientAddressAttributeName, matches[1])
			logR.Attributes().PutStr(clientPortAttributeName, matches[2])
		}
	}
}
