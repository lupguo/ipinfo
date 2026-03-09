#!/usr/bin/env bash
# install.sh — one-click installer for ipinfo
# Usage: curl -fsSL https://raw.githubusercontent.com/lupguo/ip_info/main/install.sh | bash

set -euo pipefail

REPO="lupguo/ipinfo"
BINARY="ipinfo"

# ── Detect OS and architecture ────────────────────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  os="linux"  ;;
  Darwin) os="darwin" ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64)          arch="amd64" ;;
  aarch64 | arm64) arch="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

ASSET_NAME="${BINARY}-${os}-${arch}"

# ── Fetch latest release tag ──────────────────────────────────────────────────
echo "Fetching latest release..."
LATEST_TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | head -1 \
  | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
  echo "Could not determine latest release tag." >&2
  exit 1
fi

echo "Latest version: $LATEST_TAG"

BASE_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}"
DOWNLOAD_URL="${BASE_URL}/${ASSET_NAME}"
CHECKSUM_URL="${BASE_URL}/checksums.txt"

# ── Download binary and checksums ─────────────────────────────────────────────
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading ${ASSET_NAME}..."
curl -fsSL -o "${TMP_DIR}/${ASSET_NAME}" "$DOWNLOAD_URL"
curl -fsSL -o "${TMP_DIR}/checksums.txt" "$CHECKSUM_URL"

# ── Verify SHA256 checksum ────────────────────────────────────────────────────
echo "Verifying checksum..."
cd "$TMP_DIR"

if [ "$OS" = "Darwin" ] && command -v shasum >/dev/null 2>&1; then
  grep "${ASSET_NAME}" checksums.txt | shasum -a 256 --check --status
elif command -v sha256sum >/dev/null 2>&1; then
  grep "${ASSET_NAME}" checksums.txt | sha256sum --check --status
elif command -v shasum >/dev/null 2>&1; then
  grep "${ASSET_NAME}" checksums.txt | shasum -a 256 --check --status
else
  echo "Warning: no sha256sum/shasum found, skipping verification." >&2
fi

cd - >/dev/null

# ── Install ───────────────────────────────────────────────────────────────────
chmod +x "${TMP_DIR}/${ASSET_NAME}"

INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

mv "${TMP_DIR}/${ASSET_NAME}" "${INSTALL_DIR}/${BINARY}"

echo ""
echo "Installed ${BINARY} ${LATEST_TAG} to ${INSTALL_DIR}/${BINARY}"

# ── Sync example config (only if user has no config yet) ──────────────────────
CONFIG_DIR="${HOME}/.ipinfo"
CONFIG_FILE="${CONFIG_DIR}/config.yaml"
CONFIG_RAW_URL="https://raw.githubusercontent.com/${REPO}/${LATEST_TAG}/config.example.yaml"

if [ ! -f "$CONFIG_FILE" ]; then
  mkdir -p "$CONFIG_DIR"
  if curl -fsSL -o "$CONFIG_FILE" "$CONFIG_RAW_URL" 2>/dev/null; then
    echo "Config written to ${CONFIG_FILE}"
  else
    echo "Warning: could not download example config. The binary will auto-generate a minimal config on first run." >&2
  fi
else
  echo "Config already exists at ${CONFIG_FILE} — not overwritten."
fi

# Remind user if install dir is not in PATH
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "Note: ${INSTALL_DIR} is not in your PATH."
    echo "Add the following to your shell profile:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac
