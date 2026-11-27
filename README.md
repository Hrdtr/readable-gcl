# readable-gcl

`readable-gcl` is a configurable Go linter and **GolangCI-Lint v2 module plugin** that helps teams maintain clear, self-documenting naming conventions and test structures across Go projects.

> **Note on Naming Style**: Idiomatic Go often favors concise naming in local scopes (`r`, `w`, `u`), which works well for many developers. `readable-gcl` is an opt-in tool designed for teams and codebases that prefer more explicit, self-documenting variable names, consistent initialisms, and clear test conventions. Every check is modular and configurable.

---

## Motivation

In growing codebases with rich domain logic, multiple engineers, and frequent code reviews, descriptive naming can make code easier to navigate:

- **Self-Documenting Code**: Explicit names like `user`, `server`, and `configuration` reduce mental translation compared to single-letter or abbreviated identifiers.
- **Consistent Initialisms**: Keeps acronyms aligned with standard Go conventions (`userID` and `serverURL` rather than `userId` and `serverUrl`).
- **No Redundant Type Suffixes**: Discourages encoding primitive types into variable names (e.g. `countInt`, `validBool`) since the Go compiler already enforces types.
- **Test Clarity**: Promotes clean table-driven test isolation using `t.Run` and optional test-file pairing rules.

---

## Features

- **Discourages Ambiguous Abbreviations**: Flags common shorthand abbreviations like `userCfg`, `authReq`, `userMgr`, and `usr`, encouraging clear domain names (`config`, `request`, `manager`, `user`).
- **Flags Redundant Type Suffixes**: Discourages Hungarian notation in variable names (e.g. `countInt`, `validBool`, `nameStr`), while exempting domain types like `big.Int`.
- **Enforces Go Initialisms**: Detects non-standard PascalCase initialisms like `userId` and `serverUrl`, recommending standard forms (`userID`, `serverURL`).
- **Requires Multi-Letter Parameters & Receivers**: Encourages descriptive names for function parameters and method receivers (e.g. `server *Server` rather than `s *Server`), exempting `_` and `*testing.T`.
- **Table-Driven Test Subtest Verification**: Flags table-driven test loops that execute assertions without `t.Run`, helping ensure test failures remain isolated and identifiable.
- **Optional File Pairing**: Optionally verifies that test files pair with an existing implementation file or match their package directory name.
- **Configurable Presets**:
  - `pragmatic` (default): Focuses on readability, cleaning ambiguous abbreviations, Hungarian notation, initialisms, and table subtests.
  - `strict`: Enforces standard PascalCase test naming pattern (`^Test[A-Z][a-zA-Z0-9]+$`) and strict type-to-variable name matching across all types.
- **Zero Hardcoded Project Dependencies**: 100% configurable for any Go codebase.

---

## Installation & Usage

### 1. With GolangCI-Lint v2 (Module Plugin)

Create `.custom-gcl.yml` in your project root:

```yaml
version: v2.13.2

plugins:
  - module: 'github.com/hrdtr/readable-gcl'
    import: 'github.com/hrdtr/readable-gcl/pkg/plugin'
    path: . # or version: v0.1.0
```

Build your custom `golangci-lint` binary:

```bash
golangci-lint custom
```

This generates `./custom-gcl`. Add `readablegcl` to your `.golangci.yml` (v2 syntax):

```yaml
version: "2"

linters:
  default: none
  enable:
    - readablegcl
  settings:
    custom:
      readablegcl:
        type: "module"
        description: "Readable Go naming conventions and test clarity"
        settings:
          preset: "pragmatic" # or "strict"
```

Run the linter:

```bash
./custom-gcl run
```

---

### 2. As a Standalone CLI / `go vet` Tool

Install the standalone binary:

```bash
go install github.com/hrdtr/readable-gcl/cmd/readable-gcl@latest
```

Run directly:

```bash
readable-gcl ./...
```

Or via `go vet`:

```bash
go vet -vettool=$(which readable-gcl) ./...
```

---

## Configuration Reference

