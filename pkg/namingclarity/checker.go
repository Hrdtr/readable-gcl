package namingclarity

import (
	"fmt"
	"go/ast"
	"go/types"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hrdtr/readable-gcl/pkg/config"
	"golang.org/x/tools/go/analysis"
)

func run(pass *analysis.Pass, configuration config.NamingClarityConfig) (any, error) {
	for _, fileNode := range pass.Files {
		ast.Inspect(fileNode, func(node ast.Node) bool {
			switch targetNode := node.(type) {
			case *ast.FuncDecl:
				if targetNode.Recv != nil {
					checkReceivers(pass, targetNode.Recv, configuration)
					checkFields(pass, targetNode.Recv.List, configuration, enabledOr(configuration.CheckReceiverTypes, true))
				}

			case *ast.AssignStmt:
				checkAssignment(pass, targetNode, configuration)

			case *ast.ValueSpec:
				checkValueSpec(pass, targetNode, configuration)

			case *ast.RangeStmt:
				checkRange(pass, targetNode, configuration)

			case *ast.StructType:
				if targetNode.Fields != nil {
					checkStructFields(pass, targetNode.Fields.List, configuration)
				}

			case *ast.FuncType:
				if targetNode.Params != nil {
					checkFuncParams(pass, targetNode.Params, configuration)
					checkFields(pass, targetNode.Params.List, configuration, true)
				}
				if targetNode.Results != nil {
					checkFields(pass, targetNode.Results.List, configuration, true)
				}
			}

			return true
		})
	}

	return nil, nil
}

func checkFuncParams(pass *analysis.Pass, parameters *ast.FieldList, configuration config.NamingClarityConfig) {
	if parameters == nil || !enabledOr(configuration.SingleLetterParams, configuration.SingleLetterArgs) {
		return
	}

	for _, parameterField := range parameters.List {
		var concreteTypeName string
		if parameterField.Type != nil {
			concreteTypeName, _ = resolveConcreteTypeName(pass.TypesInfo.TypeOf(parameterField.Type))
		}

		for _, identifier := range parameterField.Names {
			if len(identifier.Name) == 1 && identifier.Name != "_" {
				if isTestingT(pass, identifier) {
					continue
				}

				pass.Reportf(
					identifier.Pos(),
					"single-letter argument prohibited: %s",
					identifier.Name,
				)
				continue
			}

			inspectIdentifier(pass, identifier, concreteTypeName, configuration)
		}
	}
}

func enabledOr(setting *bool, fallback bool) bool {
	if setting == nil {
		return fallback
	}
	return *setting
}

func checkReceivers(pass *analysis.Pass, receivers *ast.FieldList, configuration config.NamingClarityConfig) {
	if receivers == nil || !enabledOr(configuration.SingleLetterReceivers, configuration.SingleLetterArgs) {
		return
	}

	for _, receiverField := range receivers.List {
		for _, identifier := range receiverField.Names {
			if len(identifier.Name) == 1 && identifier.Name != "_" {
				pass.Reportf(identifier.Pos(), "single-letter receiver prohibited: %s", identifier.Name)
			}
		}
	}
}

func checkFields(pass *analysis.Pass, fieldList []*ast.Field, configuration config.NamingClarityConfig, checkTypes bool) {
	for _, currentField := range fieldList {
		if len(currentField.Names) == 0 {
			continue
		}

		var concreteTypeName string
		var hasConcreteTypeName bool

		if checkTypes && currentField.Type != nil {
			concreteTypeName, hasConcreteTypeName = resolveConcreteTypeName(pass.TypesInfo.TypeOf(currentField.Type))
		}

		for _, identifier := range currentField.Names {
			identConcreteName := concreteTypeName
			if checkTypes && !hasConcreteTypeName {
				identConcreteName, _ = resolveConcreteTypeNameFromIdent(pass, identifier)
			}

			inspectIdentifier(pass, identifier, identConcreteName, configuration)
		}
	}
}

func checkStructFields(pass *analysis.Pass, structFieldList []*ast.Field, configuration config.NamingClarityConfig) {
	for _, structField := range structFieldList {
		if len(structField.Names) == 0 {
			continue
		}

		var concreteTypeName string
		var hasConcreteTypeName bool

		if structField.Type != nil {
			concreteTypeName, hasConcreteTypeName = resolveConcreteTypeName(pass.TypesInfo.TypeOf(structField.Type))
		}

		for _, identifier := range structField.Names {
			// Exported struct fields represent public API or serialized contracts (e.g. ID, URL, Kind).
			if len(identifier.Name) > 0 && unicode.IsUpper(rune(identifier.Name[0])) {
				continue
			}

			identConcreteName := concreteTypeName
			if !hasConcreteTypeName {
				identConcreteName, _ = resolveConcreteTypeNameFromIdent(pass, identifier)
			}

			inspectIdentifier(pass, identifier, identConcreteName, configuration)
		}
	}
}

