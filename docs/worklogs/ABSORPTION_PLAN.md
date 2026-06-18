# Absorption Plan — MCPForge → target repos

<!-- Plans document cross-repo absorption moves: file mapping, conflicts, push strategy. -->

### 2026-06-18 | PLAN | Absorption plan: MCPForge → McpKit, PhenoMCPServers, Agentora

**Context:** MCPForge needs to be decomposed and pushed to three canonical homes.
**Decision:** Three separate PRs, sequenced McpKit → PhenoMCPServers → Agentora.
**Status:** Planning only. No `git push` or `gh pr create` executed.

---

## Source inventory (MCPForge @ this repo)

| Source path | Size class | Notes |
|-------------|-----------|-------|
| `internal/lsp/` (8 .go files) | ~85 KB | client, transport, methods, protocol, server-request-handlers, typescript, detect-language, tests |
| `internal/protocol/` (7 .go files + LICENSE/README) | ~395 KB | LSP types/interfaces (tsjson, tsprotocol dominant) |
| `internal/tools/` (11 .go files) | ~45 KB | 6 tools: definition, references, diagnostics, hover, rename-symbol, edit_file (plus codelens/utilities) |
| `internal/logging/`, `internal/utilities/`, `internal/watcher/` | small | shared infra, *not in scope* for absorption (cross-cutting deps) |
| `cmd/generate/` (10 .go files) | ~70 KB | LSP types generator |
| `main.go` + `tools.go` | ~18 KB | server entry point + CLI tool wiring |
| `go.mod` / `go.sum` | small | `module github.com/isaacphi/mcp-language-server` |
| `integrationtests/` | medium | language-specific test fixtures (clangd, go, python, rust, typescript) |
| `FUNCTIONAL_REQUIREMENTS.md` | 1 KB | 7 stub FRs |
| `wip/mcp-forge-lsp-rust/` | (gone) | scaffolding already extracted into RESEARCH.md; no live files |

---

## Target 1 — `KooshaPari/McpKit` (MCP framework SDK)

- **Default branch:** `main` | **Language:** Go | **Last push:** 2026-06-17
- **Existing surface relevant to absorption:**
  - `go/go.work` references `./pheno-mcp-client`, `./pheno-mcp-server`, `./pheno-mcp-types` — **Go workspace is empty** otherwise (scaffold only per `AGENTS.md`).
  - `rust/mcp-forge/` — **existing fork of isaacphi/mcp-language-server** (different repo, same upstream class). Module path `github.com/isaacphi/mcp-language-server`, go.mod SHA `b875a3fcad`. **Not our MCPForge.** Has its own `internal/`, `cmd/generate/`, `integrationtests/`, `main.go`, `tools.go` mirroring the *upstream* layout but pinned to isaacphi.
  - `rust/Cargo.toml` workspace: `phenotype-mcp-{core,framework,asset,fast,fast-macros}` — does **not** include `mcp-forge` as a Rust member (Go code lives in a Rust workspace directory — historical oddity).
  - `python/pheno-mcp/` is a **git submodule** pointing to `KooshaPari/PhenoMCP`.
  - `registry.yaml` SSOT is empty/scaffolding only.

### Push set

