package plugin_test

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"github.com/hrdtr/readable-gcl/pkg/plugin"
)

var runGolangCIIntegration = flag.Bool("golangci-integration", false, "run the module plugin integration test using bin/custom-gcl")

func TestNewPlugin(t *testing.T) {
	pluginInstance, err := plugin.New(map[string]any{
		"preset": "pragmatic",
	})
	if err != nil {
		t.Fatalf("unexpected error creating plugin: %v", err)
	}

	analyzers, buildErr := pluginInstance.BuildAnalyzers()
	if buildErr != nil {
		t.Fatalf("unexpected error building analyzers: %v", buildErr)
	}
	if len(analyzers) != 2 {
		t.Errorf("expected 2 analyzers, got %d", len(analyzers))
	}

	loadMode := pluginInstance.GetLoadMode()
	if loadMode != register.LoadModeTypesInfo {
		t.Errorf("expected load mode %q, got %q", register.LoadModeTypesInfo, loadMode)
	}
}

func TestNewPluginDisabledAnalyzers(t *testing.T) {
	pluginInstance, err := plugin.New(map[string]any{
		"naming_clarity": map[string]any{
			"enabled": false,
		},
		"test_convention": map[string]any{
			"enabled": false,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error creating plugin: %v", err)
	}

	analyzers, buildErr := pluginInstance.BuildAnalyzers()
	if buildErr != nil {
		t.Fatalf("unexpected error building analyzers: %v", buildErr)
	}
	if len(analyzers) != 0 {
		t.Errorf("expected 0 analyzers when all disabled, got %d", len(analyzers))
	}
}

func TestNewPluginError(t *testing.T) {
	_, err := plugin.New(make(chan int))
	if err == nil {
		t.Error("expected error for un-decodable settings channel, got nil")
	}
}

func TestDocumentedExample(t *testing.T) {
	example, err := os.ReadFile("../../.golangci.example.yml")
	if err != nil {
		t.Fatal(err)
	}
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readme), "```yaml\n"+string(example)+"```\n") {
		t.Error("README YAML block must match .golangci.example.yml exactly")
	}
}

func TestGolangCIHost(t *testing.T) {
	if !*runGolangCIIntegration {
		t.Skip("enable with -args -golangci-integration")
	}
	binary, err := filepath.Abs("../../bin/custom-gcl")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(binary); err != nil {
		t.Fatalf("build the plugin with `golangci-lint custom --destination bin`: %v", err)
	}
	example, err := os.ReadFile("../../.golangci.example.yml")
	if err != nil {
		t.Fatal(err)
	}

	directory := t.TempDir()
	if err := os.CopyFS(directory, os.DirFS("../namingclarity/testdata/contract")); err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(directory, os.DirFS("../testconvention/testdata/contract")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, ".golangci.yml"), example, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "go.mod"), []byte("module example.test/contract\n\ngo 1.27.1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	command := exec.CommandContext(t.Context(), binary, "run", "--output.json.path=stdout", "--output.text.path=", "--show-stats=false", "--uniq-by-line=false", "--max-issues-per-linter=0", "--max-same-issues=0", "./...")
	command.Dir = directory
	output, err := command.Output()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("expected lint findings (exit 1), got %v\n%s\n%s", err, output, command.Stderr)
	}
	var result struct {
		Issues []struct {
			FromLinter string
			Text       string
			Pos        struct {
				Filename string
				Line     int
				Column   int
			}
		}
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("decode host output: %v\n%s", err, output)
	}
	want := hostExpectations(t, os.DirFS(directory))
	var failures []string
	for _, issue := range result.Issues {
		location := hostLocation{filepath.ToSlash(issue.Pos.Filename), issue.Pos.Line}
		if issue.FromLinter != "readablegcl" || !matchHostExpectation(want[location], issue.Text) {
			failures = append(failures, fmt.Sprintf("unexpected finding at %s:%d:%d: %s (%s)", location.file, location.line, issue.Pos.Column, issue.Text, issue.FromLinter))
		}
	}
	for location, expectations := range want {
		for prefix, found := range expectations {
			if !found {
				failures = append(failures, fmt.Sprintf("missing finding at %s:%d: %s", location.file, location.line, prefix))
			}
		}
	}
	if len(failures) > 0 {
		t.Error(strings.Join(failures, "\n"))
	}
}

type hostLocation struct {
	file string
	line int
}

func matchHostExpectation(expectations map[string]bool, diagnostic string) bool {
	for prefix := range expectations {
		if strings.HasPrefix(diagnostic, prefix) {
			expectations[prefix] = true
			return true
		}
	}
	return false
}

func hostExpectations(t *testing.T, fixtures fs.FS) map[hostLocation]map[string]bool {
	t.Helper()
	want := make(map[hostLocation]map[string]bool)
	err := fs.WalkDir(fixtures, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		content, err := fs.ReadFile(fixtures, path)
		if err != nil {
			return err
		}
		for index, line := range strings.Split(string(content), "\n") {
			if _, expectation, found := strings.Cut(line, "// want: "); found {
				prefixes := make(map[string]bool)
				for _, prefix := range strings.Split(expectation, " || ") {
					prefixes[prefix] = false
				}
				want[hostLocation{path, index + 1}] = prefixes
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return want
}
