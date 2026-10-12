package config

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// TypeRule specifies canonical variable naming rules for a specific Go type.
type TypeRule struct {
	CanonicalName      string `json:"canonical_name" yaml:"canonical_name"`
	ExpectedTail       string `json:"expected_tail" yaml:"expected_tail"`
	AllowPlural        bool   `json:"allow_plural" yaml:"allow_plural"`
	AllowAnyDomainName bool   `json:"allow_any_domain_name" yaml:"allow_any_domain_name"`
}

// NamingClarityConfig defines settings for the namingclarity analyzer.
type NamingClarityConfig struct {
	Enabled           bool                `json:"enabled" yaml:"enabled"`
	SingleLetterArgs  bool                `json:"single_letter_args" yaml:"single_letter_args"`
	BannedSuffixes    []string            `json:"banned_suffixes" yaml:"banned_suffixes"`
	BadInitialisms    map[string]string   `json:"bad_initialisms" yaml:"bad_initialisms"`
	BannedWords       []string            `json:"banned_words" yaml:"banned_words"`
	BannedCompounds   []string            `json:"banned_compounds" yaml:"banned_compounds"`
	EnforceTypeTail   bool                `json:"enforce_type_tail" yaml:"enforce_type_tail"`
	TypeRules         map[string]TypeRule `json:"type_rules" yaml:"type_rules"`
	ExemptIdentifiers []string            `json:"exempt_identifiers" yaml:"exempt_identifiers"`
}

// TestConventionConfig defines settings for the testconvention analyzer.
type TestConventionConfig struct {
	Enabled                   bool     `json:"enabled" yaml:"enabled"`
	EnforceTableSubtests      bool     `json:"enforce_table_subtests" yaml:"enforce_table_subtests"`
	Pattern                   string   `json:"pattern" yaml:"pattern"`
	RequireFolderPrefix       bool     `json:"require_folder_prefix" yaml:"require_folder_prefix"`
	RequireTierSuffix         bool     `json:"require_tier_suffix" yaml:"require_tier_suffix"`
	RequireMatchingSourceFile bool     `json:"require_matching_source_file" yaml:"require_matching_source_file"`
	SkippedDirectories        []string `json:"skipped_dirs" yaml:"skipped_dirs"`
}

// Config represents top-level configuration for the readablegcl linter plugin.
type Config struct {
	Preset         string               `json:"preset" yaml:"preset"`
	NamingClarity  NamingClarityConfig  `json:"naming_clarity" yaml:"naming_clarity"`
	TestConvention TestConventionConfig `json:"test_convention" yaml:"test_convention"`
}

// Decode extracts a Config from an arbitrary untyped input (such as map[string]any provided by golangci-lint).
func Decode(input any) (*Config, error) {
	if input == nil {
		return DefaultConfig("pragmatic"), nil
	}

	marshaledBytes, marshalErr := json.Marshal(input)
	if marshalErr != nil {
		return nil, fmt.Errorf("failed to marshal settings map: %w", marshalErr)
	}

	presetExtractor := struct {
		Preset string `json:"preset"`
	}{}
	_ = json.Unmarshal(marshaledBytes, &presetExtractor)

	selectedPreset := "pragmatic"
	if presetExtractor.Preset != "" {
		selectedPreset = presetExtractor.Preset
	}

	configuration := DefaultConfig(selectedPreset)

	unmarshalErr := json.Unmarshal(marshaledBytes, configuration)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", unmarshalErr)
	}

	return configuration, nil
}

// DecodeYAML parses YAML data into a Config struct.
func DecodeYAML(yamlData []byte) (*Config, error) {
	presetExtractor := struct {
		Preset string `yaml:"preset"`
	}{}
	_ = yaml.Unmarshal(yamlData, &presetExtractor)

	selectedPreset := "pragmatic"
	if presetExtractor.Preset != "" {
		selectedPreset = presetExtractor.Preset
	}

	configuration := DefaultConfig(selectedPreset)
	unmarshalErr := yaml.Unmarshal(yamlData, configuration)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML configuration: %w", unmarshalErr)
	}

	return configuration, nil
}
