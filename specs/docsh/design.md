# Design Document

## Overview

dockernautは、cherryshellを参考にしたGo言語ベースのマルチプラットフォーム対応CLIシェルです。このシェルは、Dockerコマンドと一般的なLinuxコマンドの対応関係を理解しやすくし、開発者がより直感的にDockerを操作できるようにします。エイリアス機能、コンテキスト管理、ヒストリ機能、高度な自動補完機能を提供します。

## Architecture

### Core Architecture
- **言語**: Go (Golang)
- **CLI Framework**: cobra-cli
- **設定管理**: viper
- **国際化**: go-i18n (日本語・英語対応)
- **マルチプラットフォーム**: Windows, macOS, Linux対応
- **パッケージ管理**: Go modules

### Shell Architecture
- **コマンドパーサー**: カスタムコマンド解析エンジン
- **マッピングエンジン**: Dockerコマンド変換システム
- **実行エンジン**: Docker CLI wrapper
- **設定システム**: YAML/JSON設定ファイル
- **国際化システム**: 多言語メッセージ・ヘルプ対応

## Components and Interfaces

### Core Components

#### 1. Alias Management System
```go
type Alias struct {
    Name        string `json:"name" yaml:"name"`
    Command     string `json:"command" yaml:"command"`
    Description string `json:"description,omitempty" yaml:"description,omitempty"`
    CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
}

type AliasManager interface {
    CreateAlias(name, command, description string) error
    DeleteAlias(name string) error
    GetAlias(name string) (*Alias, error)
    ListAliases() ([]*Alias, error)
    ResolveCommand(input string) (string, bool)
    LoadAliases() error
    SaveAliases() error
}
```

#### 2. Context Management System
```go
type ContainerContext struct {
    ContainerID   string `json:"container_id" yaml:"container_id"`
    ContainerName string `json:"container_name" yaml:"container_name"`
    Image         string `json:"image" yaml:"image"`
    Status        string `json:"status" yaml:"status"`
    SetAt         time.Time `json:"set_at" yaml:"set_at"`
}

type ContextManager interface {
    SetCurrentContainer(containerID string) error
    GetCurrentContainer() (*ContainerContext, error)
    ClearCurrentContainer()
    IsContainerSet() bool
    GetPromptString() string
    ValidateContainer(containerID string) error
}
```

#### 3. History Management System
```go
type HistoryEntry struct {
    ID        int       `json:"id"`
    Command   string    `json:"command"`
    Timestamp time.Time `json:"timestamp"`
    ExitCode  int       `json:"exit_code"`
    Duration  time.Duration `json:"duration"`
}

type HistoryManager interface {
    AddEntry(command string, exitCode int, duration time.Duration) error
    GetHistory(limit int) ([]*HistoryEntry, error)
    SearchHistory(query string) ([]*HistoryEntry, error)
    GetEntry(id int) (*HistoryEntry, error)
    ClearHistory() error
    LoadHistory() error
    SaveHistory() error
}
```

#### 4. Enhanced Auto-completion System
```go
type CompletionProvider interface {
    CompleteContainerNames(prefix string) []string
    CompleteImageNames(prefix string) []string
    CompleteContainerPaths(containerID, prefix string) []string
    CompleteDockerCommands(prefix string) []string
    CompleteCommandOptions(command, prefix string) []string
}

type EnhancedCompleter struct {
    mappingEngine     MappingEngine
    contextManager    ContextManager
    aliasManager      AliasManager
    completionProvider CompletionProvider
}
```

#### 5. Command Mapping Engine
```go
type CommandMapping struct {
    ID            string   `json:"id" yaml:"id"`
    LinuxCommand  string   `json:"linux_command" yaml:"linux_command"`
    DockerCommand string   `json:"docker_command" yaml:"docker_command"`
    Category      string   `json:"category" yaml:"category"`
    Description   string   `json:"description" yaml:"description"`
    LinuxExample  string   `json:"linux_example" yaml:"linux_example"`
    DockerExample string   `json:"docker_example" yaml:"docker_example"`
    Notes         []string `json:"notes" yaml:"notes"`
    Warnings      []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

type MappingEngine interface {
    LoadMappings() error
    FindByLinuxCommand(cmd string) (*CommandMapping, error)
    FindByDockerCommand(cmd string) (*CommandMapping, error)
    ListByCategory(category string) ([]*CommandMapping, error)
    SearchCommands(query string) ([]*CommandMapping, error)
}
```

#### 2. Command Parser
```go
type CommandParser interface {
    ParseCommand(input string) (*ParsedCommand, error)
    IsLinuxCommand(cmd string) bool
    IsDockerCommand(cmd string) bool
}

type ParsedCommand struct {
    Command   string
    Args      []string
    Options   map[string]string
    IsDocker  bool
    IsLinux   bool
}
```

