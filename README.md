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
          preset: "pragmatic" # "pragmatic" (default) or "strict"

          naming_clarity:
            enabled: true
            single_letter_args: true
            banned_suffixes:
              - "Boolean"
              - "Bool"
              - "Str"
              - "Int"
            bad_initialisms:
              Id: "ID"
              Url: "URL"
              Ttl: "TTL"
              Jwt: "JWT"
              Sid: "SID"
            banned_words:
              - "cfg"
              - "mgr"
              - "req"
              - "res"
              - "resp"
              - "val"
              - "vals"
              - "obj"
              - "usr"
            banned_compounds: []
            enforce_type_tail: false
            type_rules:
              "context.Context":
                canonical_name: "ctx"
                expected_tail: "Ctx"
              "error":
                canonical_name: "err"
                expected_tail: "Err"
            exempt_identifiers:
              - "_"
              - "ok"
              - "t"

          test_convention:
            enabled: true
            enforce_table_subtests: true
            pattern: ""
            require_folder_prefix: false
            require_tier_suffix: false
            require_matching_source_file: false # requires file_name_test.go to match file_name.go or directory name
            skipped_dirs:
              - ".git"
              - "vendor"
              - "node_modules"
              - "bin"
              - "dist"
              - ".cache"
```

---

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
