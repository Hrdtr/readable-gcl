package testconvention

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hrdtr/readable-gcl/pkg/config"
	"github.com/hrdtr/readable-gcl/pkg/namingclarity"
	"golang.org/x/tools/go/analysis"
)

func run(pass *analysis.Pass, configuration config.TestConventionConfig) (any, error) {
	if pass == nil {
		return nil, nil
	}

	var compiledPattern *regexp.Regexp
	if configuration.Pattern != "" {
		var compileErr error
		compiledPattern, compileErr = regexp.Compile(configuration.Pattern)
		if compileErr != nil {
			return nil, compileErr
		}
	}

	for _, fileNode := range pass.Files {
		tokenFile := pass.Fset.File(fileNode.Pos())
		if tokenFile == nil {
			continue
		}

		filePath := tokenFile.Name()
		if !strings.HasSuffix(filePath, "_test.go") {
			continue
		}

		if isSkippedPath(filePath, configuration.SkippedDirectories) {
			continue
		}

		if configuration.RequireMatchingSourceFile {
			checkMatchingSourceFile(pass, fileNode, filePath)
		}

		for _, declaration := range fileNode.Decls {
			functionDeclaration, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction {
				continue
			}

			checkTestFunctionName(pass, functionDeclaration, filePath, compiledPattern, configuration)
			if configuration.EnforceTableSubtests {
				checkTestTableSubtests(pass, functionDeclaration)
			}
		}
	}

	return nil, nil
}

func isSkippedPath(filePath string, skippedDirectories []string) bool {
	normalizedPath := filepath.ToSlash(filePath)
	pathSegments := strings.Split(normalizedPath, "/")

	skippedMap := make(map[string]bool, len(skippedDirectories))
	for _, directory := range skippedDirectories {
		skippedMap[directory] = true
	}

	for _, segment := range pathSegments {
		if skippedMap[segment] {
			return true
		}
	}

	return false
}

func allowedPrefixesForFolder(folderName string) []string {
	folderPascal := namingclarity.ToPascalCase(folderName)
	standardPrefix := "Test" + folderPascal
	uppercasePrefix := "Test" + strings.ToUpper(folderPascal)
	if standardPrefix == uppercasePrefix {
		return []string{standardPrefix}
	}
	return []string{uppercasePrefix, standardPrefix}
}

func checkTestFunctionName(
	pass *analysis.Pass,
	functionDeclaration *ast.FuncDecl,
	filePath string,
	compiledPattern *regexp.Regexp,
	configuration config.TestConventionConfig,
) {
	if functionDeclaration == nil || functionDeclaration.Name == nil || functionDeclaration.Type == nil {
		return
	}

	functionName := functionDeclaration.Name.Name
	if !strings.HasPrefix(functionName, "Test") {
		return
	}

	if compiledPattern != nil && !compiledPattern.MatchString(functionName) {
		pass.Reportf(
			functionDeclaration.Name.Pos(),
			"test function name %q does not match configured pattern %q",
			functionName,
			configuration.Pattern,
		)
		return
	}

	if configuration.RequireFolderPrefix {
		absolutePath, err := filepath.Abs(filePath)
		if err == nil {
			filePath = absolutePath
		}

		folderName := filepath.Base(filepath.Dir(filePath))
		expectedPrefixes := allowedPrefixesForFolder(folderName)

		hasValidPrefix := false
		for _, expectedPrefix := range expectedPrefixes {
			if strings.HasPrefix(functionName, expectedPrefix) {
				hasValidPrefix = true
				break
			}
		}

		if !hasValidPrefix {
			prefixLabel := expectedPrefixes[0]
			if len(expectedPrefixes) > 1 {
				prefixLabel = strings.Join(expectedPrefixes, "' or '")
			}
			pass.Reportf(
				functionDeclaration.Name.Pos(),
				"test function name in folder '%s' must start with '%s'",
				folderName,
				prefixLabel,
			)
			return
		}
	}

	if configuration.RequireTierSuffix {
		switch {
		case strings.HasSuffix(filePath, "_e2e_test.go"):
			if !strings.HasSuffix(functionName, "E2E") {
				pass.Reportf(
					functionDeclaration.Name.Pos(),
					"test in e2e file must end with 'E2E'",
				)
			}
		case strings.HasSuffix(filePath, "_integration_test.go"):
			if !strings.HasSuffix(functionName, "Integration") {
				pass.Reportf(
					functionDeclaration.Name.Pos(),
					"test in integration file must end with 'Integration'",
				)
			}
		default:
			if !strings.HasSuffix(functionName, "Unit") {
				pass.Reportf(
					functionDeclaration.Name.Pos(),
					"test in unit test file must end with 'Unit'",
				)
			}
		}
	}
}

