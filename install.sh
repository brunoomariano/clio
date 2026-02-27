#!/usr/bin/env bash
set -euo pipefail

REPO_URL=${REPO_URL:-"https://github.com/brunoomariano/clio"}
BRANCH=${BRANCH:-"main"}
INSTALL_BIN_NAME="clio"

ARCH=$(uname -m)
case "$ARCH" in
  x86_64) GOARCH="amd64" ;;
  aarch64|arm64) GOARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
 esac

install_dir="$HOME/.local/bin"
mkdir -p "$install_dir"

ensure_go() {
  if command -v go >/dev/null 2>&1; then
    return 0
  fi
  if ! command -v lsb_release >/dev/null 2>&1; then
    echo "Go not found and distro detection unavailable. Install Go manually." >&2
    exit 1
  fi
  distro=$(lsb_release -is | tr '[:upper:]' '[:lower:]')
  if [ "$distro" != "ubuntu" ] && [ "$distro" != "debian" ]; then
    echo "Go not found. Automatic install only supports Debian/Ubuntu." >&2
    exit 1
  fi
  go_version=$(curl -sSL https://go.dev/VERSION?m=text | head -n1)
  if [ -z "$go_version" ]; then
    echo "Failed to resolve latest Go version." >&2
    exit 1
  fi
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
  url="https://go.dev/dl/${go_version}.linux-${GOARCH}.tar.gz"
  curl -sSL "$url" -o "$tmpdir/go.tgz"
  mkdir -p "$HOME/.local"
  rm -rf "$HOME/.local/go"
  tar -C "$HOME/.local" -xzf "$tmpdir/go.tgz"
  export PATH="$HOME/.local/go/bin:$PATH"
}

fetch_source() {
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
  if command -v git >/dev/null 2>&1; then
    git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$tmpdir/src" >/dev/null 2>&1
  else
    curl -sSL "$REPO_URL/archive/refs/heads/${BRANCH}.tar.gz" -o "$tmpdir/src.tgz"
    mkdir -p "$tmpdir/src"
    tar -xzf "$tmpdir/src.tgz" -C "$tmpdir"
    mv "$tmpdir"/*/ "$tmpdir/src/"
  fi
  echo "$tmpdir/src"
}

ensure_go
src_dir=$(fetch_source)

( cd "$src_dir" && go build -o "$INSTALL_BIN_NAME" ./cmd/clio )

install -m 0755 "$src_dir/$INSTALL_BIN_NAME" "$install_dir/$INSTALL_BIN_NAME"

echo "Installed $INSTALL_BIN_NAME to $install_dir/$INSTALL_BIN_NAME"