func checkAssignment(pass *analysis.Pass, assignmentStatement *ast.AssignStmt, configuration config.NamingClarityConfig) {
	useResolvedType := genericTypeMatchingEnabled(configuration)
	for _, leftHandExpr := range assignmentStatement.Lhs {
		if identifier, isIdent := leftHandExpr.(*ast.Ident); isIdent {
			if useResolvedType {
				inspectResolvedIdentifier(pass, identifier, configuration)
			} else {
				checkProhibitedSuffix(pass, identifier, configuration)
			}
		}
	}
	if useResolvedType {
		return
	}

	if len(assignmentStatement.Lhs) == len(assignmentStatement.Rhs) {
		for assignmentIndex, rightHandExpr := range assignmentStatement.Rhs {
			identifier, isIdent := assignmentStatement.Lhs[assignmentIndex].(*ast.Ident)
			if !isIdent || hasBannedRepresentationSuffix(pass, identifier, configuration) {
				continue
			}

			concreteTypeName, _ := resolveConcreteTypeNameFromExpression(pass, rightHandExpr)
			if concreteTypeName == "" {
				concreteTypeName, _ = resolveConcreteTypeNameFromIdent(pass, identifier)
			}

			inspectIdentifier(pass, identifier, concreteTypeName, configuration)
		}
		return
	}

	if len(assignmentStatement.Rhs) == 1 && len(assignmentStatement.Lhs) > 1 {
		rightHandType := pass.TypesInfo.TypeOf(assignmentStatement.Rhs[0])
		if tupleType, isTuple := rightHandType.(*types.Tuple); isTuple && tupleType.Len() == len(assignmentStatement.Lhs) {
			for tupleIndex, leftHandExpr := range assignmentStatement.Lhs {
				identifier, isIdent := leftHandExpr.(*ast.Ident)
				if !isIdent || hasBannedRepresentationSuffix(pass, identifier, configuration) {
					continue
				}

				concreteTypeName, _ := resolveConcreteTypeName(tupleType.At(tupleIndex).Type())
				inspectIdentifier(pass, identifier, concreteTypeName, configuration)
			}
			return
		}

		for _, leftHandExpr := range assignmentStatement.Lhs {
			identifier, isIdent := leftHandExpr.(*ast.Ident)
			if !isIdent || hasBannedRepresentationSuffix(pass, identifier, configuration) {
				continue
			}
			concreteTypeName, _ := resolveConcreteTypeNameFromIdent(pass, identifier)
			inspectIdentifier(pass, identifier, concreteTypeName, configuration)
		}
	}
}

func checkValueSpec(pass *analysis.Pass, valueSpecification *ast.ValueSpec, configuration config.NamingClarityConfig) {
	if genericTypeMatchingEnabled(configuration) {
		for _, identifier := range valueSpecification.Names {
			inspectResolvedIdentifier(pass, identifier, configuration)
		}
		return
	}

	var typeConcreteName string
	var hasTypeConcreteName bool
	if valueSpecification.Type != nil {
		typeConcreteName, hasTypeConcreteName = resolveConcreteTypeName(pass.TypesInfo.TypeOf(valueSpecification.Type))
	}

	for valueIndex, identifier := range valueSpecification.Names {
		concreteTypeName := typeConcreteName
		isResolved := hasTypeConcreteName

		if !isResolved && valueIndex < len(valueSpecification.Values) {
			concreteTypeName, isResolved = resolveConcreteTypeNameFromExpression(pass, valueSpecification.Values[valueIndex])
		}

		if !isResolved {
			concreteTypeName, _ = resolveConcreteTypeNameFromIdent(pass, identifier)
		}

		inspectIdentifier(pass, identifier, concreteTypeName, configuration)
	}
}

func checkRange(pass *analysis.Pass, rangeStatement *ast.RangeStmt, configuration config.NamingClarityConfig) {
	for _, expression := range []ast.Expr{rangeStatement.Key, rangeStatement.Value} {
		if identifier, isIdent := expression.(*ast.Ident); isIdent {
			if genericTypeMatchingEnabled(configuration) {
				inspectResolvedIdentifier(pass, identifier, configuration)
			} else {
				inspectIdentifier(pass, identifier, "", configuration)
			}
		}
	}
}