func checkTestTableSubtests(pass *analysis.Pass, functionDeclaration *ast.FuncDecl) {
	if functionDeclaration.Body == nil {
		return
	}

	ast.Inspect(functionDeclaration.Body, func(node ast.Node) bool {
		if _, isFunctionLiteral := node.(*ast.FuncLit); isFunctionLiteral {
			return false
		}

		rangeStatement, isRangeStatement := node.(*ast.RangeStmt)
		if !isRangeStatement {
			return true
		}

		hasAssertions := false
		hasRunMethod := false
		hasBreakStatement := false

		ast.Inspect(rangeStatement.Body, func(childNode ast.Node) bool {
			if branchStatement, isBranch := childNode.(*ast.BranchStmt); isBranch && branchStatement.Tok == token.BREAK {
				hasBreakStatement = true
			}

			callExpression, isCall := childNode.(*ast.CallExpr)
			if !isCall {
				return true
			}

			selectorExpression, isSelector := callExpression.Fun.(*ast.SelectorExpr)
			if !isSelector {
				return true
			}

			if selectorExpression.Sel.Name == "Run" {
				hasRunMethod = true
			}

			methodName := selectorExpression.Sel.Name
			if methodName == "Fatalf" || methodName == "Errorf" || methodName == "Fatal" ||
				methodName == "Error" || methodName == "Fail" || methodName == "FailNow" {
				hasAssertions = true
			}

			if identifier, isIdent := selectorExpression.X.(*ast.Ident); isIdent {
				if identifier.Name == "assert" || identifier.Name == "require" {
					hasAssertions = true
				}
			}

			return true
		})

		if hasAssertions && !hasRunMethod && !hasBreakStatement {
			pass.Reportf(
				rangeStatement.Pos(),
				"table-driven test loop in '%s' must execute cases using t.Run",
				functionDeclaration.Name.Name,
			)
		}

		return true
	})
}

func checkMatchingSourceFile(pass *analysis.Pass, fileNode *ast.File, filePath string) {
	fileName := filepath.Base(filePath)

	var baseName string
	switch {
	case strings.HasSuffix(fileName, "_integration_test.go"):
		baseName = strings.TrimSuffix(fileName, "_integration_test.go")
	case strings.HasSuffix(fileName, "_e2e_test.go"):
		baseName = strings.TrimSuffix(fileName, "_e2e_test.go")
	case strings.HasSuffix(fileName, "_test.go"):
		baseName = strings.TrimSuffix(fileName, "_test.go")
	default:
		return
	}

	if baseName == "" {
		return
	}

	directoryPath := filepath.Dir(filePath)
	directoryName := filepath.Base(directoryPath)

	// Allow if baseName matches the containing directory name
	if baseName == directoryName {
		return
	}

	// Allow if corresponding source file exists in the same directory
	sourceFilePath := filepath.Join(directoryPath, baseName+".go")
	if _, statErr := os.Stat(sourceFilePath); statErr == nil {
		return
	}

	pass.Reportf(
		fileNode.Package,
		"test file %q has no corresponding implementation file '%s.go' and does not match directory name %q",
		fileName,
		baseName,
		directoryName,
	)
}
