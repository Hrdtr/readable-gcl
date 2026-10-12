package testconvention_test

import (
	"testing"

	"github.com/hrdtr/readable-gcl/pkg/config"
	"github.com/hrdtr/readable-gcl/pkg/testconvention"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzerPragmatic(t *testing.T) {
	testdataDirectory := analysistest.TestData()
	defaultConfiguration := config.DefaultConfig("pragmatic")
	testConventionAnalyzer := testconvention.NewAnalyzer(defaultConfiguration.TestConvention)

	analysistest.Run(t, testdataDirectory, testConventionAnalyzer, "testsample")
}

func TestAnalyzerStrictConfig(t *testing.T) {
	strictConfiguration := config.DefaultConfig("strict")
	strictAnalyzer := testconvention.NewAnalyzer(strictConfiguration.TestConvention)

	if strictAnalyzer.Name != "testconvention" {
		t.Errorf("expected analyzer name testconvention, got %s", strictAnalyzer.Name)
	}
}

func TestAnalyzerMatchingSourceFile(t *testing.T) {
	testdataDirectory := analysistest.TestData()
	customConfiguration := config.DefaultConfig("pragmatic").TestConvention
	customConfiguration.RequireMatchingSourceFile = true

	analyzer := testconvention.NewAnalyzer(customConfiguration)
	analysistest.Run(t, testdataDirectory, analyzer, "filematching")
}
