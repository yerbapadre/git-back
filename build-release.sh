#!/bin/bash
set -e

VERSION=${1:-v1.0.0}
echo "Building release $VERSION..."

mkdir -p dist

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o dist/git-back-darwin-amd64
tar -czf dist/git-back-$VERSION-darwin-amd64.tar.gz -C dist git-back-darwin-amd64

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o dist/git-back-darwin-arm64
tar -czf dist/git-back-$VERSION-darwin-arm64.tar.gz -C dist git-back-darwin-arm64

# Linux (x64)
GOOS=linux GOARCH=amd64 go build -o dist/git-back-linux-amd64
tar -czf dist/git-back-$VERSION-linux-amd64.tar.gz -C dist git-back-linux-amd64

# Linux (ARM64)
GOOS=linux GOARCH=arm64 go build -o dist/git-back-linux-arm64
tar -czf dist/git-back-$VERSION-linux-arm64.tar.gz -C dist git-back-linux-arm64

# Windows (x64)
GOOS=windows GOARCH=amd64 go build -o dist/git-back-windows-amd64.exe
zip -j dist/git-back-$VERSION-windows-amd64.zip dist/git-back-windows-amd64.exe

echo "✅ Built all binaries in dist/"
ls -lh dist/*.tar.gz dist/*.zip