#### 3. Shell Executor
```go
type ShellExecutor interface {
    Execute(cmd *ParsedCommand) (*ExecutionResult, error)
    ExecuteWithMapping(mapping *CommandMapping, args []string) (*ExecutionResult, error)
    DryRun(cmd *ParsedCommand) (string, error)
}

type ExecutionResult struct {
    Command    string
    Output     string
    Error      string
    ExitCode   int
    Duration   time.Duration
}
```

#### 4. Internationalization (i18n)
```go
type I18nManager interface {
    LoadMessages(lang string) error
    GetMessage(key string, args ...interface{}) string
    SetLanguage(lang string) error
    GetSupportedLanguages() []string
}

type LocalizedCommandMapping struct {
    CommandMapping
    LocalizedDescription map[string]string `json:"localized_description" yaml:"localized_description"`
    LocalizedNotes       map[string][]string `json:"localized_notes" yaml:"localized_notes"`
    LocalizedWarnings    map[string][]string `json:"localized_warnings,omitempty" yaml:"localized_warnings,omitempty"`
}
```

### Module Structure
```
dockernaut/
├── cmd/
│   ├── root.go          # Cobra root command
│   ├── interactive.go   # Interactive shell mode
│   ├── mapping.go       # Command mapping utilities
│   ├── alias.go         # Alias management commands
│   └── version.go       # Version command
├── internal/
│   ├── engine/          # Mapping engine
│   ├── parser/          # Command parser
│   ├── executor/        # Command executor
│   ├── config/          # Configuration management
│   ├── i18n/            # Internationalization
│   ├── alias/           # Alias management
│   ├── context/         # Container context management
│   ├── history/         # Command history management
│   └── completion/      # Enhanced auto-completion
├── data/
│   ├── mappings.yaml    # Command mappings data
│   ├── aliases.yaml     # User-defined aliases
│   └── locales/         # i18n message files
│       ├── en.yaml      # English messages
│       └── ja.yaml      # Japanese messages
├── .dockernaut/         # User data directory
│   ├── history.json     # Command history
│   ├── context.json     # Current container context
│   └── aliases.yaml     # User aliases (fallback)
└── main.go
```

## Data Models

### Command Categories
1. **リスト表示** (List Operations)
   - ls ↔ docker ps, docker images
   - ls -a ↔ docker ps -a
   - ls -l ↔ docker images
   - find ↔ docker search

2. **プロセス管理** (Process Management)
   - ps ↔ docker ps
   - kill ↔ docker kill, docker stop
   - kill -9 ↔ docker kill
   - top ↔ docker stats

3. **ログ・モニタリング** (Logs & Monitoring)
   - tail -f ↔ docker logs -f
   - logs ↔ docker logs
   - tail -n ↔ docker logs --tail
   - grep ↔ docker logs | grep

4. **ファイル操作** (File Operations)
   - cd ↔ docker exec -it bash (context switch)
   - cp ↔ docker cp
   - exec ↔ docker exec -it
   - vi ↔ docker exec -it vi
   - rm ↔ docker rm, docker rmi
   - mkdir ↔ docker volume create

5. **システム情報** (System Information)
   - df ↔ docker system df
   - free ↔ docker stats
   - uname ↔ docker version

6. **ネットワーク** (Network)
   - netstat ↔ docker port
   - ping ↔ docker exec ... ping

### Extended Data Models

#### Alias Data Structure
```yaml
# aliases.yaml
aliases:
  - name: "ll"
    command: "ls -la"
    description: "Long listing format"
    created_at: "2024-01-01T00:00:00Z"
  - name: "h"
    command: "history"
    description: "Show command history"
    created_at: "2024-01-01T00:00:00Z"
  - name: "dps"
    command: "docker ps"
    description: "List running containers"
    created_at: "2024-01-01T00:00:00Z"

standard_aliases:
  - name: "ll"
    command: "ls -la"
    description: "Long listing format"
  - name: "la"
    command: "ls -a"
    description: "List all files including hidden"
  - name: "h"
    command: "history"
    description: "Show command history"
```

#### Context Data Structure
```yaml
# context.json
{
  "current_container": {
    "container_id": "abc123def456",
    "container_name": "my-app",
    "image": "node:16-alpine",
    "status": "running",
    "set_at": "2024-01-01T12:00:00Z"
  },
  "recent_containers": [
    {
      "container_id": "abc123def456",
      "container_name": "my-app",
      "last_used": "2024-01-01T12:00:00Z"
    }
  ]
}
```

#### History Data Structure
```json
// history.json
{
  "entries": [
    {
      "id": 1,
      "command": "docker ps",
      "timestamp": "2024-01-01T12:00:00Z",
      "exit_code": 0,
      "duration": "150ms"
    },
    {
      "id": 2,
      "command": "cd my-app",
      "timestamp": "2024-01-01T12:01:00Z",
      "exit_code": 0,
      "duration": "50ms"
    }
  ],
  "max_entries": 1000
}
```

