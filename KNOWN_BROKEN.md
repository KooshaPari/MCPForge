# Known Broken Integration Tests

This file tracks integration tests in MCPForge that fail because of bugs in
the upstream language server(s) we wrap. The failures are not caused by
MCPForge and cannot be fixed in this repository without forking or patching
the upstream language server — which is explicitly out of scope per
`AGENTS.md`.

Each entry lists the failing test, the observed symptom, the upstream issue
link, and the date last observed. A matching sentinel test in
[`tests/integration_stability_test.go`](tests/integration_stability_test.go)
records the breakage in `go test ./...` output with the same upstream
links.

When upstream fixes a tracked issue, both this file and the matching gate
test should be removed in the same PR that re-enables the underlying
integration test.

Upstream repo: <https://github.com/isaacphi/mcp-language-server>

---

## 1. Rust — `TestRenameSymbol`

- **Test path**: `integrationtests/tests/rust/rename_symbol.TestRenameSymbol`
  - Subtest `SuccessfulRename`
  - Subtest `SymbolNotFound`
- **Last observed**: 2026-06-08
- **Observed duration**: ~21.4 s (the entire `SuccessfulRename` + `SymbolNotFound` run)
- **Symptom (verified locally)**:
  `RenameSymbol failed: failed to rename symbol: request failed: No references found at position (code: -32602)`.
  The test renames `SHARED_CONSTANT` at its definition in `src/types.rs`
  and asserts the rename propagates to `src/consumer.rs`. The rename
  succeeds in the definition file but the cross-file occurrence is left
  untouched, and the LSP eventually returns the `-32602` error.
- **Upstream issues**:
  - [#104 — BUG: renaming is limited to a single file](https://github.com/isaacphi/mcp-language-server/issues/104)
    (root cause: workspace-edit client capabilities are not advertised, so the
    language server falls back to single-file rename.)
  - [#121 — Symbols are never found](https://github.com/isaacphi/mcp-language-server/issues/121)
    (related symbol-lookup failure mode observed on Windows; same family
    of issues.)
- **Gate sentinel**:
  `tests/integration_stability_test.go::TestGate_RustRenameSymbol_KnownBrokenUpstream`
- **Affected snapshot**: `integrationtests/snapshots/rust/rename_symbol/{successful,not_found}.snap`
  (snapshots are not updated while the test is failing; no action needed.)

---

## 2. TypeScript — `TestDiagnostics / FileWithError`

- **Test path**: `integrationtests/tests/typescript/diagnostics.TestDiagnostics/FileWithError`
- **Last observed**: 2026-06-08
- **Symptom**:
  The test writes a TypeScript file with a deliberate type error
  (`return x; // number is not assignable to string`) and calls
  `tools.GetDiagnosticsForFile`, which issues the LSP 3.17
  `textDocument/diagnostic` request via
  `internal/lsp/methods.go:206`. The typescript-language-server returns
  `{"code":-32601,"message":"Unhandled method textDocument/diagnostic"}`
  and the assertion
  `Expected type error but got: <error message>` fails.
- **Why other subtests pass**:
  `CleanFile` and `FileDependency` rely on the cached
  `textDocument/publishDiagnostics` notification, which the server does
  push. Only `FileWithError` requires the `textDocument/diagnostic`
  request itself.
- **Upstream issues**:
  - [#60 — Unreliable diagnostics for typescript-server](https://github.com/isaacphi/mcp-language-server/issues/60)
    (root cause: typescript-language-server does not implement the
    `textDocument/diagnostic` LSP 3.17 method; linked PR
    [#131](https://github.com/isaacphi/mcp-language-server/pull/131)
    tracks the fix.)
  - [#121 — Symbols are never found](https://github.com/isaacphi/mcp-language-server/issues/121)
    (related symbol-lookup failure mode.)
- **Gate sentinel**:
  `tests/integration_stability_test.go::TestGate_TypeScriptDiagnostics_FileWithError_KnownBrokenUpstream`
- **Affected snapshot**: `integrationtests/snapshots/typescript/diagnostics/type-error.snap`
  (snapshot is not updated while the test is failing; no action needed.)

---

## Process

1. **Adding** a new known-broken test: append an entry above with the test
   path, symptom, upstream issue, last-observed date, and add a matching
   `t.Skip()`-only sentinel in `tests/integration_stability_test.go`.
2. **Removing** a known-broken test (upstream fix landed): delete the
   entry here and the matching sentinel in
   `tests/integration_stability_test.go` in the same PR that re-enables
   the integration test. Verify the integration test now passes and the
   `UPDATE_SNAPSHOTS=true` flag is used if a new snapshot is required.
3. **Disambiguating upstream vs MCPForge regressions**: if a previously
   green integration test starts failing, first check this file — only
   re-open the upstream issue if the failure is a new symptom.