func genericTypeMatchingEnabled(configuration config.NamingClarityConfig) bool {
	return configuration.EnforceTypeTail && configuration.GenericTypeMatching
}

func inspectResolvedIdentifier(pass *analysis.Pass, identifier *ast.Ident, configuration config.NamingClarityConfig) {
	// An initializer may have a different type from the variable it is assigned to.
	concreteTypeName, _ := resolveConcreteTypeNameFromIdent(pass, identifier)
	inspectIdentifier(pass, identifier, concreteTypeName, configuration)
}

func inspectIdentifier(pass *analysis.Pass, identifier *ast.Ident, concreteTypeName string, configuration config.NamingClarityConfig) bool {
	if checkProhibitedSuffix(pass, identifier, configuration) {
		return true
	}
	if checkDisallowedIdentifier(pass, identifier, configuration) {
		return true
	}
	if concreteTypeName != "" && checkIdentifier(pass, identifier, concreteTypeName, configuration) {
		return true
	}
	return false
}

func checkProhibitedSuffix(pass *analysis.Pass, identifier *ast.Ident, configuration config.NamingClarityConfig) bool {
	if identifier == nil || identifier.Name == "_" {
		return false
	}

	if hasBannedRepresentationSuffix(pass, identifier, configuration) {
		pass.Reportf(
			identifier.Pos(),
			"meaningless representation suffix prohibited: %s",
			identifier.Name,
		)
		return true
	}

	return false
}

func hasBannedRepresentationSuffix(pass *analysis.Pass, identifier *ast.Ident, configuration config.NamingClarityConfig) bool {
	if configuration.BannedSuffixExceptions == nil {
		return HasProhibitedSuffix(identifier.Name, configuration.BannedSuffixes)
	}

	for _, suffix := range configuration.BannedSuffixes {
		if !hasRepresentationSuffix(identifier.Name, suffix) {
			continue
		}
		typeName, resolved := resolveConcreteTypeNameFromIdent(pass, identifier)
		if !resolved || !slices.ContainsFunc(configuration.BannedSuffixExceptions, func(exception config.BannedSuffixException) bool {
			return exception.Suffix == suffix && slices.Contains(exception.Types, typeName)
		}) {
			return true
		}
	}
	return false
}

func checkDisallowedIdentifier(pass *analysis.Pass, identifier *ast.Ident, configuration config.NamingClarityConfig) bool {
	if identifier == nil || identifier.Name == "_" {
		return false
	}

	for _, exempt := range configuration.ExemptIdentifiers {
		if identifier.Name == exempt {
			return false
		}
	}

	componentWords := SplitIdentifierWords(identifier.Name)

	// Check bad initialism casing
	if badWord, canonicalCorrection, hasBad := CheckBadInitialism(componentWords, configuration.BadInitialisms); hasBad {
		pass.Reportf(
			identifier.Pos(),
			"disallowed initialism casing %q in %s, use %q instead",
			badWord,
			identifier.Name,
			canonicalCorrection,
		)
		return true
	}

	// Check disallowed component words
	if flaggedWord, hasDisallowedWord := checkBannedWord(pass, identifier, componentWords, configuration); hasDisallowedWord {
		pass.Reportf(
			identifier.Pos(),
			"disallowed identifier term %q prohibited: %s",
			flaggedWord,
			identifier.Name,
		)
		return true
	}

	// Check compound patterns
	lowerIdentifier := strings.ToLower(identifier.Name)
	for _, compoundPattern := range configuration.BannedCompounds {
		if strings.Contains(lowerIdentifier, strings.ToLower(compoundPattern)) {
			pass.Reportf(
				identifier.Pos(),
				"disallowed compound pattern %q prohibited: %s",
				compoundPattern,
				identifier.Name,
			)
			return true
		}
	}

	return false
}

func checkBannedWord(pass *analysis.Pass, identifier *ast.Ident, words []string, configuration config.NamingClarityConfig) (string, bool) {
	if len(configuration.BannedWordExceptions) == 0 {
		return CheckDisallowedWord(words, configuration.BannedWords)
	}

	typeName := ""
	if pass.TypesInfo != nil {
		typeName, _ = resolveConcreteTypeNameFromIdent(pass, identifier)
		if typeName == "" {
			if basic, ok := types.Unalias(pass.TypesInfo.TypeOf(identifier)).(*types.Basic); ok {
				typeName = basic.Name()
			}
		}
	}

	for index, word := range words {
		if !slices.ContainsFunc(configuration.BannedWords, func(banned string) bool { return strings.EqualFold(word, banned) }) {
			continue
		}
		if !slices.ContainsFunc(configuration.BannedWordExceptions, func(exception config.BannedWordException) bool {
			return matchesWordException(exception, words, index, typeName)
		}) {
			return strings.ToLower(word), true
		}
	}
	return "", false
}

