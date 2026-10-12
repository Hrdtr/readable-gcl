package namingclarity

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"github.com/hrdtr/readable-gcl/pkg/config"
	"golang.org/x/tools/go/analysis"
)

func TestShouldReportUnclearName(t *testing.T) {
	defaultConfiguration := config.DefaultConfig("pragmatic").NamingClarity

	t.Run("emptyOrUnderscore", func(subtest *testing.T) {
		msg, report := shouldReportUnclearName("", "int", defaultConfiguration)
		if report || msg != "" {
			subtest.Errorf("expected no report for empty ident, got %q", msg)
		}
		msg, report = shouldReportUnclearName("_", "int", defaultConfiguration)
		if report || msg != "" {
			subtest.Errorf("expected no report for underscore ident, got %q", msg)
		}
		msg, report = shouldReportUnclearName("val", "", defaultConfiguration)
		if report || msg != "" {
			subtest.Errorf("expected no report for empty type, got %q", msg)
		}
	})

	t.Run("specialTypeRuleRules", func(subtest *testing.T) {
		// CanonicalName match: "ctx" for "context.Context"
		msg, report := shouldReportUnclearName("ctx", "context.Context", defaultConfiguration)
		if report {
			subtest.Errorf("expected valid ctx, got %q", msg)
		}

		// ExpectedTail match: "requestCtx" for "context.Context"
		msg, report = shouldReportUnclearName("requestCtx", "context.Context", defaultConfiguration)
		if report {
			subtest.Errorf("expected valid requestCtx, got %q", msg)
		}

		// Invalid context name: "myContext"
		msg, report = shouldReportUnclearName("myContext", "context.Context", defaultConfiguration)
		if !report {
			subtest.Errorf("expected error for myContext, got none")
		}

		// Error rules: sentinel "ErrNotFound"
		msg, report = shouldReportUnclearName("ErrNotFound", "error", defaultConfiguration)
		if report {
			subtest.Errorf("expected valid sentinel error ErrNotFound, got %q", msg)
		}

		// Error rules: non-sentinel, non-err "failure"
		msg, report = shouldReportUnclearName("failure", "error", defaultConfiguration)
		if !report {
			subtest.Errorf("expected error for failure, got none")
		}

		// Error rules: "validationErr"
		msg, report = shouldReportUnclearName("validationErr", "error", defaultConfiguration)
		if report {
			subtest.Errorf("expected valid validationErr, got %q", msg)
		}

		// Error rules: canonical "err"
		msg, report = shouldReportUnclearName("err", "error", defaultConfiguration)
		if report {
			subtest.Errorf("expected valid err, got %q", msg)
		}

		// Rule with AllowAnyDomainName: true
		configurationWithAny := defaultConfiguration
		configurationWithAny.TypeRules = map[string]config.TypeRule{
			"custom.Type": {AllowAnyDomainName: true},
		}
		msg, report = shouldReportUnclearName("arbitrary", "custom.Type", configurationWithAny)
		if report {
			subtest.Errorf("expected allow any domain name, got %q", msg)
		}

		// Rule with AllowPlural: true
		configurationWithPlural := defaultConfiguration
		configurationWithPlural.TypeRules = map[string]config.TypeRule{
			"item.Element": {
				CanonicalName: "elem",
				ExpectedTail:  "Element",
				AllowPlural:   true,
			},
		}
		msg, report = shouldReportUnclearName("rawElements", "item.Element", configurationWithPlural)
		if report {
			subtest.Errorf("expected plural to be accepted, got %q", msg)
		}
	})

	t.Run("enforceTypeTailStrict", func(subtest *testing.T) {
		strictConfiguration := config.DefaultConfig("strict").NamingClarity

		// Match bareName / canonicalName
		msg, report := shouldReportUnclearName("user", "User", strictConfiguration)
		if report {
			subtest.Errorf("expected user to match User, got %q", msg)
		}

		// Match expectedTail
		msg, report = shouldReportUnclearName("activeUser", "User", strictConfiguration)
		if report {
			subtest.Errorf("expected activeUser to match User, got %q", msg)
		}

		// Match plural
		msg, report = shouldReportUnclearName("activeUsers", "User", strictConfiguration)
		if report {
			subtest.Errorf("expected activeUsers to match User, got %q", msg)
		}

		// Mismatch
		msg, report = shouldReportUnclearName("account", "User", strictConfiguration)
		if !report {
			subtest.Errorf("expected report for account as User, got none")
		}

		// Non-strict mode should return false for normal types
		msg, report = shouldReportUnclearName("account", "User", defaultConfiguration)
		if report {
			subtest.Errorf("expected no report in pragmatic mode, got %q", msg)
		}
	})
}