### Sample Data Structure (YAML)
```yaml
mappings:
  - id: "ls-docker-ps"
    linux_command: "ls"
    docker_command: "docker ps"
    category: "list-operations"
    description: "リスト表示 - 実行中のプロセス/コンテナを表示"
    linux_example: "ls -la"
    docker_example: "docker ps -a"
    notes:
      - "docker psはコンテナのみを表示"
      - "-aオプションで停止中のコンテナも表示"
    localized_description:
      en: "List display - Show running processes/containers"
      ja: "リスト表示 - 実行中のプロセス/コンテナを表示"
    localized_notes:
      en:
        - "docker ps shows only containers"
        - "-a option shows stopped containers too"
      ja:
        - "docker psはコンテナのみを表示"
        - "-aオプションで停止中のコンテナも表示"
  - id: "ps-docker-ps"
    linux_command: "ps"
    docker_command: "docker ps"
    category: "process-management"
    description: "プロセス一覧表示"
    linux_example: "ps aux"
    docker_example: "docker ps -a"
    notes:
      - "psは全プロセス、docker psはコンテナのみ"
    localized_description:
      en: "Process list display"
      ja: "プロセス一覧表示"
    localized_notes:
      en:
        - "ps shows all processes, docker ps shows containers only"
      ja:
        - "psは全プロセス、docker psはコンテナのみ"
```

### Configuration Structure
```yaml
# config.yaml
shell:
  prompt: "dockernaut> "
  history_size: 1000
  auto_complete: true
  
mapping:
  data_file: "data/mappings.yaml"
  cache_enabled: true
  
docker:
  default_options: []
  dry_run_mode: false
  
display:
  show_warnings: true
  color_output: true
  verbose_mode: false

i18n:
  default_language: "ja"
  supported_languages: ["ja", "en"]
  locale_dir: "data/locales"
  fallback_language: "en"
```

### Locale Message Files
```yaml
# data/locales/ja.yaml
messages:
  welcome: "Dockernaut へようこそ"
  command_not_found: "コマンドが見つかりません: %s"
  mapping_found: "マッピングが見つかりました"
  executing_command: "コマンドを実行中: %s"
  docker_not_available: "Dockerが利用できません"
  
help:
  usage: "使用方法"
  examples: "例"
  options: "オプション"

categories:
  list-operations: "リスト表示"
  process-management: "プロセス管理"
  file-operations: "ファイル操作"
  system-information: "システム情報"
  network: "ネットワーク"
```

```yaml
# data/locales/en.yaml
messages:
  welcome: "Welcome to Dockernaut"
  command_not_found: "Command not found: %s"
  mapping_found: "Mapping found"
  executing_command: "Executing command: %s"
  docker_not_available: "Docker is not available"
  
help:
  usage: "Usage"
  examples: "Examples"
  options: "Options"

categories:
  list-operations: "List Operations"
  process-management: "Process Management"
  file-operations: "File Operations"
  system-information: "System Information"
  network: "Network"
```
```

## Error Handling

### CLI Error Handling
1. **コマンド解析エラー**: 不正なコマンド構文の適切なエラーメッセージ表示
2. **マッピングデータエラー**: YAML/JSON読み込み失敗時のフォールバック処理
3. **Docker実行エラー**: Docker CLI実行失敗時のエラー詳細表示
4. **設定ファイルエラー**: 設定ファイル不正時のデフォルト値使用

### Runtime Error Prevention
1. **入力検証**: コマンド引数の妥当性チェック
2. **Docker接続確認**: Docker daemon接続状態の事前チェック
3. **権限チェック**: 必要な実行権限の確認
4. **プラットフォーム対応**: OS固有の処理の適切な分岐

## Testing Strategy

### Unit Testing
- **パッケージテスト**: Go標準のtestingパッケージを使用
- **マッピングエンジンテスト**: コマンド検索・変換機能のテスト
- **パーサーテスト**: コマンド解析ロジックのテスト
- **設定管理テスト**: YAML/JSON設定読み込みのテスト

### Integration Testing
- **CLI統合テスト**: cobra-cliコマンド実行のテスト
- **Docker連携テスト**: Docker CLI実行とレスポンス処理のテスト
- **ファイルI/Oテスト**: 設定ファイル・マッピングデータの読み書きテスト

### Cross-Platform Testing
- **マルチプラットフォームテスト**: Windows, macOS, Linux環境での動作確認
- **パスハンドリングテスト**: OS固有のファイルパス処理のテスト
- **権限テスト**: 各OS環境での実行権限確認

### Performance Testing
- **起動時間テスト**: シェル起動速度の測定
- **コマンド実行速度テスト**: マッピング検索・実行時間の測定
- **メモリ使用量テスト**: 長時間実行時のメモリリーク確認

### Test Data
- **テストマッピングデータ**: 開発・テスト用のサンプルコマンドマッピング
- **エッジケーステスト**: 特殊文字、長いコマンド名、不正入力等のテストケース
- **モックDocker環境**: Docker CLI実行をモックするテストヘルパー