func matchesWordException(exception config.BannedWordException, words []string, index int, typeName string) bool {
	if !strings.EqualFold(exception.Word, words[index]) {
		return false
	}
	if len(exception.Types) > 0 && (typeName == "" || !slices.Contains(exception.Types, typeName)) {
		return false
	}
	if exception.Previous != "" && (index == 0 || !strings.EqualFold(exception.Previous, words[index-1])) {
		return false
	}
	if exception.Next != "" && (index+1 == len(words) || !strings.EqualFold(exception.Next, words[index+1])) {
		return false
	}
	return len(exception.Types) > 0 || exception.Previous != "" || exception.Next != ""
}

func checkIdentifier(pass *analysis.Pass, identifier *ast.Ident, concreteTypeName string, configuration config.NamingClarityConfig) bool {
	if isExemptIdentifier(pass, identifier, configuration) {
		return false
	}

	reportMessage, shouldReport := shouldReportUnclearName(identifier.Name, concreteTypeName, configuration)
	if !shouldReport {
		return false
	}

	pass.Reportf(identifier.Pos(), "%s", reportMessage)
	return true
}

func isExemptIdentifier(pass *analysis.Pass, identifier *ast.Ident, configuration config.NamingClarityConfig) bool {
	identifierName := identifier.Name
	for _, exempt := range configuration.ExemptIdentifiers {
		if identifierName == exempt {
			return true
		}
	}

	return isTestingT(pass, identifier)
}

func shouldReportUnclearName(identifierName string, concreteTypeName string, configuration config.NamingClarityConfig) (string, bool) {
	if identifierName == "" || identifierName == "_" || concreteTypeName == "" {
		return "", false
	}

	// Consult configured special type rules
	if rule, exists := configuredTypeRule(concreteTypeName, configuration); exists {
		message := typeRuleMismatch(identifierName, concreteTypeName, rule)
		return message, message != ""
	}

	if !configuration.EnforceTypeTail {
		return "", false
	}

	bareName := BareTypeName(concreteTypeName)
	canonicalName := CanonicalVariableName(bareName)
	if configuration.GenericTypeMatching {
		message := genericTypeMismatch(identifierName, bareName, canonicalName)
		return message, message != ""
	}
	expectedTail := strings.ToUpper(bareName[:1]) + bareName[1:]

	if strings.EqualFold(identifierName, bareName) || strings.EqualFold(identifierName, canonicalName) {
		return "", false
	}

	nameSingular := strings.TrimSuffix(identifierName, "s")
	if strings.HasSuffix(identifierName, bareName) || strings.HasSuffix(identifierName, expectedTail) ||
		strings.HasSuffix(nameSingular, bareName) || strings.HasSuffix(nameSingular, expectedTail) {
		return "", false
	}

	return fmt.Sprintf("%q should be named %q or suffixed with %q", identifierName, canonicalName, expectedTail), true
}

func genericTypeMismatch(name, bareType, canonicalName string) string {
	if strings.EqualFold(name, canonicalName) || hasFoldedSuffix(name, bareType) ||
		hasFoldedSuffix(strings.TrimSuffix(name, "s"), bareType) {
		return ""
	}

	message := fmt.Sprintf("%q should be named %q or suffixed with %q (case-insensitive, optionally followed by s)", name, canonicalName, bareType)
	if !ast.IsExported(bareType) {
		for index, letter := range bareType {
			if index == 0 || !unicode.IsUpper(letter) {
				continue
			}
			shorter := bareType[index:]
			if strings.EqualFold(name, shorter) || strings.HasSuffix(name, shorter) {
				return ""
			}
			message += fmt.Sprintf(", or match %q exactly (case-insensitive) or as a case-sensitive suffix", shorter)
		}
	}
	return message
}

func hasFoldedSuffix(name, suffix string) bool {
	end := len(name)
	for range suffix {
		_, size := utf8.DecodeLastRuneInString(name[:end])
		end -= size
	}
	return strings.EqualFold(name[end:], suffix)
}

func configuredTypeRule(concreteTypeName string, configuration config.NamingClarityConfig) (config.TypeRule, bool) {
	for index := len(configuration.TypeRulesList) - 1; index >= 0; index-- {
		entry := configuration.TypeRulesList[index]
		if entry.Type == concreteTypeName {
			return entry.TypeRule, true
		}
	}
	return findSpecialTypeRule(concreteTypeName, configuration.TypeRules)
}

