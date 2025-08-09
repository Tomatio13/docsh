#!/bin/bash
# docsh Cross-platform Build Script

echo "Building docsh for multiple platforms..."

# Create Release directory if it doesn't exist
mkdir -p Release

# Windows 64bit
echo "Building for Windows 64bit..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o Release/docsh-windows-x64.exe .
if [ $? -eq 0 ]; then
    echo "✓ Windows 64bit build successful"
    ls -lh Release/docsh-windows-x64.exe
else
    echo "✗ Windows 64bit build failed"
fi

## Windows 32bit (dropped)
## Removed due to syscall incompatibilities and low demand

# Linux 64bit
echo "Building for Linux 64bit..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o Release/docsh-linux-x64 .
if [ $? -eq 0 ]; then
    echo "✓ Linux 64bit build successful"
    ls -lh Release/docsh-linux-x64
else
    echo "✗ Linux 64bit build failed"
fi

# macOS 64bit (Intel)
echo "Building for macOS 64bit (Intel)..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o Release/docsh-macos-x64 .
if [ $? -eq 0 ]; then
    echo "✓ macOS 64bit build successful"
    ls -lh Release/docsh-macos-x64
else
    echo "✗ macOS 64bit build failed"
fi

# macOS ARM64 (Apple Silicon)
echo "Building for macOS ARM64 (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o Release/docsh-macos-arm64 .
if [ $? -eq 0 ]; then
    echo "✓ macOS ARM64 build successful"
    ls -lh Release/docsh-macos-arm64
else
    echo "✗ macOS ARM64 build failed"
fi

echo ""
echo "Build completed! Available binaries in Release folder:"
echo "Windows 64bit: Release/docsh-windows-x64.exe"
echo "Windows 32bit: (dropped)"
echo "Linux 64bit:   Release/docsh-linux-x64"
echo "macOS Intel:   Release/docsh-macos-x64"
echo "macOS ARM64:   Release/docsh-macos-arm64"