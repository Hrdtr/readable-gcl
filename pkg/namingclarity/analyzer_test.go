package namingclarity_test

import (
	"testing"

	"github.com/hrdtr/readable-gcl/pkg/config"
	"github.com/hrdtr/readable-gcl/pkg/namingclarity"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdataDirectory := analysistest.TestData()
	defaultConfiguration := config.DefaultConfig("pragmatic")
	namingAnalyzer := namingclarity.NewAnalyzer(defaultConfiguration.NamingClarity)

	analysistest.Run(t, testdataDirectory, namingAnalyzer, "sample")
}

func TestAnalyzerDisabled(t *testing.T) {
	disabledAnalyzer := namingclarity.NewAnalyzer(config.NamingClarityConfig{Enabled: false})
	result, err := disabledAnalyzer.Run(nil)
	if err != nil || result != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", result, err)
	}
}
