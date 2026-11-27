package namingclarity

import (
	"github.com/hrdtr/readable-gcl/pkg/config"
	"golang.org/x/tools/go/analysis"
)

const (
	AnalyzerName = "namingclarity"
	AnalyzerDoc  = "variable, argument, and identifier naming must be clear, readable, and adhere to repository conventions"
)

// Analyzer is the default namingclarity analyzer instantiated with pragmatic preset defaults.
var Analyzer = NewAnalyzer(config.DefaultConfig("pragmatic").NamingClarity)

// NewAnalyzer creates a configured analysis.Analyzer using the supplied NamingClarityConfig.
func NewAnalyzer(configuration config.NamingClarityConfig) *analysis.Analyzer {
	if configuration.BadInitialismsList != nil {
		configuration.BadInitialisms = make(map[string]string, len(configuration.BadInitialismsList))
		for _, rule := range configuration.BadInitialismsList {
			configuration.BadInitialisms[rule.Word] = rule.Replacement
		}
	}
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
