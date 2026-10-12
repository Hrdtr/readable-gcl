package testconvention

import (
	"github.com/hrdtr/readable-gcl/pkg/config"
	"golang.org/x/tools/go/analysis"
)

const (
	AnalyzerName = "testconvention"
	AnalyzerDoc  = "test function naming conventions and table-driven subtest execution requirements"
)

// Analyzer is the default testconvention analyzer instantiated with pragmatic preset defaults.
var Analyzer = NewAnalyzer(config.DefaultConfig("pragmatic").TestConvention)

// NewAnalyzer creates a configured analysis.Analyzer using the supplied TestConventionConfig.
func NewAnalyzer(configuration config.TestConventionConfig) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: AnalyzerName,
		Doc:  AnalyzerDoc,
		Run: func(pass *analysis.Pass) (any, error) {
			if !configuration.Enabled {
				return nil, nil
			}
			return run(pass, configuration)
		},
	}
}
