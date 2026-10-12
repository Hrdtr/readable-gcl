package config

// DefaultBadInitialisms returns standard initialisms mapped from unidiomatic PascalCase to standard uppercase initialisms.
func DefaultBadInitialisms() map[string]string {
	return map[string]string{
		"Id":  "ID",
		"Url": "URL",
		"Ttl": "TTL",
		"Jwt": "JWT",
		"Sid": "SID",
	}
}

// DefaultBannedSuffixes returns prohibited Hungarian notation suffixes.
func DefaultBannedSuffixes() []string {
	return []string{
		"Boolean",
		"Bool",
		"Str",
		"Int",
	}
}

// DefaultBannedWords returns ambiguous shorthand component words.
func DefaultBannedWords() []string {
	return []string{
		"cfg",
		"mgr",
		"req",
		"res",
		"resp",
		"val",
		"vals",
		"obj",
		"usr",
	}
}

// DefaultTypeRules returns standard library type naming conventions.
func DefaultTypeRules() map[string]TypeRule {
	return map[string]TypeRule{
		"context.Context": {
			CanonicalName: "ctx",
			ExpectedTail:  "Ctx",
			AllowPlural:   false,
		},
		"context.CancelFunc": {
			CanonicalName: "cancel",
			ExpectedTail:  "Cancel",
			AllowPlural:   false,
		},
		"error": {
			CanonicalName: "err",
			ExpectedTail:  "Err",
			AllowPlural:   false,
		},
		"net.Conn": {
			CanonicalName: "connection",
			ExpectedTail:  "Connection",
			AllowPlural:   true,
		},
		"net.Listener": {
			CanonicalName: "listener",
			ExpectedTail:  "Listener",
			AllowPlural:   true,
		},
		"http.Handler": {
			CanonicalName: "handler",
			ExpectedTail:  "Handler",
			AllowPlural:   true,
		},
		"http.HandlerFunc": {
			CanonicalName: "handler",
			ExpectedTail:  "Handler",
			AllowPlural:   true,
		},
		"sync.Pool": {
			CanonicalName: "pool",
			ExpectedTail:  "Pool",
			AllowPlural:   false,
		},
		"time.Time": {
			CanonicalName: "time",
			ExpectedTail:  "Time",
			AllowPlural:   false,
		},
		"time.Duration": {
			AllowAnyDomainName: true,
		},
		"strings.Builder": {
			AllowAnyDomainName: true,
		},
		"big.Int": {
			AllowAnyDomainName: true,
		},
	}
}

// DefaultSkippedDirectories returns directory names ignored during repository traversal.
func DefaultSkippedDirectories() []string {
	return []string{
		".git",
		"vendor",
		"node_modules",
		"bin",
		"dist",
		".cache",
	}
}

// DefaultConfig generates a populated Config based on the requested preset name ("pragmatic" or "strict").
func DefaultConfig(presetName string) *Config {
	isStrict := presetName == "strict"

	testPattern := ""
	if isStrict {
		testPattern = `^Test[A-Z][a-zA-Z0-9]+$`
	}

	return &Config{
		Preset: presetName,
		NamingClarity: NamingClarityConfig{
			Enabled:           true,
			SingleLetterArgs:  true,
			BannedSuffixes:    DefaultBannedSuffixes(),
			BadInitialisms:    DefaultBadInitialisms(),
			BannedWords:       DefaultBannedWords(),
			BannedCompounds:   []string{},
			EnforceTypeTail:   isStrict,
			TypeRules:         DefaultTypeRules(),
			ExemptIdentifiers: []string{"_", "ok", "t"},
		},
		TestConvention: TestConventionConfig{
			Enabled:                   true,
			EnforceTableSubtests:      true,
			Pattern:                   testPattern,
			RequireFolderPrefix:       false,
			RequireTierSuffix:         false,
			RequireMatchingSourceFile: false,
			SkippedDirectories:        DefaultSkippedDirectories(),
		},
	}
}
