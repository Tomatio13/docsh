# 🐳 docsh - Docker Command Mapping Shell

<p align="center">
    <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
    <img src="https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey?style=for-the-badge" alt="Platform">
    <img src="https://img.shields.io/badge/i18n-English%20%7C%20Japanese-blue?style=for-the-badge" alt="i18n">
</p>

<p align="center">
    <a href="README.md"><img src="https://img.shields.io/badge/english-document-white.svg" alt="EN doc"></a>
    <a href="README_ja.md"><img src="https://img.shields.io/badge/ドキュメント-日本語-white.svg" alt="JA doc"/></a>
</p>

**docsh** is a Docker command mapping shell that simplifies Docker operations by providing intuitive command mappings and an interactive shell environment. Formerly known as "docknaut", this tool helps you manage Docker containers and images more efficiently.

## ✨ Features

- **🐳 Docker Command Mapping**: Simplified and intuitive Docker command aliases
- **🌍 Cross-platform**: Works on Windows, macOS, and Linux
- **⚡ Interactive Shell**: Built-in interactive command-line interface
- **🔧 Configurable**: YAML-based configuration with extensive customization
- **🔗 Alias Support**: Create custom command shortcuts
- **🌐 Internationalization**: English and Japanese language support
- **📜 Command History**: Persistent command history
- **🎨 Customizable Prompts**: Personalize your shell experience

## 📦 Installation

### 🚀 Binary Download

Download the latest release:

[![Latest Release](https://img.shields.io/github/v/release/your-username/docsh?style=for-the-badge)](https://github.com/your-username/docsh/releases/latest)

> **📥 Download from [Releases Page](https://github.com/your-username/docsh/releases)**

### 🛠️ Build from Source

```bash
git clone https://github.com/your-username/docsh.git
cd docsh
go build -o docsh main.go
```

Or use the build script for all platforms:

```bash
./build.sh
```

## 🚀 Usage

### Interactive Mode

```bash
# Start docsh interactive shell
./docsh
```

### Direct Command Execution

```bash
# Execute commands directly
./docsh ps                    # Docker ps
./docsh images               # Docker images
./docsh run nginx           # Docker run nginx
```

### Basic Docker Commands

```bash
# Container management
ps                          # List running containers
psa                         # List all containers
images                      # List images
logs <container>            # Show container logs
exec <container> <command>  # Execute command in container
stop <container>            # Stop container
rm <container>             # Remove container
rmi <image>                # Remove image

# System commands
system prune               # Clean up unused resources
network ls                 # List networks
volume ls                  # List volumes
```

### Built-in Aliases

```bash
# Standard aliases
ll                         # ls -la
la                         # ls -a
h                          # history

# Docker aliases
dps                        # docker ps
dpa                        # docker ps -a
di                         # docker images
dlog                       # docker logs
dlogf                      # docker logs -f
```

## 🌐 Language Support

docsh supports multiple languages with automatic detection:

### Command Line Options
```bash
./docsh --lang en          # English
./docsh --lang ja          # Japanese
```

### Environment Variables
```bash
export DOCSH_LANG=en       # English
export DOCSH_LANG=ja       # Japanese
./docsh
```

### System Locale
docsh automatically detects your system locale. If `LANG` environment variable is set to `ja_JP.UTF-8` or similar, it will use Japanese.

## ⚙️ Configuration

docsh uses a YAML configuration file located at `data/config.yaml`:

```yaml
shell:
  prompt: "🐳 docsh> "
  history_size: 1000
  auto_complete: true
  dry_run_mode: false
  show_mappings: true
  
mapping:
  data_file: "data/mappings.yaml"
  cache_enabled: true
  auto_suggest: true
  
docker:
  default_options: []
  timeout: 30
  auto_detect: true
  
display:
  show_warnings: true
  color_output: true
  verbose_mode: false
  show_examples: true
  show_descriptions: true

i18n:
  default_language: "ja"
  supported_languages: ["ja", "en"]
  locale_dir: "data/locales"
  fallback_language: "en"

features:
  aliases: true
  context_management: true
  history: true
  completion: true
  command_mapping: true
  git_integration: true

aliases:
  ll: "ls -la"
  la: "ls -a"
  h: "history"
  dps: "docker ps"
  dpa: "docker ps -a"
  di: "docker images"
```

### User Configuration

You can also use a traditional configuration file at `~/.docknautrc`:

```bash
# Language setting
LANG="en"

# GitHub authentication (for git operations)
GITHUB_TOKEN="ghp_your_token_here"
GITHUB_USER="your_username"

# Custom aliases
alias ll="ls -la"
alias la="ls -la"
alias myapp="docker run -d myapp:latest"
```

## 🛠️ Development

### Building

```bash
# Build for current platform
go build -o docsh main.go

# Build for all platforms
./build.sh

# Run tests
go test ./...
```

### Adding New Languages

1. Create a new message file at `data/locales/<lang>.yaml`
2. Translate all message keys
3. Add the language code to `i18n/i18n.go`
4. Update language detection logic if needed

### 📁 Project Structure

```
docsh/
├── main.go                 # Entry point
├── config/                 # Configuration management
│   ├── config.go          # Main config logic
│   ├── alias.go           # Alias handling
│   └── yaml.go            # YAML configuration
├── i18n/                   # Internationalization
│   └── i18n.go            # i18n management
├── shell/                  # Shell implementation
│   ├── shell.go           # Main shell logic
│   ├── command.go         # Command processing
│   └── prompt.go          # Prompt generation
├── tui/                    # Terminal UI components
├── data/                   # Configuration and data files
│   ├── config.yaml        # Main configuration
│   ├── mappings.yaml      # Docker command mappings
│   └── locales/           # Translation files
└── themes/                 # Theme system
    └── theme.go           # Theme definitions
```

## 🌍 Supported Languages

- **🇺🇸 English (en)**: Full support
- **🇯🇵 Japanese (ja)**: Full support

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

## 🙏 Acknowledgments

- Inspired by the simplicity and elegance of command-line tools
- Built with Go for cross-platform compatibility
- Docker community for continuous innovation

---

<p align="center">
🐳 <strong>docsh</strong> - Simplifying Docker operations, one command at a time.
</p>