func TestIsExemptIdentifier(t *testing.T) {
	defaultConfiguration := config.DefaultConfig("pragmatic").NamingClarity

	if !isExemptIdentifier(nil, &ast.Ident{Name: "_"}, "", defaultConfiguration) {
		t.Error("expected _ to be exempt")
	}
	if !isExemptIdentifier(nil, &ast.Ident{Name: "ok"}, "", defaultConfiguration) {
		t.Error("expected ok to be exempt")
	}
	if isExemptIdentifier(nil, &ast.Ident{Name: "foo"}, "", defaultConfiguration) {
		t.Error("expected foo to not be exempt")
	}
}

func TestNilFieldHandlers(t *testing.T) {
	defaultConfiguration := config.DefaultConfig("pragmatic").NamingClarity

	// checkFuncParams with nil
	checkFuncParams(nil, nil, defaultConfiguration)

	// checkReceivers with nil
	checkReceivers(nil, nil, defaultConfiguration)

	// checkStructFields with nil or empty
	checkStructFields(nil, nil, defaultConfiguration)

	// checkFields with nil
	checkFields(nil, nil, defaultConfiguration)
}

func TestFindSpecialTypeRule(t *testing.T) {
	rules := config.DefaultTypeRules()
	rule, found := findSpecialTypeRule("Context", rules)
	if !found || rule.CanonicalName != "ctx" {
		t.Errorf("expected to find Context rule, got (%+v, %v)", rule, found)
	}

	ruleFull, foundFull := findSpecialTypeRule("context.Context", rules)
	if !foundFull || ruleFull.CanonicalName != "ctx" {
		t.Errorf("expected to find context.Context rule, got (%+v, %v)", ruleFull, foundFull)
	}

	_, notFound := findSpecialTypeRule("unknown.Type", rules)
	if notFound {
		t.Error("expected unknown.Type to not be found")
	}
}

func TestResolveConcreteTypeName(t *testing.T) {
	// 1. Nil type
	name, ok := resolveConcreteTypeName(nil)
	if ok || name != "" {
		t.Errorf("expected empty for nil, got (%q, %v)", name, ok)
	}

	// 2. Basic primitive without named wrapper
	name, ok = resolveConcreteTypeName(types.Typ[types.Int])
	if ok || name != "" {
		t.Errorf("expected empty for basic int, got (%q, %v)", name, ok)
	}

	// 3. Universe named type (error)
	errorType := types.Universe.Lookup("error").Type()
	name, ok = resolveConcreteTypeName(errorType)
	if !ok || name != "error" {
		t.Errorf("expected error, got (%q, %v)", name, ok)
	}

	// 4. Custom named package type
	pkg := types.NewPackage("mypkg", "mypkg")
	namedTypeName := types.NewTypeName(0, pkg, "MyStruct", nil)
	namedType := types.NewNamed(namedTypeName, types.Typ[types.Int], nil)

	name, ok = resolveConcreteTypeName(namedType)
	if !ok || name != "mypkg.MyStruct" {
		t.Errorf("expected mypkg.MyStruct, got (%q, %v)", name, ok)
	}

	// 5. Pointer to named type
	ptrType := types.NewPointer(namedType)
	name, ok = resolveConcreteTypeName(ptrType)
	if !ok || name != "mypkg.MyStruct" {
		t.Errorf("expected mypkg.MyStruct for pointer, got (%q, %v)", name, ok)
	}

	// 6. Double pointer to named type
	doublePtr := types.NewPointer(ptrType)
	name, ok = resolveConcreteTypeName(doublePtr)
	if !ok || name != "mypkg.MyStruct" {
		t.Errorf("expected mypkg.MyStruct for double pointer, got (%q, %v)", name, ok)
	}

	// 7. Alias type
	aliasType := types.NewAlias(types.NewTypeName(0, nil, "MyAlias", nil), namedType)
	name, ok = resolveConcreteTypeName(aliasType)
	if !ok || name != "mypkg.MyStruct" {
		t.Errorf("expected mypkg.MyStruct for alias, got (%q, %v)", name, ok)
	}
}

func TestCheckDisallowedIdentifierCompound(t *testing.T) {
	reported := false
	pass := &analysis.Pass{
		Report: func(diag analysis.Diagnostic) {
			reported = true
		},
	}

	clarityConfiguration := config.NamingClarityConfig{
		BannedCompounds: []string{"tempObj"},
	}

	ident := &ast.Ident{Name: "userTempObj"}
	if !checkDisallowedIdentifier(pass, ident, clarityConfiguration) {
		t.Error("expected checkDisallowedIdentifier to return true for banned compound")
	}
	if !reported {
		t.Error("expected diagnostic reported for banned compound")
	}

	cleanIdent := &ast.Ident{Name: "userAccount"}
	if checkDisallowedIdentifier(pass, cleanIdent, clarityConfiguration) {
		t.Error("expected checkDisallowedIdentifier to return false for clean ident")
	}
}

