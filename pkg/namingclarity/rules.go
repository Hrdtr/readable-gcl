package namingclarity

import (
	"strings"
	"unicode"
)

// KnownInitialisms contains standard Go acronyms to preserve or lowercase consistently.
var KnownInitialisms = []string{
	"OAuth", "OIDC", "UUID", "HTTP", "JSON", "HTML", "SMTP", "TOTP",
	"REST", "JWT", "URL", "TTL", "API", "SMS", "OTP", "TLS", "TCP",
	"UDP", "SQL", "URI", "UID", "GID", "IP", "ID", "DB",
}

// SplitIdentifierWords divides an identifier into individual component words,
// correctly separating camelCase, PascalCase, snake_case, and acronym boundaries.
func SplitIdentifierWords(name string) []string {
	var words []string
	var currentRunes []rune
	allRunes := []rune(name)

	for index := 0; index < len(allRunes); index++ {
		character := allRunes[index]
		if character == '_' || unicode.IsDigit(character) {
			if len(currentRunes) > 0 {
				words = append(words, string(currentRunes))
				currentRunes = nil
			}
			continue
		}

		if unicode.IsUpper(character) {
			if len(currentRunes) > 0 {
				previousIsUpper := unicode.IsUpper(allRunes[index-1])
				nextIsLower := index+1 < len(allRunes) && unicode.IsLower(allRunes[index+1])
				if !previousIsUpper || nextIsLower {
					words = append(words, string(currentRunes))
					currentRunes = nil
				}
			}
		}

		currentRunes = append(currentRunes, character)
	}

	if len(currentRunes) > 0 {
		words = append(words, string(currentRunes))
	}

	return words
}

// HasProhibitedSuffix checks whether an identifier ends with a forbidden Hungarian notation suffix
// while exempting domain types like math/big.Int.
func HasProhibitedSuffix(name string, bannedSuffixes []string) bool {
	if name == "" || name == "_" {
		return false
	}

	// BigInt and BigInteger represent math/big.Int, not Hungarian notation "Int".
	if strings.HasSuffix(name, "BigInt") || strings.HasSuffix(name, "bigInt") ||
		strings.HasSuffix(name, "BigInteger") || strings.HasSuffix(name, "bigInteger") {
		return false
	}

	for _, suffix := range bannedSuffixes {
		if strings.HasSuffix(name, suffix) && len(name) > len(suffix) {
			return true
		}
	}

	return false
}

// CheckBadInitialism examines component words for unidiomatic PascalCase initialisms (e.g. "Id", "Url").
func CheckBadInitialism(words []string, badInitialisms map[string]string) (string, string, bool) {
	for _, word := range words {
		if canonicalReplacement, isBad := badInitialisms[word]; isBad {
			return word, canonicalReplacement, true
		}
	}
	return "", "", false
}

// CheckDisallowedWord evaluates whether any component word matches a configured banned word list.
func CheckDisallowedWord(words []string, bannedWords []string) (string, bool) {
	bannedMap := make(map[string]bool, len(bannedWords))
	for _, bannedWord := range bannedWords {
		bannedMap[strings.ToLower(bannedWord)] = true
	}

	for _, word := range words {
		lowerWord := strings.ToLower(word)
		if bannedMap[lowerWord] {
			return lowerWord, true
		}
	}

	return "", false
}

// ToPascalCase converts a separated string into a PascalCase identifier.
func ToPascalCase(input string) string {
	parts := strings.FieldsFunc(input, func(character rune) bool {
		return character == '_' || character == '-' || character == '.' || character == ' ' || character == '/'
	})

	var builder strings.Builder
	for _, part := range parts {
		builder.WriteString(strings.ToUpper(part[:1]))
		builder.WriteString(part[1:])
	}

	return builder.String()
}

// CanonicalVariableName computes the standard variable name for a given type name.
func CanonicalVariableName(typeName string) string {
	bareName := BareTypeName(typeName)
	return lowerPrefix(bareName)
}

// BareTypeName strips the package qualifier from a type name (e.g. "context.Context" -> "Context").
func BareTypeName(concreteName string) string {
	lastDotIndex := strings.LastIndex(concreteName, ".")
	if lastDotIndex != -1 {
		return concreteName[lastDotIndex+1:]
	}
	return concreteName
}

// IsSentinelError checks whether an identifier adheres to standard Go error sentinel naming (ErrSomething).
func IsSentinelError(name string) bool {
	return strings.HasPrefix(name, "Err") && len(name) > 3 && unicode.IsUpper(rune(name[3]))
}

func lowerPrefix(value string) string {
	if value == "" {
		return value
	}

	// 1. Check if the value starts with any known initialism.
	for _, initialism := range KnownInitialisms {
		if strings.HasPrefix(value, initialism) {
			remainder := value[len(initialism):]
			if remainder == "" {
				return strings.ToLower(initialism)
			}
			if unicode.IsUpper(rune(remainder[0])) {
				return strings.ToLower(initialism) + remainder
			}
		}
	}

	runes := []rune(value)

	// 2. If all runes are uppercase, lowercase all of them.
	allUpper := true
	for _, currentRune := range runes {
		if !unicode.IsUpper(currentRune) {
			allUpper = false
			break
		}
	}
	if allUpper {
		return strings.ToLower(value)
	}

	// 3. If prefix starts with consecutive uppercase runes followed by PascalCase,
	// lowercase the leading acronym up to the last uppercase rune before a lowercase rune.
	if len(runes) > 1 && unicode.IsUpper(runes[0]) && unicode.IsUpper(runes[1]) {
		upperCount := 0
		for upperCount < len(runes) && unicode.IsUpper(runes[upperCount]) {
			if upperCount+1 < len(runes) && unicode.IsLower(runes[upperCount+1]) {
				break
			}
			upperCount++
		}
		for index := 0; index < upperCount; index++ {
			runes[index] = unicode.ToLower(runes[index])
		}
		return string(runes)
	}

	// 4. Standard PascalCase: lowercase only the first rune.
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
