# docsh Makefile

APP_NAME = docsh
VERSION = 1.0.0
BUILD_FLAGS = -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: all clean windows windows64 windows32 linux macos test

# デフォルトターゲット
all: windows64 windows32 linux macos

# Windows 64bit
windows64:
	@echo "Building for Windows 64bit..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(APP_NAME)-windows-x64.exe .

# Windows 32bit
windows32:
	@echo "Building for Windows 32bit..."
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(APP_NAME)-windows-x86.exe .

# Windows (両方)
windows: windows64 windows32

# Linux 64bit
linux:
	@echo "Building for Linux 64bit..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(APP_NAME)-linux-x64 .

# macOS Intel
macos-intel:
	@echo "Building for macOS Intel..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(APP_NAME)-macos-x64 .

# macOS ARM64
macos-arm64:
	@echo "Building for macOS ARM64..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(APP_NAME)-macos-arm64 .

# macOS (両方)
macos: macos-intel macos-arm64

# ローカル環境用ビルド
local:
	@echo "Building for local environment..."
	go build $(BUILD_FLAGS) -o $(APP_NAME) .

# テスト実行
test:
	@echo "Running tests..."
	go test -v ./...

# 依存関係の更新
deps:
	@echo "Updating dependencies..."
	go mod tidy
	go mod download

# クリーンアップ
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(APP_NAME)-*
	rm -f $(APP_NAME).exe
	rm -f $(APP_NAME)

# リリース用ビルド（全プラットフォーム）
release: clean all
	@echo "Creating release builds..."
	@echo "Build completed for version $(VERSION)"
	@ls -la $(APP_NAME)-*

# 開発用（ローカルビルド+テスト）
dev: test local
	@echo "Development build completed"

# バージョン情報表示
version:
    @echo "docsh version: $(VERSION)"
	@go version

# ヘルプ
help:
    @echo "docsh Build System"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all          - Build for all platforms"
	@echo "  windows      - Build for Windows (32bit + 64bit)"
	@echo "  windows64    - Build for Windows 64bit"
	@echo "  windows32    - Build for Windows 32bit"
	@echo "  linux        - Build for Linux 64bit"
	@echo "  macos        - Build for macOS (Intel + ARM64)"
	@echo "  macos-intel  - Build for macOS Intel"
	@echo "  macos-arm64  - Build for macOS ARM64"
	@echo "  local        - Build for current platform"
	@echo "  test         - Run tests"
	@echo "  deps         - Update dependencies"
	@echo "  clean        - Clean build artifacts"
	@echo "  release      - Create release builds"
	@echo "  dev          - Development build (test + local)"
	@echo "  version      - Show version information"
	@echo "  help         - Show this help"