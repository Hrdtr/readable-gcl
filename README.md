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
            enforce_type_tail: true
            generic_type_matching: true
            single_letter_params: true
            single_letter_receivers: true
            check_receiver_types: true
            # Lists replace the corresponding preset settings.
            banned_suffixes: [Boolean, Bool, Str, Int]
            banned_suffix_exceptions:
              - suffix: Int
                types: [big.Int]
            bad_initialisms_list:
              - {word: Id, replacement: ID}
              - {word: Url, replacement: URL}
              - {word: Uuid, replacement: UUID}
              - {word: Ttl, replacement: TTL}
              - {word: Jwt, replacement: JWT}
              - {word: Sid, replacement: SID}
            banned_words: [cfg, mgr, req, res, resp, val, vals, obj, usr, pool]
            banned_compounds: [tempobj]
            banned_word_exceptions:
              - word: pool
                types: [sync.Pool]
              - word: pool
                previous: cert
              - word: pool
                types: [int]
                previous: database
                next: options
            type_rules_list:
              - type: time.Time
                allowed_names: [time, now, start, end]
                allowed_suffixes: [Time, Start, End, At, Until, Since]
              - type: fixture.Subscription
                allowed_names: [subscription, subscriptionAlias]
                # Enumerate each form without adding another plural via allow_plural.
                allowed_suffixes: [Link, Links, Box, Boxes]
              - type: fixture.DomainResource
                allow_any_domain_name: true
              - type: explicit.fileTransferHandler
                allowed_names: [transfer]
              - type: error
                allowed_names: [err]
                allowed_suffixes: [Err]

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

- Each entry replaces the **whole rule for that type**; unmentioned types keep their preset rules. List entries take priority over map entries for the same type, and the last list entry wins if a type is repeated. `type_rules_list: []` supplies no overrides and leaves preset rules in place. A rejection by a type rule never falls back to generic matching.
- Within a rule, `allowed_names` and `canonical_name` are exact, case-sensitive alternatives. `allowed_suffixes` and `expected_tail` are literal, case-sensitive suffix alternatives. `allow_plural: true` also permits each suffix followed by `s`. To support a precise set of `s` and `es` forms, enumerate the singulars and plurals and leave `allow_plural` false so that a second `s` is not added. Empty lists add no alternatives; a rule with no names or suffixes rejects every name subject to that rule unless `allow_any_domain_name: true` is set or the existing `ErrX` sentinel convention applies to `error`.
- With `enforce_type_tail: true`, `generic_type_matching: true` applies the opt-in policy to unlisted types. It accepts case-insensitive exact names and full type suffixes, including after removing one trailing lowercase `s`. For unexported types, shorter suffixes starting at uppercase letters also match exactly (ignoring case) or as case-sensitive suffixes, without the plural allowance. Omit `generic_type_matching` or set it to `false` for the legacy matcher. `enforce_type_tail: false` disables unlisted-type enforcement without disabling explicit type rules.
- Slice settings such as `banned_words`, `banned_suffixes`, `banned_compounds`, `banned_word_exceptions`, `exempt_identifiers`, and `skipped_dirs` replace the preset list when present; `[]` clears it. Omit the setting to inherit the preset. An explicit `false` disables the corresponding boolean feature, even if the preset enables it; `pattern: ""` disables just the test-name pattern.

### Lexical rules and receiver checks

Use `bad_initialisms_list` in GolangCI-Lint so case-sensitive components like `Id` and `Url` survive the loader. When present, the list replaces the entire preset or legacy `bad_initialisms` map; last entry for a component wins, and `[]` disables the check. The map remains supported when keys are preserved.

`banned_suffix_exceptions` takes a literal banned `suffix` and a `types` list. The exception applies only to that suffix on those resolved named types (including pointers and aliases), not to other banned suffixes or other checks. Omit the list to retain the legacy `BigInt`/`BigInteger` exceptions; set it to `[]` to apply `banned_suffixes` without exceptions.

`banned_word_exceptions` entries use `word` plus at least one condition: `types` (exact resolved type names such as `sync.Pool`), `previous`, or `next` (immediate neighboring identifier components, matched ignoring case). All supplied conditions in one entry must match; separate entries are alternatives. In the example, `pool` is permitted for `sync.Pool`, after `cert` in `rootCertPool`, or between `database` and `options` when the identifier is an `int`. An unrelated `workerPool` still violates the rule. The exception only removes the finding for that banned component; it does not skip other words, initialisms, or type naming.

`single_letter_params` and `single_letter_receivers` independently override the legacy `single_letter_args` setting. `check_receiver_types: false` turns off receiver type naming without disabling suffix, initialism, or banned-word checks on receivers. The example keeps receiver, parameter, and named-result checks enabled.

