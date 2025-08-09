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

**docsh** is a Docker command mapping shell that simplifies Docker operations by providing intuitive command mappings and an interactive shell environment.

## ✨ Features

- **🐳 Docker Command Mapping**: Simplified and intuitive Docker command aliases
- **🌍 Cross-platform**: Works on Windows, macOS, and Linux
- **⚡ Interactive Shell**: Built-in interactive command-line interface
- **🔧 Configurable**: YAML-based configuration with extensive customization
- **🔗 Alias Support**: Create custom command shortcuts via `.docshrc` or YAML
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
# Execute commands directly (Linux-like → Docker mapping)
./docsh ls                   # mapped to: docker ps
./docsh kill myapp           # mapped to: docker stop myapp
./docsh rm myapp             # mapped to: docker rm myapp

# Or pass a Docker command explicitly
./docsh "docker ps"
./docsh "docker images"
```

### Common Operations (inside interactive shell)

```bash
# Container management (mapped/built-in)
ps                           # docker ps
logs <container>             # docker logs <container>
exec <container> <command>   # docker exec <container> <command>
stop <container>             # docker stop <container>
rm <container>               # docker rm <container>
rmi <image>                  # docker rmi <image>
```

### Docker Lifecycle Commands (from `help`)

```
🐳 Docker Lifecycle Commands:
  pull <image>            Pull image from registry
  start <container>       Start stopped container
  stop <container>        Stop running container
  exec <container> <cmd>  Execute command in container
  login <container>        Login to container (/bin/bash)
  rm [--force] <container> Remove container
  rmi [--force] <image>   Remove image
  log     <container>          Show container logs
  tail -f <container>          Follow container logs in real-time
  top                                       Show resource usage
  htop                                      Show resource usage (graph)
⚠️  Note: To exit 'tail -f' and 'top', type 'exit' while displaying.
```

### Project/Compose Operations (project commands)

Treat containers with Docker Compose labels as a "project" and operate by service.

- List all projects
  ```bash
  ps --by-project
  # or
  project ps
  ```

- List services in a project
  ```bash
  project <project> ps
  ```

- Show service logs (recommended)
  ```bash
  project <project> logs <service> -f --tail 100
  # Follows docker logs arg order: pass [OPTIONS] first, the container name is resolved and appended at the end
  ```

- Shorthand when the service name is globally unique
  ```bash
  project <service> logs -f --tail 100
  # If the same service name exists across multiple projects, you will get an ambiguity error
  ```

- Start project/service (Compose-aware)
  ```bash
  # Start the whole project (docker-compose.yml preferred if present)
  project <project> start

  # Start a specific service
  project <project> start <service>
  ```

- Restart/Stop (Compose-aware)
  ```bash
  project <project> restart [<service>]
  project <project> stop    [<service>]
  ```

Reference from help (excerpt):

```
🐳 Docker Compose Lifecycle Commands:
  project ps                          List services by project
  project <service> start             Start a specific service
  project <service> logs              Show logs of a specific service
  project <service> restart           Restart a specific service
  project <service> stop              Stop all services
  ps --by-project                     List containers grouped by project
```

### Aliases

Aliases can be defined in YAML (`data/config.yaml`) or in your `~/.docshrc`.

YAML example (shipped config):

```yaml
aliases:
  dps: "docker ps"
  dpa: "docker ps -a"
  di: "docker images"
```

`.docshrc` example (user overrides):

```bash
alias dps="docker ps"
alias dpa="docker ps -a"
alias di="docker images"
```

## 🌐 Language Support
Language is configured via your user config file only:

```bash
# ~/.docshrc
LANG="en"   # or "ja"
```

After editing `~/.docshrc`, restart `docsh` to apply the change.

Note: In the current version, command-line flags like `--lang` and environment
variables (e.g., `DOCSH_LANG`) are not used when `LANG` is set in `~/.docshrc`.

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

### ~/.docshrc sample

`docsh` reads user settings from `~/.docshrc` (if present). Example:

```bash
# Language (en or ja)
LANG="en"

# Theme (optional)
THEME="default"

# Aliases (optional)
alias dps="docker ps"
alias dpa="docker ps -a"
alias di="docker images"

# Example: quick-run helper
alias myapp="docker run -d --name myapp nginx:latest"
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
    ├── theme.go           # Prompt themes
    └── banner.go          # Startup banners
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