package testconvention

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"testing"

	"github.com/hrdtr/readable-gcl/pkg/config"
	"golang.org/x/tools/go/analysis"
)

func TestAllowedPrefixesForFolder(t *testing.T) {
	prefixes := allowedPrefixesForFolder("api")
	if len(prefixes) != 2 || prefixes[0] != "TestAPI" || prefixes[1] != "TestApi" {
		t.Errorf("expected [TestAPI, TestApi], got %v", prefixes)
	}

	samePrefixes := allowedPrefixesForFolder("API")
	if len(samePrefixes) != 1 || samePrefixes[0] != "TestAPI" {
		t.Errorf("expected [TestAPI], got %v", samePrefixes)
	}
}

func TestIsSkippedPath(t *testing.T) {
	skippedDirs := []string{"vendor", ".git", "node_modules"}

	if !isSkippedPath("vendor/pkg/foo_test.go", skippedDirs) {
		t.Error("expected vendor path to be skipped")
	}
	if !isSkippedPath("/root/.git/hooks/foo_test.go", skippedDirs) {
		t.Error("expected .git path to be skipped")
	}
	if isSkippedPath("pkg/domain/user_test.go", skippedDirs) {
		t.Error("expected pkg/domain/user_test.go not to be skipped")
	}
}

func TestAnalyzerDisabled(t *testing.T) {
	disabledAnalyzer := NewAnalyzer(config.TestConventionConfig{Enabled: false})
	analyzerResult, err := disabledAnalyzer.Run(nil)
	if err != nil || analyzerResult != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", analyzerResult, err)
	}

	// Also directly invoke run with disabled config
	runResult, runErr := run(nil, config.TestConventionConfig{Enabled: false})
	if runErr != nil || runResult != nil {
		t.Errorf("expected run to exit early when disabled, got (%v, %v)", runResult, runErr)
	}
}

func TestCheckTestFunctionName(t *testing.T) {
	fset := token.NewFileSet()
	reportedMessages := []string{}
	pass := &analysis.Pass{
		Fset: fset,
		Report: func(diag analysis.Diagnostic) {
			reportedMessages = append(reportedMessages, diag.Message)
		},
	}

	// 1. Nil funcDecl, nil Name, or nil Type
	checkTestFunctionName(pass, nil, "foo_test.go", nil, config.TestConventionConfig{})
	checkTestFunctionName(pass, &ast.FuncDecl{Name: nil}, "foo_test.go", nil, config.TestConventionConfig{})
	checkTestFunctionName(pass, &ast.FuncDecl{Name: &ast.Ident{Name: "Helper"}, Type: nil}, "foo_test.go", nil, config.TestConventionConfig{})

	// 2. Non-Test function name
	checkTestFunctionName(pass, &ast.FuncDecl{Name: &ast.Ident{Name: "Helper"}, Type: &ast.FuncType{}}, "foo_test.go", nil, config.TestConventionConfig{})
	if len(reportedMessages) != 0 {
		t.Errorf("expected no reports for helper functions, got %v", reportedMessages)
	}

	// 3. Pattern mismatch
	pattern := regexp.MustCompile(`^Test[A-Z][a-zA-Z0-9]+$`)
	conventionWithPattern := config.TestConventionConfig{Pattern: `^Test[A-Z][a-zA-Z0-9]+$`}
	checkTestFunctionName(pass, &ast.FuncDecl{
		Name: &ast.Ident{Name: "Test_snake_case"},
		Type: &ast.FuncType{},
	}, "foo_test.go", pattern, conventionWithPattern)

	if len(reportedMessages) != 1 {
		t.Fatalf("expected 1 report for pattern mismatch, got %v", reportedMessages)
	}
	reportedMessages = nil

	// 4. RequireFolderPrefix
	conventionFolder := config.TestConventionConfig{RequireFolderPrefix: true}
	// valid prefix: folder is "service", func is TestServiceLogin
	checkTestFunctionName(pass, &ast.FuncDecl{
		Name: &ast.Ident{Name: "TestServiceLogin"},
		Type: &ast.FuncType{},
	}, "/project/service/login_test.go", nil, conventionFolder)
	if len(reportedMessages) != 0 {
		t.Errorf("expected valid folder prefix, got %v", reportedMessages)
	}

	// invalid prefix: folder is "service", func is TestLogin
	checkTestFunctionName(pass, &ast.FuncDecl{
		Name: &ast.Ident{Name: "TestLogin"},
		Type: &ast.FuncType{},
	}, "/project/service/login_test.go", nil, conventionFolder)
	if len(reportedMessages) != 1 {
		t.Fatalf("expected 1 report for invalid folder prefix, got %v", reportedMessages)
	}
	reportedMessages = nil

	// 5. RequireTierSuffix
	conventionTier := config.TestConventionConfig{RequireTierSuffix: true}
	// e2e missing E2E suffix
	checkTestFunctionName(pass, &ast.FuncDecl{
		Name: &ast.Ident{Name: "TestUser"},
		Type: &ast.FuncType{},
	}, "user_e2e_test.go", nil, conventionTier)
	if len(reportedMessages) != 1 {
		t.Fatalf("expected report for missing E2E suffix, got %v", reportedMessages)
	}
	reportedMessages = nil

	// e2e with E2E suffix
	checkTestFunctionName(pass, &ast.FuncDecl{
		Name: &ast.Ident{Name: "TestUserE2E"},
		Type: &ast.FuncType{},
	}, "user_e2e_test.go", nil, conventionTier)
	if len(reportedMessages) != 0 {
		t.Errorf("expected no report for valid E2E suffix, got %v", reportedMessages)
	}

	// integration missing Integration suffix
	checkTestFunctionName(pass, &ast.FuncDecl{
		Name: &ast.Ident{Name: "TestUser"},
		Type: &ast.FuncType{},
	}, "user_integration_test.go", nil, conventionTier)
	if len(reportedMessages) != 1 {
		t.Fatalf("expected report for missing Integration suffix, got %v", reportedMessages)
	}
	reportedMessages = nil

	// integration with Integration suffix
	checkTestFunctionName(pass, &ast.FuncDecl{
		Name: &ast.Ident{Name: "TestUserIntegration"},
		Type: &ast.FuncType{},
	}, "user_integration_test.go", nil, conventionTier)
	if len(reportedMessages) != 0 {
		t.Errorf("expected no report for valid Integration suffix, got %v", reportedMessages)
	}

	// unit test missing Unit suffix
	checkTestFunctionName(pass, &ast.FuncDecl{
		Name: &ast.Ident{Name: "TestUser"},
		Type: &ast.FuncType{},
	}, "user_test.go", nil, conventionTier)
	if len(reportedMessages) != 1 {
		t.Fatalf("expected report for missing Unit suffix, got %v", reportedMessages)
	}
	reportedMessages = nil

	// unit test with Unit suffix
	checkTestFunctionName(pass, &ast.FuncDecl{
		Name: &ast.Ident{Name: "TestUserUnit"},
		Type: &ast.FuncType{},
	}, "user_test.go", nil, conventionTier)
	if len(reportedMessages) != 0 {
		t.Errorf("expected no report for valid Unit suffix, got %v", reportedMessages)
	}
}

