<p align="center">
  <img src="assets/logo.png" alt="Clio logo" width="320" />
</p>


# Clio

Clio is a fast, minimalist TUI for personal notes in Markdown. It prioritizes performance, real-time search, and a clean architecture. Notes are plain `.md` files with YAML frontmatter.

## Installation

Single-command install/update:

```bash
curl -sSL https://raw.githubusercontent.com/brunoomariano/clio/main/install.sh | bash
```

The installer:

- Detects `amd64`/`arm64`
- Builds locally with Go if available
- Installs Go automatically on Debian/Ubuntu if missing
- Installs to `~/.local/bin/clio`

## Update

Run the same install command again to update to the latest version:

```bash
curl -sSL https://raw.githubusercontent.com/brunoomariano/clio/main/install.sh | bash
```

## Usage

Run:

```bash
clio
```

Somente pasta atual:

```bash
clio --cwd
clio --cwd --suffixes="['*.md','*.json']" --ignore_paths="['test.*']"
```

Default config:

```
~/.config/clio.yaml
```

Default search configuration:

```
search_dirs:
  - path: "~/.local/share/clio/notes"
  - path: "~/work/wiki"
    suffixes: [".*\\.md$", ".*\\.rst$"] # override do global_suffixes para esta pasta
    ignore_paths: ["^archive/.*"]       # override do global_ignore_paths para esta pasta

global_suffixes: ["*.md", "*.txt", "*.json", "*.yaml"]
global_ignore_paths: ["ignore/*", "tests/*"]
```

Editor configuration (default `nvim`, fallback to `nano` if editor missing):

```
editor: "nvim"
terminal: ""   # optional; if empty, auto-detects a terminal
```

## Key Bindings

- `?` open menu (all actions live here)

## Architecture

```
cmd/clio/main.go
internal/model
internal/store
internal/index
internal/ui
```

- `model`: note and config structs, parsing and atomic saves
- `store`: filesystem-backed note store + watcher
- `index`: in-memory BM25 index and search
- `ui`: Bubble Tea TUI and key bindings

## BM25 Overview

BM25 ranks documents by term frequency and document length normalization. A term´s score grows with frequency but saturates using `k1` and `b`. This gives better relevance than plain term counts by penalizing overly long documents.

## Debounce + Cancellation

Search runs after a configurable debounce (default 100ms). Each new keystroke cancels the previous search using `context.Context`, so outdated work is dropped and the UI remains responsive.

## Testing Strategy

- Full coverage across `model`, `store`, `index`, and `ui`
- Explicit docstrings in every test describing scenario, relevance, and expected behavior
- CI enforces `>= 90%` coverage

Run tests:

```bash
go test ./... -cover
```

## Makefile

- `make build`
- `make test`
- `make install`
- `make release`

# Clio TUI Contracts

Este repositorio contem os contratos de layout, estado, snapshots e diffs para a TUI.
