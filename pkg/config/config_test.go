package config_test

import (
	"testing"

	"github.com/hrdtr/readable-gcl/pkg/config"
	"gopkg.in/yaml.v3"
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

func TestExplicitListsOverrideDefaults(t *testing.T) {
	const input = `
preset: strict
naming_clarity:
  banned_words: []
  banned_suffixes: [Str]
  type_rules:
    time.Time:
      allowed_names: [now]
  type_rules_list:
    - type: net.Conn
      allowed_names: []
      allowed_suffixes: [Link]
test_convention:
  skipped_dirs: []
`
	for name, decode := range map[string]func() (*config.Config, error){
		"yaml": func() (*config.Config, error) { return config.DecodeYAML([]byte(input)) },
		"host-settings": func() (*config.Config, error) {
			var settings map[string]any
			if err := yaml.Unmarshal([]byte(input), &settings); err != nil {
				return nil, err
			}
			return config.Decode(settings)
		},
	} {
		t.Run(name, func(t *testing.T) {
			configuration, err := decode()
			if err != nil {
				t.Fatal(err)
			}
			naming := configuration.NamingClarity
			if len(naming.BannedWords) != 0 || len(configuration.TestConvention.SkippedDirectories) != 0 {
				t.Fatal("explicit empty lists must clear preset lists")
			}
			if len(naming.BannedSuffixes) != 1 || naming.BannedSuffixes[0] != "Str" || len(naming.ExemptIdentifiers) != 3 {
				t.Fatal("explicit lists must replace, omitted lists must inherit")
			}
			timeRule := naming.TypeRules["time.Time"]
			if timeRule.CanonicalName != "" || timeRule.ExpectedTail != "" || len(timeRule.AllowedNames) != 1 || timeRule.AllowedNames[0] != "now" {
				t.Errorf("map entry must replace the preset rule: %+v", timeRule)
			}
			entries := naming.TypeRulesList
			if len(entries) != 1 || entries[0].Type != "net.Conn" || len(entries[0].AllowedNames) != 0 || len(entries[0].AllowedSuffixes) != 1 || entries[0].AllowedSuffixes[0] != "Link" {
				t.Errorf("list entry must preserve the exact type and explicit alternatives: %+v", entries)
			}
			if naming.TypeRules["error"].CanonicalName != "err" {
				t.Error("unmentioned rule must inherit")
			}
		})
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