func TestCheckTestTableSubtestsNilBody(t *testing.T) {
	checkTestTableSubtests(nil, &ast.FuncDecl{Body: nil})
}

func TestCheckMatchingSourceFileEdgeCases(t *testing.T) {
	pass := &analysis.Pass{}
	fileNode := &ast.File{}

	// Not a test file
	checkMatchingSourceFile(pass, fileNode, "main.go")

	// Base name is empty
	checkMatchingSourceFile(pass, fileNode, "_test.go")
	checkMatchingSourceFile(pass, fileNode, "_integration_test.go")
	checkMatchingSourceFile(pass, fileNode, "_e2e_test.go")
}

func TestRunInvalidRegexPattern(t *testing.T) {
	pass := &analysis.Pass{}
	invalidConfiguration := config.TestConventionConfig{
		Enabled: true,
		Pattern: "[invalid-regex",
	}

	executionResult, err := run(pass, invalidConfiguration)
	if err == nil || executionResult != nil {
		t.Errorf("expected regex compile error for invalid pattern, got (%v, %v)", executionResult, err)
	}
}

func TestCheckTestTableSubtestsWithPlainCallAndAssert(t *testing.T) {
	src := `package foo
func TestLoop(t *testing.T) {
	items := []int{1}
	for _, item := range items {
		println(item)
		assert.Equal(t, item, 1)
	}
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "foo_test.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	reported := false
	pass := &analysis.Pass{
		Fset: fset,
		Report: func(diag analysis.Diagnostic) {
			reported = true
		},
	}

	funcDecl := file.Decls[0].(*ast.FuncDecl)
	checkTestTableSubtests(pass, funcDecl)
	if !reported {
		t.Error("expected report for assert without t.Run")
	}
}

func TestRunNonTestFileAndHelper(t *testing.T) {
	src := `package foo
func Helper() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "foo.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Skipped file
	skippedFile, err := parser.ParseFile(fset, "vendor/foo_test.go", "package foo", 0)
	if err != nil {
		t.Fatal(err)
	}

	// File with no pos in fset
	emptyFsetFile := &ast.File{
		Package: token.NoPos,
	}

	pass := &analysis.Pass{
		Fset:   fset,
		Files:  []*ast.File{file, skippedFile, emptyFsetFile},
		Report: func(diag analysis.Diagnostic) {},
	}

	conventionConfiguration := config.DefaultConfig("pragmatic").TestConvention
	_, runErr := run(pass, conventionConfiguration)
	if runErr != nil {
		t.Errorf("unexpected run error: %v", runErr)
	}
}
