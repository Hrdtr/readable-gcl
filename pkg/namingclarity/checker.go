package namingclarity

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
	"unicode"

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
					checkFields(pass, targetNode.Recv.List, configuration)
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
					checkFields(pass, targetNode.Params.List, configuration)
				}
				if targetNode.Results != nil {
					checkFields(pass, targetNode.Results.List, configuration)
				}
			}

			return true
		})
	}

	return nil, nil
}

func checkFuncParams(pass *analysis.Pass, parameters *ast.FieldList, configuration config.NamingClarityConfig) {
	if parameters == nil || !configuration.SingleLetterArgs {
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

func checkReceivers(pass *analysis.Pass, receivers *ast.FieldList, configuration config.NamingClarityConfig) {
	if receivers == nil || !configuration.SingleLetterArgs {
		return
	}

	for _, receiverField := range receivers.List {
		var concreteTypeName string
		if receiverField.Type != nil {
			concreteTypeName, _ = resolveConcreteTypeName(pass.TypesInfo.TypeOf(receiverField.Type))
		}

		for _, identifier := range receiverField.Names {
			if len(identifier.Name) == 1 && identifier.Name != "_" {
				pass.Reportf(
					identifier.Pos(),
					"single-letter receiver prohibited: %s",
					identifier.Name,
				)
				continue
			}

			inspectIdentifier(pass, identifier, concreteTypeName, configuration)
		}
	}
}

func checkFields(pass *analysis.Pass, fieldList []*ast.Field, configuration config.NamingClarityConfig) {
	for _, currentField := range fieldList {
		if len(currentField.Names) == 0 {
			continue
		}

		var concreteTypeName string
		var hasConcreteTypeName bool

		if currentField.Type != nil {
			concreteTypeName, hasConcreteTypeName = resolveConcreteTypeName(pass.TypesInfo.TypeOf(currentField.Type))
		}

		for _, identifier := range currentField.Names {
			identConcreteName := concreteTypeName
			if !hasConcreteTypeName {
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
	for _, leftHandExpr := range assignmentStatement.Lhs {
		if identifier, isIdent := leftHandExpr.(*ast.Ident); isIdent {
			checkProhibitedSuffix(pass, identifier, configuration)
		}
	}

	if len(assignmentStatement.Lhs) == len(assignmentStatement.Rhs) {
		for assignmentIndex, rightHandExpr := range assignmentStatement.Rhs {
			identifier, isIdent := assignmentStatement.Lhs[assignmentIndex].(*ast.Ident)
			if !isIdent || HasProhibitedSuffix(identifier.Name, configuration.BannedSuffixes) {
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
				if !isIdent || HasProhibitedSuffix(identifier.Name, configuration.BannedSuffixes) {
					continue
				}

				concreteTypeName, _ := resolveConcreteTypeName(tupleType.At(tupleIndex).Type())
				inspectIdentifier(pass, identifier, concreteTypeName, configuration)
			}
			return
		}

		for _, leftHandExpr := range assignmentStatement.Lhs {
			identifier, isIdent := leftHandExpr.(*ast.Ident)
			if !isIdent || HasProhibitedSuffix(identifier.Name, configuration.BannedSuffixes) {
				continue
			}
			concreteTypeName, _ := resolveConcreteTypeNameFromIdent(pass, identifier)
			inspectIdentifier(pass, identifier, concreteTypeName, configuration)
		}
	}
}

func checkValueSpec(pass *analysis.Pass, valueSpecification *ast.ValueSpec, configuration config.NamingClarityConfig) {
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
	if keyIdentifier, isIdent := rangeStatement.Key.(*ast.Ident); isIdent {
		inspectIdentifier(pass, keyIdentifier, "", configuration)
	}
	if valueIdentifier, isIdent := rangeStatement.Value.(*ast.Ident); isIdent {
		inspectIdentifier(pass, valueIdentifier, "", configuration)
	}
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

	if HasProhibitedSuffix(identifier.Name, configuration.BannedSuffixes) {
		pass.Reportf(
			identifier.Pos(),
			"meaningless representation suffix prohibited: %s",
			identifier.Name,
		)
		return true
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
	if flaggedWord, hasDisallowedWord := CheckDisallowedWord(componentWords, configuration.BannedWords); hasDisallowedWord {
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

func checkIdentifier(pass *analysis.Pass, identifier *ast.Ident, concreteTypeName string, configuration config.NamingClarityConfig) bool {
	if isExemptIdentifier(pass, identifier, concreteTypeName, configuration) {
		return false
	}

	reportMessage, shouldReport := shouldReportUnclearName(identifier.Name, concreteTypeName, configuration)
	if !shouldReport {
		return false
	}

	pass.Reportf(identifier.Pos(), "%s", reportMessage)
	return true
}

func isExemptIdentifier(pass *analysis.Pass, identifier *ast.Ident, concreteTypeName string, configuration config.NamingClarityConfig) bool {
	identifierName := identifier.Name
	for _, exempt := range configuration.ExemptIdentifiers {
		if identifierName == exempt {
			return true
		}
	}

	if isTestingT(pass, identifier) {
		return true
	}

	if rule, exists := findSpecialTypeRule(concreteTypeName, configuration.TypeRules); exists {
		if rule.AllowAnyDomainName {
			return true
		}
		if rule.CanonicalName != "" && identifierName == rule.CanonicalName {
			return true
		}
		if rule.ExpectedTail != "" && strings.HasSuffix(identifierName, rule.ExpectedTail) {
			return true
		}
		if rule.AllowPlural {
			if strings.HasSuffix(identifierName, "s") && strings.HasSuffix(strings.TrimSuffix(identifierName, "s"), rule.ExpectedTail) {
				return true
			}
		}
		if concreteTypeName == "error" && IsSentinelError(identifierName) {
			return true
		}
	}

	return false
}

func shouldReportUnclearName(identifierName string, concreteTypeName string, configuration config.NamingClarityConfig) (string, bool) {
	if identifierName == "" || identifierName == "_" || concreteTypeName == "" {
		return "", false
	}

	// Consult configured special type rules
	if rule, exists := findSpecialTypeRule(concreteTypeName, configuration.TypeRules); exists {
		if rule.AllowAnyDomainName {
			return "", false
		}
		if rule.CanonicalName != "" && identifierName == rule.CanonicalName {
			return "", false
		}
		if rule.ExpectedTail != "" && strings.HasSuffix(identifierName, rule.ExpectedTail) {
			return "", false
		}
		if rule.AllowPlural {
			if strings.HasSuffix(identifierName, "s") && strings.HasSuffix(strings.TrimSuffix(identifierName, "s"), rule.ExpectedTail) {
				return "", false
			}
		}
		if concreteTypeName == "error" {
			if IsSentinelError(identifierName) {
				return "", false
			}
			return fmt.Sprintf("%q should be named %q, suffixed with %q, or prefixed with %q", identifierName, "err", "Err", "Err"), true
		}
		return fmt.Sprintf("%q should be named %q or suffixed with %q", identifierName, rule.CanonicalName, rule.ExpectedTail), true
	}

	if !configuration.EnforceTypeTail {
		return "", false
	}

	bareName := BareTypeName(concreteTypeName)
	canonicalName := CanonicalVariableName(bareName)
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