func TestCheckAssignmentFallbackNonTuple(t *testing.T) {
	pass := &analysis.Pass{
		TypesInfo: &types.Info{
			Types: map[ast.Expr]types.TypeAndValue{
				&ast.Ident{Name: "rhsExpr"}: {Type: types.Typ[types.Int]},
			},
			Defs: map[*ast.Ident]types.Object{},
		},
		Report: func(diag analysis.Diagnostic) {},
	}

	assign := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{Name: "firstVal"},
			&ast.Ident{Name: "countInt"}, // has prohibited suffix: triggers continue
			&ast.SelectorExpr{X: &ast.Ident{Name: "entity"}, Sel: &ast.Ident{Name: "field"}}, // not an ident: triggers continue
		},
		Rhs: []ast.Expr{&ast.Ident{Name: "rhsExpr"}},
	}
	clarityConfiguration := config.DefaultConfig("pragmatic").NamingClarity
	checkAssignment(pass, assign, clarityConfiguration)
}

func TestIsExemptIdentifierAdditionalBranches(t *testing.T) {
	fset := token.NewFileSet()
	tokenFile := fset.AddFile("user_test.go", -1, 100)

	pass := &analysis.Pass{
		Fset: fset,
		TypesInfo: &types.Info{
			Types: map[ast.Expr]types.TypeAndValue{},
		},
	}

	configurationWithoutT := config.NamingClarityConfig{
		ExemptIdentifiers: []string{"_"},
		TypeRules: map[string]config.TypeRule{
			"custom.Element": {
				ExpectedTail: "Element",
				AllowPlural:  true,
			},
		},
	}

	// 1. isTestingT with position in _test.go
	identT := &ast.Ident{
		Name:    "t",
		NamePos: tokenFile.Pos(10),
	}
	if !isExemptIdentifier(pass, identT, "", configurationWithoutT) {
		t.Error("expected t to be exempt in test file")
	}

	// 2. Plural rule in isExemptIdentifier
	identPlural := &ast.Ident{Name: "customElements"}
	if !isExemptIdentifier(pass, identPlural, "custom.Element", configurationWithoutT) {
		t.Error("expected plural customElements to be exempt")
	}
}

func TestResolveConcreteTypeNameFromIdentAndExpr(t *testing.T) {
	pkg := types.NewPackage("mypkg", "mypkg")
	namedTypeName := types.NewTypeName(0, pkg, "MyStruct", nil)
	namedType := types.NewNamed(namedTypeName, types.Typ[types.Int], nil)

	identDef := &ast.Ident{Name: "defVar"}
	identUse := &ast.Ident{Name: "useVar"}
	identTypeOnly := &ast.Ident{Name: "typeOnlyVar"}
	identEmpty := &ast.Ident{Name: "emptyVar"}

	exprNil := &ast.Ident{Name: "nilExpr"}
	exprValid := &ast.Ident{Name: "validExpr"}

	pass := &analysis.Pass{
		TypesInfo: &types.Info{
			Defs: map[*ast.Ident]types.Object{
				identDef: types.NewVar(0, pkg, "defVar", namedType),
			},
			Uses: map[*ast.Ident]types.Object{
				identUse: types.NewVar(0, pkg, "useVar", namedType),
			},
			Types: map[ast.Expr]types.TypeAndValue{
				identTypeOnly: {Type: namedType},
				exprValid:     {Type: namedType},
			},
		},
	}

	// Expression resolution
	name, ok := resolveConcreteTypeNameFromExpression(pass, exprNil)
	if ok || name != "" {
		t.Errorf("expected empty for exprNil, got (%q, %v)", name, ok)
	}

	name, ok = resolveConcreteTypeNameFromExpression(pass, exprValid)
	if !ok || name != "mypkg.MyStruct" {
		t.Errorf("expected mypkg.MyStruct, got (%q, %v)", name, ok)
	}

	// Ident resolution
	name, ok = resolveConcreteTypeNameFromIdent(pass, identDef)
	if !ok || name != "mypkg.MyStruct" {
		t.Errorf("expected def resolution to succeed, got (%q, %v)", name, ok)
	}

	name, ok = resolveConcreteTypeNameFromIdent(pass, identUse)
	if !ok || name != "mypkg.MyStruct" {
		t.Errorf("expected use resolution to succeed, got (%q, %v)", name, ok)
	}

	name, ok = resolveConcreteTypeNameFromIdent(pass, identTypeOnly)
	if !ok || name != "mypkg.MyStruct" {
		t.Errorf("expected typeOnly resolution to succeed, got (%q, %v)", name, ok)
	}

	name, ok = resolveConcreteTypeNameFromIdent(pass, identEmpty)
	if ok || name != "" {
		t.Errorf("expected empty resolution to fail, got (%q, %v)", name, ok)
	}
}