| MCPForge source | Target path in McpKit | Action |
|---|---|---|
| `internal/lsp/*` (8 files) | `rust/mcp-forge/internal/lsp/` | **Add** (alongside isaacphi fork's files; rename file collisions only if SHAs differ) |
| `internal/protocol/*` (7 files + LICENSE/README) | `rust/mcp-forge/internal/protocol/` | **Add** — note existing dir is `isaacphi/protocol/...` not MCPForge |
| `FUNCTIONAL_REQUIREMENTS.md` | `docs/mcpforge/FR.md` | **Add** (new path under `docs/` to avoid clashing with `docs/sota`, `docs/research`, etc.) |
| (informational) `docs/worklogs/RESEARCH.md` excerpt | `docs/mcpforge/RESEARCH-RUST-LSP.md` | **Add** pointer doc |

### Conflicts to resolve before push

1. **`rust/mcp-forge/` is occupied** by the isaacphi fork (367 files). Two options:
   - **(A)** Push our content into a new sub-path `rust/mcp-forge-upstream-mcpforge/` (zero merge risk; ugly).
   - **(B)** Push into `rust/mcp-forge/` and accept that this dir will contain two parallel Go modules. Will break `go.mod` resolution unless the new files are namespaced.
   - **(C) Preferred:** Create `rust/mcp-forge-bridge/` (or `go/mcp-forge-bridge/`) for the new absorption and rename the existing dir later in a separate PR. Document this in the PR description.
2. **`go/go.work`** references three non-existent module dirs (`pheno-mcp-client`, etc.). Don't touch in this PR — out of scope; flag as follow-up.
3. **License mismatch risk:** MCPForge ships `internal/protocol/LICENSE` (BSD-3) from upstream Microsoft tsprotocol. McpKit is dual Apache-2.0/MIT. New files inherit the BSD-3 attribution — keep the LICENSE file verbatim, do **not** relicense.

### Strategy

`git subtree add --prefix=rust/mcp-forge-bridge https://github.com/KooshaPari/MCPForge.git main --squash` is the cleanest option if the maintainer agrees to receive via subtree. **However**, since the receiving side already has its own `rust/mcp-forge/` subtree, prefer **manual copy via a feature branch + PR**:

```bash
# In a worktree at repos/McpKit-wtrees/absorb-mcpforge
git fetch origin
git switch -c absorb/mcpforge-bridge origin/main

# Copy files (use rsync to preserve modes, then git add)
rsync -av --delete \
  $MCPFORGE/internal/lsp/    rust/mcp-forge-bridge/internal/lsp/
rsync -av --delete \
  $MCPFORGE/internal/protocol/ rust/mcp-forge-bridge/internal/protocol/
mkdir -p docs/mcpforge
cp $MCPFORGE/FUNCTIONAL_REQUIREMENTS.md docs/mcpforge/FR.md

# Reconcile go.mod (target uses isaacphi module path; new subtree needs its own go.mod)
cp $MCPFORGE/go.mod rust/mcp-forge-bridge/go.mod
# …and rename module path to github.com/KooshaPari/McpKit-contrib/mcpforge-bridge
# to avoid colliding with the isaacphi fork's module path.

git add -A
git commit -m "feat(mcpforge-bridge): absorb MCPForge internal/lsp and internal/protocol"
git push origin absorb/mcpforge-bridge
gh pr create --draft   # title + body below
```

### PR title and description

- **Title:** `feat(mcpforge-bridge): absorb MCPForge LSP bridge layer (internal/lsp, internal/protocol)`
- **Body:**
  > Absorbs the LSP bridge layer and protocol types from `KooshaPari/MCPForge` (KooshaPari's fork of mcp-language-server) into a new sibling subtree `rust/mcp-forge-bridge/`. Leaves the existing `rust/mcp-forge/` (isaacphi upstream fork) untouched for now; a follow-up PR will reconcile.
  >
  > **What's included:**
  > - `rust/mcp-forge-bridge/internal/lsp/` — client, transport, methods, protocol, server-request-handlers, typescript, detect-language (+ tests)
  > - `rust/mcp-forge-bridge/internal/protocol/` — LSP type tables/interfaces (BSD-3 from upstream Microsoft tsprotocol; LICENSE preserved)
  > - `docs/mcpforge/FR.md` — functional requirements from MCPForge
  >
  > **What's NOT included:** `internal/tools/`, `main.go`, `cmd/generate/` (these go to PhenoMCPServers per the ecosystem plan).
  >
  > **Followups:** rename module path, reconcile with existing `rust/mcp-forge/`, wire into `go/go.work`.

### Risks / blockers

- **R-MK-1 (HIGH):** Maintainer of McpKit may reject duplication of `rust/mcp-forge/`. Pre-discuss in an issue before opening the PR.
- **R-MK-2 (MED):** `internal/protocol/tsprotocol.go` is 278 KB — likely GitHub blob warning. Use Git LFS or split before push if maintainers request.
- **R-MK-3 (LOW):** `go.work` references missing modules; not in scope here but will surface in CI. Note in PR.
- **R-MK-4 (LOW):** License compat. BSD-3 + Apache-2.0/MIT is fine (BSD-3 is permissive, Apache notice requirements met by preserving LICENSE).

---

## Target 2 — `KooshaPari/PhenoMCPServers` (implementations registry)

- **Default branch:** `main` | **Language:** Python | **Last push:** 2026-06-18 (very active)
- **Existing surface relevant to absorption:**
  - `servers/external/mcpforge` — **git submodule pointing at THIS repo** (`KooshaPari/MCPForge.git`). Pin: SHA `7c36761287`. **This is the current absorption channel for the whole repo.**
  - `servers/external/ops-mcp` — another submodule (`KooshaPari/phenotype-ops-mcp`).
  - `servers/substrate/` — Python `substrate_server.py`, `dispatch_server.py`, `lead_server.py`, `team_mailbox_server.py`. Uses `fastmcp>=3.4.2`. **Pre-existing reference impl.**
  - `servers/pheno-org/` — Python `pheno_org_server.py` (FastMCP). Another reference impl.
  - `templates/mcp-server/` — Python template with `server.py.tmpl`, `pyproject.toml.tmpl`, `mcp.json.tmpl`.
  - `catalog/registry.yaml` SSOT.
  - All Python servers depend on `phenofastmcp @ git+https://github.com/KooshaPari/PhenoFastMCP.git@v3.4.2`.

### Push set

| MCPForge source | Target path in PhenoMCPServers | Action |
|---|---|---|
| `internal/tools/` (11 files, 6 tools) | `servers/external/mcpforge/internal/tools/` | **Already present via submodule** — needs `git submodule update --remote` or pointer update |
| `main.go` + `tools.go` | `servers/external/mcpforge/` | Same — present via submodule |
| `cmd/generate/` | `servers/external/mcpforge/cmd/generate/` | Same |
| `integrationtests/` (entire dir, 5 language subdirs) | `servers/external/mcpforge/integrationtests/` | Same — present |
| New: `docs/mcpforge-impl/README.md` | `docs/mcpforge-impl/` | **Add** in-tree doc referencing the submodule and listing the 6 tools |

### Critical insight

The "absorption" into PhenoMCPServers **already happened**: `servers/external/mcpforge/` is a git submodule whose pointer is at SHA `7c36761` of MCPForge. Any new push from MCPForge is therefore not a "copy" — it's a **submodule pointer bump** on the receiving side. The actual code lives in MCPForge and is referenced by SHA.

The work to do here is:

1. **In MCPForge:** commit any pending changes (the repo is the source of truth).
2. **In PhenoMCPServers:** bump the submodule pointer to the new MCPForge SHA.
3. **Add an in-tree pointer document** at `docs/mcpforge-impl/` that catalogs the 6 tools, links to MCPForge, and notes the Go-tool implementation is a reference impl for ports to other languages.

### Conflicts to resolve before push

1. **C-PMCP-1 (NONE for code):** No file collisions because the absorption channel is a submodule.
2. **C-PMCP-2 (LOW):** `catalog/registry.yaml` already has an `mcpforge` entry (referenced in README). Confirm the new SHA + version are reflected; bump version field if registry has version pinning.
3. **C-PMCP-3 (MED):** `docs/` in PhenoMCPServers already has structure. New `docs/mcpforge-impl/` is additive — verify no `docs/mcpforge*` prefix collides.

### Strategy (submodule pointer bump + in-tree doc)

```bash
# In PhenoMCPServers worktree at repos/PhenoMCPServers-wtrees/bump-mcpforge
git fetch origin
git switch -c chore/bump-mcpforge origin/main

# 1. Bump submodule to the new MCPForge SHA
cd servers/external/mcpforge
git fetch   # inside the submodule
git checkout <NEW_MCPFORGE_SHA>
cd ../..
git add servers/external/mcpforge
git commit -m "chore(submodules): bump mcpforge to <NEW_SHA>"

# 2. Add pointer doc
mkdir -p docs/mcpforge-impl
# Write README listing 6 tools with paths and FR references
git add docs/mcpforge-impl
git commit -m "docs(mcpforge): add implementation reference doc"
git push origin chore/bump-mcpforge
gh pr create
```

### PR title and description

- **Title:** `chore(submodules): bump mcpforge + add implementation reference doc`
- **Body:**
  > Bumps the `servers/external/mcpforge` submodule pointer to the latest MCPForge SHA, picking up the 6 LSP tool implementations (`definition`, `references`, `diagnostics`, `hover`, `rename_symbol`, `edit_file`), the Go entry point (`main.go` + `tools.go`), and the `cmd/generate/` LSP types generator.
  >
  > Also adds `docs/mcpforge-impl/README.md` as an in-tree reference catalog for ports to non-Go languages (the `templates/mcp-server/` Python template and `servers/substrate/` Python impls are the pattern to follow).
  >
  > No new top-level server directory is created — `mcpforge` remains a submodule by design (see repo README "Servers" table).

### Risks / blockers

- **R-PMCP-1 (HIGH):** The receiving repo is very active (last push today). Coordinate with the PhenoMCPServers maintainer — submodule bumps are usually fine but multiple in-flight PRs may conflict on the submodule SHA.
- **R-PMCP-2 (MED):** If MCPForge's `main` branch is force-pushed or rewritten, the pinned SHA becomes invalid. Confirm MCPForge maintainer agrees to non-destructive history.
- **R-PMCP-3 (LOW):** `docs/` already houses `LANGUAGE-TIERS-AND-ROLES.md`. New doc must be consistent with tier model; cross-link rather than duplicate.

---

## Target 3 — `KooshaPari/Agentora` (Rust hexagonal AI agent framework)

- **Default branch:** `main` | **Language:** Python (description) / Rust (actual `Cargo.toml`) | **Last push:** 2026-06-17
- **Existing surface relevant to absorption:**
  - Root `Cargo.toml` is a workspace with **9 members** registered (`pheno-agent/*`, `pheno-proc-runtime/*`, `bifrost-routing`, `forgecode-core`). Excludes legacy duplicates under `agents/phenoagent/`.
  - `rust-toolchain.toml` pinned to `1.95.0`; MSRV `1.75`.
  - `crates/phenotype-forge/` is a **staging stub** (single `MIGRATED.md` file noting canonical owner is `KooshaPari/HexaKit`). **Do not depend on it.**
  - `crates/ABSORPTION_MANIFEST.md` documents the staging policy for PhenoProc wave 5–6. Pattern: crates under `crates/<name>/` are staged sources; only some are workspace members.
  - `crates/bifrost-routing/`, `crates/forgecode-core/` — real Rust crates, pattern to follow.
  - The `wip/mcp-forge-lsp-rust/` worklog note (in `docs/worklogs/RESEARCH.md`) says scaffolding was built for `mcp-forge-lsp` with `tokio + lsp-server + lsp-types + uniffi` but abandoned with 13 compile errors. **No live Rust code to push.**

### Push set

| MCPForge source | Target path in Agentora | Action |
|---|---|---|
| (none — no live Rust files) | — | **No code push.** This target receives *only* dependency-choice recommendations + worklog pointer. |
| New: `crates/phenotype-forge/RECIPES.md` | `crates/phenotype-forge/RECIPES.md` | **Add** in the existing staging crate — documents the abandoned dependency choices (`tokio + lsp-server + lsp-types + uniffi`) for future re-attempt |
| New: `docs/research/rust-lsp-forge-scaffolding.md` | `docs/research/rust-lsp-forge-scaffolding.md` | **Add** pointer to MCPForge `docs/worklogs/RESEARCH.md` so Agentora doesn't re-derive the abandoned work |

### Conflicts to resolve before push

1. **C-AG-1 (HIGH — pre-discuss):** `crates/phenotype-forge/MIGRATED.md` declares the canonical owner is `KooshaPari/HexaKit`, **not** Agentora. Pushing new files there may be rejected. Alternative path: add to `docs/research/` only and leave the staging crate untouched.
2. **C-AG-2 (LOW):** `Cargo.toml` MSRV is `1.75`; toolchain pinned `1.95.0`. Rust scaffolding research (if re-attempted) must respect MSRV — `uniffi` may require a newer MSRV; flag this before any new Rust work.
3. **C-AG-3 (LOW):** The worklog mentions a "wave" pattern for staged sources. New docs should follow `crates/ABSORPTION_MANIFEST.md` style.

### Strategy (docs only, no code)

```bash
# In Agentora worktree at repos/Agentora-wtrees/absorb-mcpforge-research
git fetch origin
git switch -c docs/mcpforge-lsp-research origin/main

# Option A (preferred): docs only
mkdir -p docs/research
# Write docs/research/rust-lsp-forge-scaffolding.md pointing at MCPForge's
# docs/worklogs/RESEARCH.md and the abandoned dependency choices.

git add docs/research/rust-lsp-forge-scaffolding.md
git commit -m "docs(research): record MCPForge Rust LSP scaffolding abandonment"
git push origin docs/mcpforge-lsp-research
gh pr create
```

### PR title and description

- **Title:** `docs(research): record MCPForge Rust LSP scaffolding abandonment + dep recipe`
- **Body:**
  > Documents the abandoned Rust LSP scaffolding effort that was recorded in MCPForge's `docs/worklogs/RESEARCH.md` (2026-06-18). No code is added — this preserves the institutional knowledge in Agentora so the dependency choices (`tokio + lsp-server + lsp-types + uniffi`) and known compile-error mode aren't re-derived later.
  >
  > Does **not** modify `Cargo.toml`, `crates/phenotype-forge/MIGRATED.md`, or workspace membership. The staging crate at `crates/phenotype-forge/` remains pointing at its canonical owner (`KooshaPari/HexaKit`).

### Risks / blockers

- **R-AG-1 (HIGH):** MCPForge's `RESEARCH.md` says the Rust layer was "deprioritized pending resolution of compile errors and uniffi CLI integration complexity." Pushing a "recipe" doc into Agentora risks re-opening a wound. Frame the PR as **archival**, not a re-start.
- **R-AG-2 (MED):** If maintainers prefer the note live only in MCPForge's worklog, this PR may be closed in favor of a cross-repo link from Agentora's existing docs.
- **R-AG-3 (LOW):** No code, so no MSRV/CI risk. Pure-doc PRs usually merge fast.

---

## Cross-target sequencing & coordination

| Order | Target | Reason |
|-------|--------|--------|
| 1 | **McpKit** | Highest conflict surface; needs maintainer pre-discussion (R-MK-1). |
| 2 | **PhenoMCPServers** | Submodule pointer bump is mechanical once MCPForge is settled; depends on MCPForge SHA being stable. |
| 3 | **Agentora** | Docs only, lowest risk; can be deferred or dropped if maintainers decline. |

**Hard dependencies:**

- McpKit PR must be merged (or path chosen) before PhenoMCPServers README can confidently cross-reference `docs/mcpforge/FR.md`.
- PhenoMCPServers PR depends on the **final MCPForge SHA** — coordinate the submodule bump with the MCPForge maintainer so the pointer doesn't churn.

**Recommended PR pairing:** open all three as **draft** PRs simultaneously, mark cross-repo dependencies in each description, then iterate.

---

## Subtree vs manual copy decision matrix

| Target | Subtree add | Subtree split | Manual copy | Chosen |
|--------|------------|---------------|-------------|--------|
| McpKit | Risky — receiving side already has a `rust/mcp-forge/` subtree | Clean — MCPForge→McpKit/bridge as split repo | Simplest — branch + rsync + PR | **Manual copy** (worktree branch) |
| PhenoMCPServers | N/A — absorption is via submodule pointer | N/A | Submodule bump + in-tree doc | **Submodule bump** |
| Agentora | N/A — no code to push | N/A | Docs only | **Manual copy** (single md file) |

---

## Pre-flight checklist (before opening any PR)

- [ ] Confirm MCPForge `main` branch is **not** going to be force-pushed after the absorption SHAs are pinned (R-PMCP-2).
- [ ] File issue on McpKit proposing the `rust/mcp-forge-bridge/` path; wait for maintainer ack (R-MK-1).
- [ ] Confirm PhenoMCPServers maintainer is OK with submodule pointer bump landing in `main` (R-PMCP-1).
- [ ] Confirm Agentora maintainer accepts archival-only docs PR (R-AG-1).
- [ ] All three receiving repos have `AGENTS.md` requiring Conventional Commits + `<type>/<topic>` branches — branch names in this plan already follow that convention.

**Status:** Plan only. No pushes executed.
**Tags:** `[cross-repo]` `[PLAN]` `[MCPForge]` `[McpKit]` `[PhenoMCPServers]` `[Agentora]`
