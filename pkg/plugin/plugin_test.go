package plugin_test

import (
	"encoding/json"
	"errors"
	"flag"
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

	for _, form := range []string{"list", "normalized-map", "empty-lists", "receiver-type-only", "receiver-length-only"} {
		t.Run(form, func(t *testing.T) {
			settings := string(example)
			switch form {
			case "normalized-map":
				settings = strings.Replace(settings, "type_rules_list:\n              - type: time.Time", "type_rules:\n              \"time.Time\":", 1)
			case "empty-lists":
				settings = strings.Replace(settings, "[time, now, start, end]", "[]", 1)
				settings = strings.Replace(settings, "[Time, Start, End, At, Until, Since]", "[]", 1)
				settings = strings.Replace(settings, "[cfg, mgr, req, res, resp, val, vals, obj, usr, pool]", "[]", 1)
			case "receiver-type-only":
				settings = strings.Replace(settings, "check_receiver_types: false", "check_receiver_types: true", 1)
			case "receiver-length-only":
				settings = strings.Replace(settings, "single_letter_receivers: false", "single_letter_receivers: true", 1)
				settings = strings.Replace(settings, "single_letter_params: true", "single_letter_params: false", 1)
			}
			directory := t.TempDir()
			for path, content := range map[string]string{
				".golangci.yml": settings,
				"go.mod":        "module example.test/fixture\n\ngo 1.27.1\n",
				"fixture.go": `package fixture
import ("sync"; "time")
type Instant struct{}
type Server struct{}
func (s *Server) Serve() {}
func (cfg *Server) Start() {}
func serve(x int) {}
var now time.Time
var createdAt time.Time
var lockedUntil time.Time
var banana time.Time
var createdAtValue time.Time
func demo() { var now Instant; _ = now }
var pool sync.Pool
var rootCertPool int
var databasePoolOptions int
var workerPool int
var poolCfg sync.Pool
var poolId sync.Pool
var poolWrapper sync.Pool
`,
			} {
				if err := os.WriteFile(filepath.Join(directory, path), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			command := exec.CommandContext(t.Context(), binary, "run", "--output.json.path=stdout", "--output.text.path=", "--show-stats=false", "--uniq-by-line=false", "./...")
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
			type location struct {
				file         string
				line, column int
			}
			want := map[location]string{
				{"fixture.go", 6, 7}:   `disallowed identifier term "cfg"`,
				{"fixture.go", 7, 12}:  `single-letter argument prohibited: x`,
				{"fixture.go", 11, 5}:  `"banana" should be`,
				{"fixture.go", 12, 5}:  `"createdAtValue" should be`,
				{"fixture.go", 13, 19}: `"now" should be`,
				{"fixture.go", 17, 5}:  `disallowed identifier term "pool"`,
				{"fixture.go", 18, 5}:  `disallowed identifier term "cfg"`,
				{"fixture.go", 19, 5}:  `disallowed initialism casing "Id"`,
				{"fixture.go", 20, 5}:  `"poolWrapper" should be`,
			}
			switch form {
			case "normalized-map":
				// This host lowercases the mapping key and the preset rule remains in effect.
				want[location{"fixture.go", 8, 5}] = `"now" should be named "time"`
				want[location{"fixture.go", 9, 5}] = `"createdAt" should be named "time"`
				want[location{"fixture.go", 10, 5}] = `"lockedUntil" should be named "time"`
			case "empty-lists":
				for line := 8; line <= 12; line++ {
					want[location{"fixture.go", line, 5}] = `is not allowed by the naming rule for time.Time`
				}
				delete(want, location{"fixture.go", 6, 7})
				delete(want, location{"fixture.go", 17, 5})
				want[location{"fixture.go", 18, 5}] = `"poolCfg" should be`
			case "receiver-type-only":
				want[location{"fixture.go", 5, 7}] = `"s" should be named "server"`
			case "receiver-length-only":
				want[location{"fixture.go", 5, 7}] = `single-letter receiver prohibited: s`
				delete(want, location{"fixture.go", 7, 12})
			}
			for _, issue := range result.Issues {
				position := location{filepath.ToSlash(issue.Pos.Filename), issue.Pos.Line, issue.Pos.Column}
				expected, exists := want[position]
				if issue.FromLinter != "readablegcl" || !exists || !strings.Contains(issue.Text, expected) {
					t.Errorf("unexpected finding at %+v: %s (%s)", position, issue.Text, issue.FromLinter)
					continue
				}
				delete(want, position)
			}
			for position, message := range want {
				t.Errorf("missing finding at %+v: %s", position, message)
			}
		})
	}
}
