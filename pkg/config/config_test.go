package config_test

import (
	"testing"

	"github.com/hrdtr/readable-gcl/pkg/config"
)

func TestDefaultConfigPragmatic(t *testing.T) {
	configuration := config.DefaultConfig("pragmatic")
	if configuration == nil {
		t.Fatal("expected non-nil configuration")
	}
	if !configuration.NamingClarity.Enabled {
		t.Errorf("expected NamingClarity.Enabled to be true")
	}
	if !configuration.NamingClarity.SingleLetterArgs {
		t.Errorf("expected SingleLetterArgs to be true")
	}
	if configuration.NamingClarity.EnforceTypeTail {
		t.Errorf("expected EnforceTypeTail to be false in pragmatic mode")
	}
	if configuration.TestConvention.RequireFolderPrefix {
		t.Errorf("expected RequireFolderPrefix to be false in pragmatic mode")
	}
	if configuration.TestConvention.RequireTierSuffix {
		t.Errorf("expected RequireTierSuffix to be false in pragmatic mode")
	}
	if configuration.TestConvention.RequireMatchingSourceFile {
		t.Errorf("expected RequireMatchingSourceFile to be false in pragmatic mode")
	}
	if !configuration.TestConvention.EnforceTableSubtests {
		t.Errorf("expected EnforceTableSubtests to be true in pragmatic mode")
	}
}

func TestDefaultConfigStrict(t *testing.T) {
	configuration := config.DefaultConfig("strict")
	if configuration == nil {
		t.Fatal("expected non-nil configuration")
	}
	if !configuration.NamingClarity.EnforceTypeTail {
		t.Errorf("expected EnforceTypeTail to be true in strict mode")
	}
	if configuration.TestConvention.Pattern != `^Test[A-Z][a-zA-Z0-9]+$` {
		t.Errorf("expected pattern `^Test[A-Z][a-zA-Z0-9]+$`, got %s", configuration.TestConvention.Pattern)
	}
}

func TestDecodeNilOrEmpty(t *testing.T) {
	configuration, decodeErr := config.Decode(nil)
	if decodeErr != nil {
		t.Fatalf("unexpected error decoding nil: %v", decodeErr)
	}
	if configuration.Preset != "pragmatic" {
		t.Errorf("expected default preset pragmatic, got %s", configuration.Preset)
	}

	emptyConfiguration, emptyErr := config.Decode(map[string]any{})
	if emptyErr != nil {
		t.Fatalf("unexpected error decoding empty map: %v", emptyErr)
	}
	if emptyConfiguration.Preset != "pragmatic" {
		t.Errorf("expected default preset pragmatic, got %s", emptyConfiguration.Preset)
	}
}

func TestDecodeCustomMap(t *testing.T) {
	inputMap := map[string]any{
		"preset": "pragmatic",
		"naming_clarity": map[string]any{
			"single_letter_args": false,
			"banned_words":       []any{"exampleWord", "temporaryWord"},
		},
		"test_convention": map[string]any{
			"require_folder_prefix":        true,
			"require_matching_source_file": true,
		},
	}
	configuration, decodeErr := config.Decode(inputMap)
	if decodeErr != nil {
		t.Fatalf("unexpected error decoding config: %v", decodeErr)
	}
	if configuration.NamingClarity.SingleLetterArgs {
		t.Errorf("expected SingleLetterArgs overridden to false")
	}
	if len(configuration.NamingClarity.BannedWords) != 2 {
		t.Errorf("expected 2 banned words, got %d", len(configuration.NamingClarity.BannedWords))
	}
	if !configuration.TestConvention.RequireFolderPrefix {
		t.Errorf("expected RequireFolderPrefix overridden to true")
	}
	if !configuration.TestConvention.RequireMatchingSourceFile {
		t.Errorf("expected RequireMatchingSourceFile overridden to true")
	}
}

func TestDecodeYAMLString(t *testing.T) {
	yamlInput := `
preset: strict
naming_clarity:
  single_letter_args: false
  banned_suffixes:
    - "Int"
`
	configuration, decodeErr := config.DecodeYAML([]byte(yamlInput))
	if decodeErr != nil {
		t.Fatalf("unexpected error decoding YAML: %v", decodeErr)
	}
	if configuration.Preset != "strict" {
		t.Errorf("expected preset strict, got %s", configuration.Preset)
	}
	if configuration.NamingClarity.SingleLetterArgs {
		t.Errorf("expected SingleLetterArgs false")
	}
	if len(configuration.NamingClarity.BannedSuffixes) != 1 || configuration.NamingClarity.BannedSuffixes[0] != "Int" {
		t.Errorf("expected [Int], got %v", configuration.NamingClarity.BannedSuffixes)
	}
}

func TestDecodeErrors(t *testing.T) {
	// Channels cannot be marshaled to JSON
	_, marshalErr := config.Decode(make(chan int))
	if marshalErr == nil {
		t.Error("expected marshal error for channel, got nil")
	}

	// String cannot be unmarshaled into Config struct
	_, unmarshalErr := config.Decode("not-a-config-object")
	if unmarshalErr == nil {
		t.Error("expected unmarshal error for string primitive, got nil")
	}

	// Invalid YAML type structure
	_, yamlErr := config.DecodeYAML([]byte("naming_clarity: [invalid, sequence]"))
	if yamlErr == nil {
		t.Error("expected YAML unmarshal error, got nil")
	}
}
