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

Default config:

```
~/.config/clio.yaml
```

Default notes directory:

```
~/.local/share/clio/notes
```

## Key Bindings

- `/` focus search
- `enter` open
- `n` new
- `e` edit
- `d` delete
- `t` edit tags
- `x` set/clear expiry
- `r` toggle regex
- `+` add boost tag
- `-` add exclude tag
- `q` quit

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
