package plugin

import (
	"github.com/golangci/plugin-module-register/register"
	"github.com/hrdtr/readable-gcl/pkg/config"
	"github.com/hrdtr/readable-gcl/pkg/namingclarity"
	"github.com/hrdtr/readable-gcl/pkg/testconvention"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("readablegcl", New)
}

// Plugin wraps the configured analysis.Analyzers for execution inside golangci-lint.
type Plugin struct {
	configuration *config.Config
}

var _ register.LinterPlugin = (*Plugin)(nil)

// New initializes the readablegcl linter plugin with settings decoded from golangci-lint.
func New(settings any) (register.LinterPlugin, error) {
	decodedConfiguration, decodeErr := config.Decode(settings)
	if decodeErr != nil {
		return nil, decodeErr
	}
	return &Plugin{configuration: decodedConfiguration}, nil
}

// BuildAnalyzers returns the active analysis.Analyzer instances for golangci-lint.
func (pluginInstance *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	var analyzers []*analysis.Analyzer

	if pluginInstance.configuration.NamingClarity.Enabled {
		analyzers = append(
			analyzers,
			namingclarity.NewAnalyzer(pluginInstance.configuration.NamingClarity),
		)
	}

	if pluginInstance.configuration.TestConvention.Enabled {
		analyzers = append(
			analyzers,
			testconvention.NewAnalyzer(pluginInstance.configuration.TestConvention),
		)
	}

	return analyzers, nil
}

// GetLoadMode declares that readablegcl requires type-checking information (types.Info).
func (pluginInstance *Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
