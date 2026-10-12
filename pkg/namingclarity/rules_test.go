package namingclarity_test

import (
	"reflect"
	"testing"

	"github.com/hrdtr/readable-gcl/pkg/config"
	"github.com/hrdtr/readable-gcl/pkg/namingclarity"
)

func TestSplitIdentifierWords(t *testing.T) {
	testCases := []struct {
		identifier    string
		expectedWords []string
	}{
		{"userID", []string{"user", "ID"}},
		{"httpServer", []string{"http", "Server"}},
		{"oauthToken", []string{"oauth", "Token"}},
		{"userCfg", []string{"user", "Cfg"}},
		{"validate", []string{"validate"}},
		{"result", []string{"result"}},
		{"channel", []string{"channel"}},
		{"serverURL", []string{"server", "URL"}},
		{"getHTTPResponse", []string{"get", "HTTP", "Response"}},
		{"simple", []string{"simple"}},
		{"CamelCase", []string{"Camel", "Case"}},
		{"snake_case_ident", []string{"snake", "case", "ident"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.identifier, func(subtest *testing.T) {
			actualWords := namingclarity.SplitIdentifierWords(testCase.identifier)
			if !reflect.DeepEqual(actualWords, testCase.expectedWords) {
				subtest.Errorf(
					"SplitIdentifierWords(%q) = %v; want %v",
					testCase.identifier,
					actualWords,
					testCase.expectedWords,
				)
			}
		})
	}
}

func TestHasProhibitedSuffix(t *testing.T) {
	bannedSuffixes := config.DefaultBannedSuffixes()

	positiveCases := []string{
		"countInt",
		"activeBool",
		"enabledBoolean",
		"nameStr",
		"userAgeInt",
	}

	for _, positiveCase := range positiveCases {
		t.Run("positive_"+positiveCase, func(subtest *testing.T) {
			if !namingclarity.HasProhibitedSuffix(positiveCase, bannedSuffixes) {
				subtest.Errorf("expected %q to have prohibited suffix", positiveCase)
			}
		})
	}

	negativeCases := []string{
		"bigInt",
		"userBigInt",
		"bigInteger",
		"totalBigInteger",
		"count",
		"active",
		"name",
		"_",
	}

	for _, negativeCase := range negativeCases {
		t.Run("negative_"+negativeCase, func(subtest *testing.T) {
			if namingclarity.HasProhibitedSuffix(negativeCase, bannedSuffixes) {
				subtest.Errorf("expected %q to NOT have prohibited suffix", negativeCase)
			}
		})
	}
}

func TestCheckBadInitialism(t *testing.T) {
	badInitialisms := config.DefaultBadInitialisms()

	t.Run("badInitialismDetected", func(subtest *testing.T) {
		wordsWithBadInitialism := []string{"user", "Id"}
		detectedWord, canonicalCorrection, hasBadInitialism := namingclarity.CheckBadInitialism(wordsWithBadInitialism, badInitialisms)
		if !hasBadInitialism {
			subtest.Errorf("expected to detect bad initialism in %v", wordsWithBadInitialism)
		}
		if detectedWord != "Id" || canonicalCorrection != "ID" {
			subtest.Errorf("unexpected correction: got %s -> %s, want Id -> ID", detectedWord, canonicalCorrection)
		}
	})

	t.Run("goodInitialismPreserved", func(subtest *testing.T) {
		wordsWithGoodInitialism := []string{"user", "ID"}
		_, _, hasBadInitialismInGood := namingclarity.CheckBadInitialism(wordsWithGoodInitialism, badInitialisms)
		if hasBadInitialismInGood {
			subtest.Errorf("expected no bad initialisms in %v", wordsWithGoodInitialism)
		}
	})
}

func TestCheckDisallowedWord(t *testing.T) {
	bannedWords := config.DefaultBannedWords()

	t.Run("flagDisallowedWord", func(subtest *testing.T) {
		wordsWithBannedWord := []string{"user", "Cfg"}
		flaggedWord, hasBannedWord := namingclarity.CheckDisallowedWord(wordsWithBannedWord, bannedWords)
		if !hasBannedWord || flaggedWord != "cfg" {
			subtest.Errorf("expected to flag 'cfg', got %q (%v)", flaggedWord, hasBannedWord)
		}
	})

	t.Run("cleanWordAllowed", func(subtest *testing.T) {
		wordsWithoutBannedWord := []string{"user", "Configuration"}
		_, hasBannedWordInClean := namingclarity.CheckDisallowedWord(wordsWithoutBannedWord, bannedWords)
		if hasBannedWordInClean {
			subtest.Errorf("expected no banned words in %v", wordsWithoutBannedWord)
		}
	})
}

func TestToPascalCase(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"user_service", "UserService"},
		{"api-handler", "ApiHandler"},
		{"auth.token", "AuthToken"},
		{"standalone", "Standalone"},
		{"foo__bar", "FooBar"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(subtest *testing.T) {
			actual := namingclarity.ToPascalCase(testCase.input)
			if actual != testCase.expected {
				subtest.Errorf("ToPascalCase(%q) = %q; want %q", testCase.input, actual, testCase.expected)
			}
		})
	}
}

func TestCanonicalVariableName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"User", "user"},
		{"HTTPClient", "httpClient"},
		{"ID", "id"},
		{"OAuthProvider", "oauthProvider"},
		{"JSONDecoder", "jsonDecoder"},
		{"ABCDef", "abcDef"},
		{"XYZ", "xyz"},
		{"", ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(subtest *testing.T) {
			actual := namingclarity.CanonicalVariableName(testCase.input)
			if actual != testCase.expected {
				subtest.Errorf("CanonicalVariableName(%q) = %q; want %q", testCase.input, actual, testCase.expected)
			}
		})
	}
}

func TestBareTypeName(t *testing.T) {
	if got := namingclarity.BareTypeName("context.Context"); got != "Context" {
		t.Errorf("expected Context, got %s", got)
	}
	if got := namingclarity.BareTypeName("MyType"); got != "MyType" {
		t.Errorf("expected MyType, got %s", got)
	}
}

func TestIsSentinelError(t *testing.T) {
	if !namingclarity.IsSentinelError("ErrNotFound") {
		t.Error("expected ErrNotFound to be sentinel error")
	}
	if namingclarity.IsSentinelError("errNotFound") {
		t.Error("expected errNotFound to not be sentinel error")
	}
	if namingclarity.IsSentinelError("Err") {
		t.Error("expected Err to not be sentinel error")
	}
	if namingclarity.IsSentinelError("Error") {
		t.Error("expected Error to not be sentinel error")
	}
	if namingclarity.IsSentinelError("Errsmall") {
		t.Error("expected Errsmall to not be sentinel error")
	}
}
