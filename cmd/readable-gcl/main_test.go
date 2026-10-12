package main

import (
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestMainFunction(t *testing.T) {
	wasCalled := false
	originalMain := multicheckerMain
	defer func() {
		multicheckerMain = originalMain
	}()

	multicheckerMain = func(analyzers ...*analysis.Analyzer) {
		wasCalled = true
		if len(analyzers) != 2 {
			t.Errorf("expected 2 analyzers, got %d", len(analyzers))
		}
	}

	main()

	if !wasCalled {
		t.Error("expected multicheckerMain to be invoked")
	}
}
