#!/bin/sh
# go-agent installer.
#
#   curl -fsSL https://raw.githubusercontent.com/tqrcisio/go-agent/main/install.sh | sh
#
# Downloads the latest release binary for your platform and installs it to
# ~/.local/bin (override with GO_AGENT_INSTALL_DIR). Pin a version with
# GO_AGENT_VERSION=v1.2.0.

set -eu

REPO="tqrcisio/go-agent"
BIN="go-agent"
INSTALL_DIR="${GO_AGENT_INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${GO_AGENT_VERSION:-latest}"

err() {
	printf 'error: %s\n' "$1" >&2
	exit 1
}

command -v curl >/dev/null 2>&1 || err "curl is required"
command -v tar  >/dev/null 2>&1 || err "tar is required"

OS="$(uname -s)"
case "$OS" in
	Linux | Darwin) ;;
	*) err "unsupported OS: $OS (download the Windows zip from https://github.com/$REPO/releases)" ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
	x86_64 | amd64) ARCH="x86_64" ;;
	arm64 | aarch64) ARCH="arm64" ;;
	*) err "unsupported architecture: $ARCH" ;;
esac

ASSET="${BIN}_${OS}_${ARCH}.tar.gz"
if [ "$VERSION" = "latest" ]; then
	URL="https://github.com/$REPO/releases/latest/download/$ASSET"
else
	URL="https://github.com/$REPO/releases/download/$VERSION/$ASSET"
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

printf 'Downloading %s...\n' "$ASSET"
if ! curl -fsSL "$URL" -o "$TMP/$ASSET"; then
	err "could not download $URL (does the release exist for your platform?)"
fi

tar -xzf "$TMP/$ASSET" -C "$TMP"
[ -f "$TMP/$BIN" ] || err "archive did not contain $BIN"

mkdir -p "$INSTALL_DIR"
install -m 0755 "$TMP/$BIN" "$INSTALL_DIR/$BIN"

printf '\ngo-agent installed to %s/%s\n' "$INSTALL_DIR" "$BIN"

case ":$PATH:" in
	*":$INSTALL_DIR:"*) ;;
	*)
		printf '\n%s is not on your PATH. Add this to your shell profile:\n' "$INSTALL_DIR"
		printf '  export PATH="%s:$PATH"\n' "$INSTALL_DIR"
		;;
esac

printf '\nRun "go-agent version" to confirm, then "go-agent chat" to start.\n'
