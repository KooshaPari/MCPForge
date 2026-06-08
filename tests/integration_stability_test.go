// Package smoke_test is the test-only package for repo-root tests in
// MCPForge. It contains both the basic smoke test (smoke_test.go) and the
// integration stability gate (integration_stability_test.go).
//
// File purpose:
// This file contains sentinel tests that track known-broken upstream tests
// in the integration suite. Each test intentionally calls t.Skip() in its
// body so the skip is recorded in `go test ./...` output with a clear
// reason and a link to the upstream issue. The tests do not exercise the
// broken behavior; they document the known breakage so it cannot regress
// silently when the upstream issue is eventually fixed.
//
// When upstream resolves a tracked issue, the corresponding gate test in
// this file should be removed in the same PR that re-enables the underlying
// integration test. The repo-root KNOWN_BROKEN.md manifest should be
// updated at the same time.
//
// Tracked upstream repo: https://github.com/isaacphi/mcp-language-server
package smoke_test

import "testing"

// TestGate_RustRenameSymbol_KnownBrokenUpstream tracks the failing rust
// rename_symbol integration test. Upstream mcp-language-server#104 reports
// that rust-analyzer rename is limited to a single file: the rename succeeds
// in the file containing the definition but is not propagated to consumers.
// The integration test asserts that renaming SHARED_CONSTANT in src/types.rs
// also rewrites occurrences in src/consumer.rs, and fails on the latter.
//
// Upstream issues:
//   - https://github.com/isaacphi/mcp-language-server/issues/104
//     "BUG: renaming is limited to a single file"
//   - https://github.com/isaacphi/mcp-language-server/issues/121
//     "Symbols are never found" (related symbol-lookup failure mode)
//
// Affected:
//   - integrationtests/tests/rust/rename_symbol.TestRenameSymbol
//   - subtest "SuccessfulRename"
//   - subtest "SymbolNotFound"
//
// Last observed: 2026-06-08 (FAIL in ~21.4s; error: "No references found at
// position (code: -32602)").
func TestGate_RustRenameSymbol_KnownBrokenUpstream(t *testing.T) {
	t.Skip("KNOWN BROKEN — upstream mcp-language-server#104. " +
		"Rename is limited to a single file: rust-analyzer renames in the " +
		"file containing the definition but does not propagate to consumer " +
		"files. The integration test asserts cross-file renaming and fails. " +
		"See https://github.com/isaacphi/mcp-language-server/issues/104 " +
		"(related: https://github.com/isaacphi/mcp-language-server/issues/121). " +
		"Affects integrationtests/tests/rust/rename_symbol.TestRenameSymbol " +
		"subtests SuccessfulRename and SymbolNotFound.")
}

// TestGate_TypeScriptDiagnostics_FileWithError_KnownBrokenUpstream tracks the
// failing typescript FileWithError integration test. Upstream
// mcp-language-server#60 reports that typescript-language-server does not
// implement the LSP `textDocument/diagnostic` request (LSP 3.17). The server
// responds with JSON-RPC code -32601 "Unhandled method textDocument/diagnostic".
// tools.GetDiagnosticsForFile issues that call (internal/lsp/methods.go:206)
// and fails on this subtest, even though the cached `textDocument/publishDiagnostics`
// notification path still works (so CleanFile and FileDependency subtests pass).
//
// Upstream issues:
//   - https://github.com/isaacphi/mcp-language-server/issues/60
//     "Unreliable diagnostics for typescript-server"
//   - https://github.com/isaacphi/mcp-language-server/issues/121
//     "Symbols are never found" (related symbol-lookup failure mode)
//
// Affected:
//   - integrationtests/tests/typescript/diagnostics.TestDiagnostics/FileWithError
//
// Last observed: 2026-06-08 (FAIL with "Unhandled method textDocument/diagnostic
// (code: -32601)").
func TestGate_TypeScriptDiagnostics_FileWithError_KnownBrokenUpstream(t *testing.T) {
	t.Skip("KNOWN BROKEN — upstream mcp-language-server#60. " +
		"typescript-language-server does not implement the LSP " +
		"`textDocument/diagnostic` method (LSP 3.17) and returns " +
		"-32601 \"Unhandled method textDocument/diagnostic\". " +
		"The diagnostic cache falls back to the legacy " +
		"`textDocument/publishDiagnostics` notification, so the CleanFile " +
		"and FileDependency subtests pass; FileWithError asserts on the " +
		"diagnostic content and fails. " +
		"See https://github.com/isaacphi/mcp-language-server/issues/60 " +
		"(related: https://github.com/isaacphi/mcp-language-server/issues/121). " +
		"Affects integrationtests/tests/typescript/diagnostics.TestDiagnostics/FileWithError.")
}
