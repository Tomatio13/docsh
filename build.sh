#!/bin/bash
# Go-Zsh Cross-platform Build Script

echo "Building Go-Zsh for multiple platforms..."

# Windows 64bit
echo "Building for Windows 64bit..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o go-zsh-windows-x64.exe .
if [ $? -eq 0 ]; then
    echo "✓ Windows 64bit build successful"
    ls -lh go-zsh-windows-x64.exe
else
    echo "✗ Windows 64bit build failed"
fi

# Windows 32bit
echo "Building for Windows 32bit..."
GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o go-zsh-windows-x86.exe .
if [ $? -eq 0 ]; then
    echo "✓ Windows 32bit build successful"
    ls -lh go-zsh-windows-x86.exe
else
    echo "✗ Windows 32bit build failed"
fi

# Linux 64bit
echo "Building for Linux 64bit..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o go-zsh-linux-x64 .
if [ $? -eq 0 ]; then
    echo "✓ Linux 64bit build successful"
    ls -lh go-zsh-linux-x64
else
    echo "✗ Linux 64bit build failed"
fi

# macOS 64bit (Intel)
echo "Building for macOS 64bit (Intel)..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o go-zsh-macos-x64 .
if [ $? -eq 0 ]; then
    echo "✓ macOS 64bit build successful"
    ls -lh go-zsh-macos-x64
else
    echo "✗ macOS 64bit build failed"
fi

# macOS ARM64 (Apple Silicon)
echo "Building for macOS ARM64 (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o go-zsh-macos-arm64 .
if [ $? -eq 0 ]; then
    echo "✓ macOS ARM64 build successful"
    ls -lh go-zsh-macos-arm64
else
    echo "✗ macOS ARM64 build failed"
fi

echo ""
echo "Build completed! Available binaries:"
echo "Windows 64bit: go-zsh-windows-x64.exe"
echo "Windows 32bit: go-zsh-windows-x86.exe"
echo "Linux 64bit:   go-zsh-linux-x64"
echo "macOS Intel:   go-zsh-macos-x64"
echo "macOS ARM64:   go-zsh-macos-arm64"