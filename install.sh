#!/bin/bash
set -e

REPO="yerbapadre/git-back"
VERSION_ARG=${1:-latest}

# Resolve "latest" to actual version tag
if [ "$VERSION_ARG" = "latest" ]; then
    echo "Fetching latest version..."
    VERSION=$(curl -sfL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        echo "❌ Failed to fetch latest version"
        exit 1
    fi
    echo "Latest version: $VERSION"
else
    VERSION="$VERSION_ARG"
fi

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  darwin)
    OS_NAME="darwin"
    ;;
  linux)
    OS_NAME="linux"
    ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64)
    ARCH_NAME="amd64"
    ;;
  arm64|aarch64)
    ARCH_NAME="arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

echo "Installing git-back for $OS_NAME-$ARCH_NAME..."

# Download and install
BINARY_NAME="git-back-$OS_NAME-$ARCH_NAME"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/git-back-$VERSION-$OS_NAME-$ARCH_NAME.tar.gz"

TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

cd "$TMP_DIR"

echo "Downloading from $DOWNLOAD_URL..."
if ! curl -sfL "$DOWNLOAD_URL" -o git-back.tar.gz; then
    echo "❌ Failed to download git-back"
    echo "Check if version $VERSION exists at: https://github.com/$REPO/releases"
    exit 1
fi

# Download checksum file
CHECKSUM_URL="https://github.com/$REPO/releases/download/$VERSION/checksums.txt"
if curl -sfL "$CHECKSUM_URL" -o checksums.txt 2>/dev/null; then
    echo "Verifying checksum..."
    if command -v shasum >/dev/null 2>&1; then
        grep "git-back-$VERSION-$OS_NAME-$ARCH_NAME.tar.gz" checksums.txt | shasum -a 256 -c - || {
            echo "❌ Checksum verification failed!"
            exit 1
        }
        echo "✅ Checksum verified"
    else
        echo "⚠️  shasum not found, skipping checksum verification"
    fi
fi

tar xzf git-back.tar.gz

# Install to user's local bin
mkdir -p "$HOME/.local/bin"
mv "$BINARY_NAME" "$HOME/.local/bin/git-back"
chmod +x "$HOME/.local/bin/git-back"

echo "✅ git-back installed to $HOME/.local/bin/git-back"
echo ""
echo "Make sure $HOME/.local/bin is in your PATH:"
echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
echo ""
echo "Run 'git-back' in any git repository to get started!"

cd - > /dev/null
rm -rf "$TMP_DIR"