### Testing the module plugin

Follow GolangCI-Lint’s [Module Plugin System](https://golangci-lint.run/docs/plugins/module-plugins/) workflow. From the repository root:

```sh
golangci-lint custom --destination bin
go test ./pkg/plugin -run TestGolangCIHost -count=1 -args -golangci-integration
```

The integration test loads the example unchanged as `.golangci.yml` through the custom binary. Every `// want:` marker in the fixtures names the required rule and diagnostic; unmarked locations must produce no finding.

### Rule-to-setting checklist

| Rule | Setting in the example | Passing host fixture: accepted → rejected |
| --- | --- | --- |
| Generic unexported names | `enforce_type_tail`, `generic_type_matching` | [`fixture/naming.go`](pkg/namingclarity/testdata/contract/fixture/naming.go): `handler`, `downloadHandler`, `cachedfiletransferhandler`, `cachedFileTransferHandlers` → `downloadHandlers`, `downloadhandler`, `banana` |
| Generic exported names | Same settings, no per-type override | `fixture/naming.go`: `fileTransferHandler`, `downloadFileTransferHandler` → `handler`, `downloadHandler`, `cachedfiletransferhandlerss` |
| Aliases, pointers, and unresolved types | Resolved by the analyzer | `fixture/naming.go`: `downloadTransferHandler` through an alias to a double pointer; `payload` of type `any` stays lexical → `userId` still fails lexical casing |
| Exact rule overrides generic matching | `type_rules_list` entry for `explicit.fileTransferHandler` | [`explicit/naming.go`](pkg/namingclarity/testdata/contract/explicit/naming.go): `transfer` → `handler` |
| Literal names, suffixes, and plurals | `allowed_names`, `allowed_suffixes`, including `Links` and `Boxes` | `fixture/naming.go`: `subscriptionAlias`, `backupLinks`, `priorityBoxes` → `SubscriptionAlias`, `prioritybox`, `backupLinkss`, `priorityBoxess` |
| Any domain name | `allow_any_domain_name: true` | `fixture/naming.go`: `inventory` → `inventoryCfg` still fails the word rule |
| Mixed-case type key | `type_rules_list` with `type: time.Time` | `fixture/naming.go`: `now`, `createdAt`, `lockedUntil` → `banana`, `createdAtValue`, and `now` on an unrelated type |
| Representation suffixes | `banned_suffixes`, `banned_suffix_exceptions` | `fixture/naming.go`: `totalBigInt` on `big.Int` → `countInt` and `totalBigInt` on `int` |
| Initialism casing | `bad_initialisms_list` | `fixture/naming.go`: `userID`, `serverURL`, `userUUID` → `userId`, `serverUrl`, `userUuid` |
| Banned words and conditional exceptions | `banned_words`, `banned_word_exceptions` | `fixture/naming.go`: `pool`, `rootCertPool`, `databasePoolOptions` → `workerPool`, `databasePool`, `otherDatabasePoolOptions`; `poolCfg`, `poolId`, and `poolWrapper` still fail their own rules |
| Compound substrings | `banned_compounds` | `fixture/naming.go`: `temporaryObject` → `mytempobject`, `poolTempobj` |
| Error names | Explicit `error` rule and the sentinel convention | `fixture/naming.go`: `err`, `readErr`, `ErrNotFound`, `ErrÉchec` → `error`, `readError`, `ErrnotFound`, `Erré` |
| Receiver, parameter, and named-result naming | `single_letter_receivers`, `check_receiver_types`, `single_letter_params`, type rules | `fixture/naming.go`: `server`, `count`, `fileTransferHandler` → `s`, `wrongReceiver`, `x`, `wrongParameter`, `wrongResult` |
| Test-name pattern and folder prefix | `pattern`, `require_folder_prefix` | [`suite/suite_test.go`](pkg/testconvention/testdata/contract/suite/suite_test.go): `TestSuiteWorksUnit` → `TestSuite_badUnit`, `TestOtherWorksUnit` |
| Test tier | `require_tier_suffix` | [`suite`](pkg/testconvention/testdata/contract/suite): matching unit, integration, and E2E suffixes → the wrong tier in each file |
| Table subtests | `enforce_table_subtests` | `suite/suite_test.go`: loop with `t.Run` → loop asserting without `t.Run` |
| File pairing and skipped directories | `require_matching_source_file`, `skipped_dirs` | `suite/suite_test.go` is paired by folder; [`suite/orphan_test.go`](pkg/testconvention/testdata/contract/suite/orphan_test.go) is rejected; [`ignored/orphan_test.go`](pkg/testconvention/testdata/contract/ignored/orphan_test.go) skips test conventions but still reports `userId` |

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
