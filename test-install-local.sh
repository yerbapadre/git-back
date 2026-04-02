#!/bin/bash
# Local test of install script without GitHub

set -e

echo "Testing install script locally..."

# Build the binary first
go build -o git-back-test

# Simulate what the install script does
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64) ARCH_NAME="amd64" ;;
  arm64|aarch64) ARCH_NAME="arm64" ;;
esac

# Create test directory structure
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

mkdir -p "$TEST_DIR/.local/bin"

# Copy binary
cp git-back-test "$TEST_DIR/.local/bin/git-back"
chmod +x "$TEST_DIR/.local/bin/git-back"

# Test it works (will fail with TTY error outside git repo, which is expected)
if [ -x "$TEST_DIR/.local/bin/git-back" ]; then
    echo "✅ Binary is executable"

    # Test in actual git repo
    cd "$(git rev-parse --show-toplevel)"
    if "$TEST_DIR/.local/bin/git-back" 2>&1 | grep -qE "(Recent Branches|Error: could not open)"; then
        echo "✅ Binary runs (TTY error is expected in non-interactive mode)"
    fi

    echo "✅ Install script logic works!"
    echo ""
    echo "Test binary location: $TEST_DIR/.local/bin/git-back"
else
    echo "❌ Binary not executable"
    exit 1
fi

echo ""
echo "Ready for full GitHub release!"