```yaml
version: "2"

linters:
  default: none
  enable:
    - readablegcl
  settings:
    custom:
      readablegcl:
        type: "module"
        description: "Readable Go naming conventions and test clarity"
        settings:
          preset: strict

          naming_clarity:
            single_letter_params: true
            single_letter_receivers: false
            check_receiver_types: false
            # Lists replace the preset's list; include defaults you want to keep.
            banned_words: [cfg, mgr, req, res, resp, val, vals, obj, usr, pool]
            banned_word_exceptions:
              - word: pool
                types: [sync.Pool]
              - word: pool
                previous: cert
              - word: pool
                previous: database
                next: options
            type_rules_list:
              - type: time.Time
                allowed_names: [time, now, start, end]
                allowed_suffixes: [Time, Start, End, At, Until, Since]

          test_convention:
            pattern: '^Test[A-Z][a-zA-Z0-9]+$'
            require_folder_prefix: true
            require_tier_suffix: true
            enforce_table_subtests: true
            require_matching_source_file: true
            skipped_dirs: [.git, vendor, node_modules, bin, dist, .cache, ignored]
```

---

### Type rules and overrides

See [`.golangci.example.yml`](.golangci.example.yml) for a complete GolangCI-Lint configuration. The GolangCI-Lint loader (verified with v2.13.2 and v2.15.0) lowercases map keys inside plugin settings (even quoted `"time.Time"`). Set `type_rules_list` in `.golangci.yml` so the Go type is a **value** and keeps its case. The original `type_rules` map remains supported in standalone YAML or when passed to `config.Decode` with its keys intact.

- Each entry replaces the **whole rule for that type**; unmentioned types keep their preset rules. List entries take priority over map entries for the same type, and the last list entry wins if a type is repeated. `type_rules_list: []` supplies no overrides and leaves preset rules in place.
- Within a rule, `allowed_names` and `canonical_name` are exact, case-sensitive alternatives. `allowed_suffixes` and `expected_tail` are literal, case-sensitive suffix alternatives. `allow_plural: true` also permits each suffix followed by `s`. Empty lists add no alternatives; a rule with no names or suffixes rejects every name subject to that rule unless `allow_any_domain_name: true` is set or the existing `ErrX` sentinel convention applies to `error`.
- Other slice settings (such as `banned_words`, `banned_suffixes`, `banned_compounds`, `banned_word_exceptions`, `exempt_identifiers`, and `skipped_dirs`) replace the preset list when present; `[]` clears it. Omit the setting to inherit the preset.

### Banned-word exceptions and receiver scope

`banned_word_exceptions` entries use `word` plus at least one condition: `types` (exact resolved type names such as `sync.Pool`), `previous`, or `next` (immediate neighboring identifier components, matched ignoring case). All supplied conditions in one entry must match; separate entries are alternatives. In the example, `pool` is permitted for `sync.Pool`, after `cert` in `rootCertPool`, or between `database` and `options` in `databasePoolOptions`. An unrelated `workerPool` still violates the rule. The exception only removes the finding for that banned component; it does not skip other words, initialisms, or type naming.

`single_letter_params` and `single_letter_receivers` independently override the legacy `single_letter_args` setting. `check_receiver_types: false` turns off receiver type naming without disabling suffix, initialism, or banned-word checks on receivers. In strict mode, set both receiver options to `false` to accept `func (s *Server) Serve()` while still rejecting `func serve(x int)`.

### Testing the module plugin

Follow GolangCI-Lint’s [Module Plugin System](https://golangci-lint.run/docs/plugins/module-plugins/) workflow. From the repository root:

```sh
golangci-lint custom --destination bin
go test ./pkg/plugin -run TestGolangCIHost -count=1 -args -golangci-integration
```

The integration test loads `.golangci.yml` through the custom binary and checks findings by file, line, column, and rule. This is a regression test for the plugin configuration path, not a comparison against a separate analyzer.

## Presets Comparison

| Rule Feature | `pragmatic` (Default) | `strict` |
| --- | --- | --- |
| Multi-letter parameter & receiver requirement | Yes (except `_`, `t`) | Yes (except `_`, `t`) |
| Banned Hungarian suffixes (`*Bool`, `*Int`, etc.) | Yes | Yes |
| Bad initialisms (`userId` $\rightarrow$ `userID`) | Yes | Yes |
| Ambiguous abbreviations (`cfg`, `req`, `mgr`, etc.) | Yes | Yes |
| Table-test `t.Run` verification | Yes | Yes |
| Standard library type rules (`ctx`, `err`) | Yes | Yes |
| Variable names must end with concrete type | No | Yes |
| Test function PascalCase pattern enforcement | No | Yes (`^Test[A-Z][a-zA-Z0-9]+$`) |
| Test file paired with `.go` or directory name | Optional | Optional (configurable) |
| Test folder prefix enforcement | Optional | Optional (configurable) |
| Test tier suffix enforcement (`*Unit`, `*E2E`) | Optional | Optional (configurable) |

---

## License

MIT License.