// An empty mismatch means the name is accepted. Each alternative is matched and
// described in the same branch so diagnostics cannot drift from the rule.
func typeRuleMismatch(name, concreteTypeName string, rule config.TypeRule) string {
	if rule.AllowAnyDomainName {
		return ""
	}

	var alternatives []string
	for _, allowedName := range slices.Concat([]string{rule.CanonicalName}, rule.AllowedNames) {
		if allowedName != "" {
			if name == allowedName {
				return ""
			}
			alternatives = append(alternatives, fmt.Sprintf("named %q", allowedName))
		}
	}
	for _, suffix := range slices.Concat([]string{rule.ExpectedTail}, rule.AllowedSuffixes) {
		if suffix == "" {
			continue
		}
		if strings.HasSuffix(name, suffix) || (rule.AllowPlural && strings.HasSuffix(name, suffix+"s")) {
			return ""
		}
		description := fmt.Sprintf("suffixed with %q", suffix)
		if rule.AllowPlural {
			description += fmt.Sprintf(" or %q", suffix+"s")
		}
		alternatives = append(alternatives, description)
	}
	if concreteTypeName == "error" {
		if IsSentinelError(name) {
			return ""
		}
		alternatives = append(alternatives, `prefixed with "Err" followed by an uppercase letter`)
	}

	switch len(alternatives) {
	case 0:
		return fmt.Sprintf("%q is not allowed by the naming rule for %s", name, concreteTypeName)
	case 1:
		return fmt.Sprintf("%q should be %s", name, alternatives[0])
	case 2:
		return fmt.Sprintf("%q should be %s or %s", name, alternatives[0], alternatives[1])
	default:
		return fmt.Sprintf("%q should be %s, or %s", name, strings.Join(alternatives[:len(alternatives)-1], ", "), alternatives[len(alternatives)-1])
	}
}

func findSpecialTypeRule(concreteTypeName string, typeRules map[string]config.TypeRule) (config.TypeRule, bool) {
	if rule, exists := typeRules[concreteTypeName]; exists {
		return rule, true
	}
	dotSuffix := "." + concreteTypeName
	for key, rule := range typeRules {
		if strings.HasSuffix(key, dotSuffix) {
			return rule, true
		}
	}
	return config.TypeRule{}, false
}

func isTestingT(pass *analysis.Pass, identifier *ast.Ident) bool {
	if identifier.Name != "t" {
		return false
	}
	if typeValue := pass.TypesInfo.TypeOf(identifier); typeValue != nil {
		if strings.Contains(typeValue.String(), "testing.") {
			return true
		}
	}
	tokenFile := pass.Fset.File(identifier.Pos())
	return tokenFile != nil && strings.HasSuffix(tokenFile.Name(), "_test.go")
}

func resolveConcreteTypeNameFromExpression(pass *analysis.Pass, expression ast.Expr) (string, bool) {
	typeValue := pass.TypesInfo.TypeOf(expression)
	if typeValue == nil {
		return "", false
	}
	return resolveConcreteTypeName(typeValue)
}

func resolveConcreteTypeNameFromIdent(pass *analysis.Pass, identifier *ast.Ident) (string, bool) {
	if defObject := pass.TypesInfo.Defs[identifier]; defObject != nil && defObject.Type() != nil {
		return resolveConcreteTypeName(defObject.Type())
	}
	if useObject := pass.TypesInfo.Uses[identifier]; useObject != nil && useObject.Type() != nil {
		return resolveConcreteTypeName(useObject.Type())
	}
	if typeValue := pass.TypesInfo.TypeOf(identifier); typeValue != nil {
		return resolveConcreteTypeName(typeValue)
	}
	return "", false
}

func resolveConcreteTypeName(typeValue types.Type) (string, bool) {
	if typeValue == nil {
		return "", false
	}

	// Remove pointers
	for {
		pointerType, isPointer := typeValue.(*types.Pointer)
		if !isPointer {
			break
		}
		typeValue = pointerType.Elem()
	}

	// Handle named types
	if namedType, isNamed := typeValue.(*types.Named); isNamed {
		typeName := namedType.Obj().Name()
		if parentPackage := namedType.Obj().Pkg(); parentPackage != nil {
			return parentPackage.Name() + "." + typeName, true
		}
		return typeName, true
	}

	// Handle aliases
	if aliasType, isAlias := typeValue.(*types.Alias); isAlias {
		return resolveConcreteTypeName(aliasType.Rhs())
	}

	return "", false
}
