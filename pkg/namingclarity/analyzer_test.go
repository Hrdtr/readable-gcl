package namingclarity_test

import (
	"testing"

	"github.com/hrdtr/readable-gcl/pkg/config"
	"github.com/hrdtr/readable-gcl/pkg/namingclarity"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdataDirectory := analysistest.TestData()
	defaultConfiguration := config.DefaultConfig("pragmatic")
	namingAnalyzer := namingclarity.NewAnalyzer(defaultConfiguration.NamingClarity)

	analysistest.Run(t, testdataDirectory, namingAnalyzer, "./sample")
}

func TestConfiguredAnalyzer(t *testing.T) {
	configuration, err := config.DecodeYAML([]byte(`
preset: strict
naming_clarity:
  single_letter_params: true
  single_letter_receivers: false
  check_receiver_types: false
  banned_words: [pool, cfg]
  banned_word_exceptions:
    - word: pool
      types: [sync.Pool]
    - word: pool
      previous: cert
    - word: pool
      types: [int]
      previous: database
      next: options
    - word: cfg
  type_rules_list:
    - type: time.Time
      allowed_names: []
      allowed_suffixes: []
    - type: configured.Widget
      allowed_names: [widget]
`))
	if err != nil {
		t.Fatal(err)
	}
	directory, cleanup, err := analysistest.WriteFiles(map[string]string{"configured/configured.go": `package configured
import ("net"; "sync"; "time")
type Server struct{}
type Widget struct{}
func (s *Server) Serve() {}
func (cfg *Server) Start() {} // want "disallowed identifier term.*cfg"
func serve(x int) {} // want "single-letter argument prohibited: x"
var pool sync.Pool
var rootCertPool int
var databasePoolOptions int
var databasePool int // want "disallowed identifier term.*pool"
var wrongDatabasePoolOptions string // want "disallowed identifier term.*pool"
var workerPool int // want "disallowed identifier term.*pool"
var poolCfg sync.Pool // want "disallowed identifier term.*cfg"
var poolId sync.Pool // want "disallowed initialism casing.*Id"
var poolWrapper sync.Pool // want "poolWrapper.*should be named"
var timeValue time.Time // want "is not allowed by the naming rule for time.Time"
var widget Widget
var wrongWidget Widget // want "wrongWidget.*should be named .widget."
var peers net.Conn // want "suffixed with .Connection. or .Connections."
var connections net.Conn // want "suffixed with .Connection. or .Connections."
var clientConnections net.Conn
`})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	analysistest.Run(t, directory, namingclarity.NewAnalyzer(configuration.NamingClarity), "configured")
}

func TestGenericMatchingAnalyzer(t *testing.T) {
	configuration, err := config.DecodeYAML([]byte(`
preset: strict
naming_clarity:
  generic_type_matching: true
  banned_suffix_exceptions:
    - suffix: Int
      types: [big.Int]
  bad_initialisms_list:
    - {word: Uuid, replacement: UUID}
`))
	if err != nil {
		t.Fatal(err)
	}
	directory, cleanup, err := analysistest.WriteFiles(map[string]string{"matching/matching.go": `package matching
import "math/big"
type fileTransferHandler struct{}
type FileTransferHandler struct{}
var userId int
var userUuid int // want "disallowed initialism casing.*Uuid"
var totalBigInt big.Int
var countInt int // want "meaningless representation suffix"
var wrapped any = FileTransferHandler{}
var handler fileTransferHandler
var downloadHandlers fileTransferHandler // want "downloadHandlers.*should be named"
var downloadFileTransferHandler FileTransferHandler
var downloadHandler FileTransferHandler // want "downloadHandler.*should be named"
func mixture() {
	wrapped = FileTransferHandler{}
	for _, handler := range []fileTransferHandler{{}} { _ = handler }
	for _, wrong := range []fileTransferHandler{{}} { _ = wrong } // want "wrong.*should be named"
}
`})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	analysistest.Run(t, directory, namingclarity.NewAnalyzer(configuration.NamingClarity), "matching")
}

func TestEmptyListsClearLexicalDefaults(t *testing.T) {
	configuration, err := config.Decode(map[string]any{
		"naming_clarity": map[string]any{
			"bad_initialisms_list":     []any{},
			"banned_suffix_exceptions": []any{},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	directory, cleanup, err := analysistest.WriteFiles(map[string]string{"empty/empty.go": `package empty
import "math/big"
var userId int
var totalBigInt big.Int // want "meaningless representation suffix"
`})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	analysistest.Run(t, directory, namingclarity.NewAnalyzer(configuration.NamingClarity), "empty")
}

func TestAnalyzerDisabled(t *testing.T) {
	disabledAnalyzer := namingclarity.NewAnalyzer(config.NamingClarityConfig{Enabled: false})
	result, err := disabledAnalyzer.Run(nil)
	if err != nil || result != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", result, err)
	}
}
