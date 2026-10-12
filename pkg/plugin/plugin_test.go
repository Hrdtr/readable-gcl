package plugin_test

import (
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"github.com/hrdtr/readable-gcl/pkg/plugin"
)

func TestNewPlugin(t *testing.T) {
	pluginInstance, err := plugin.New(map[string]any{
		"preset": "pragmatic",
	})
	if err != nil {
		t.Fatalf("unexpected error creating plugin: %v", err)
	}

	analyzers, buildErr := pluginInstance.BuildAnalyzers()
	if buildErr != nil {
		t.Fatalf("unexpected error building analyzers: %v", buildErr)
	}
	if len(analyzers) != 2 {
		t.Errorf("expected 2 analyzers, got %d", len(analyzers))
	}

	loadMode := pluginInstance.GetLoadMode()
	if loadMode != register.LoadModeTypesInfo {
		t.Errorf("expected load mode %q, got %q", register.LoadModeTypesInfo, loadMode)
	}
}

func TestNewPluginDisabledAnalyzers(t *testing.T) {
	pluginInstance, err := plugin.New(map[string]any{
		"naming_clarity": map[string]any{
			"enabled": false,
		},
		"test_convention": map[string]any{
			"enabled": false,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error creating plugin: %v", err)
	}

	analyzers, buildErr := pluginInstance.BuildAnalyzers()
	if buildErr != nil {
		t.Fatalf("unexpected error building analyzers: %v", buildErr)
	}
	if len(analyzers) != 0 {
		t.Errorf("expected 0 analyzers when all disabled, got %d", len(analyzers))
	}
}

func TestNewPluginError(t *testing.T) {
	_, err := plugin.New(make(chan int))
	if err == nil {
		t.Error("expected error for un-decodable settings channel, got nil")
	}
}
