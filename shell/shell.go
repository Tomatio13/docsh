package shell

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"cherrysh/config"
	"cherrysh/themes"
)

type Shell struct {
	reader *bufio.Reader
	cwd    string
	config *config.Config
}

func NewShell() *Shell {
	cwd, _ := os.Getwd()
	cfg := config.NewConfig()

	shell := &Shell{
		reader: bufio.NewReader(os.Stdin),
		cwd:    cwd,
		config: cfg,
	}

	// 実行環境の情報を表示
	fmt.Printf("=== 🌸 Cherry Shell 🌸 ===\n")
	fmt.Printf("Runtime OS: %s\n", runtime.GOOS)
	fmt.Printf("Runtime ARCH: %s\n", runtime.GOARCH)
	fmt.Printf("==========================\n")

	// Windows環境の初期化
	shell.initializeWindowsEnvironment()

	// Windows ANSI サポートを有効化
	enableWindowsAnsiSupport()

	// 設定ファイルを読み込み
	if err := cfg.LoadConfigFile(); err != nil {
		fmt.Printf("Warning: Could not load config file: %v\n", err)
	}

	return shell
}

func (s *Shell) Start() error {
	fmt.Printf("\nWelcome to Cherry Shell! 🌸 Type 'exit' to quit.\n\n")

	for {
		s.showPrompt()

		input, err := s.reader.ReadString('\n')
		if err != nil {
			return err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if err := s.executeCommand(input); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	return nil
}

func (s *Shell) getCurrentDir() string {
	if cwd, err := os.Getwd(); err == nil {
		s.cwd = cwd
		return cwd
	}
	return s.cwd
}

func (s *Shell) executeCommand(input string) error {
	// エイリアス展開
	if s.config != nil {
		input = s.config.ExpandAlias(input)
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "cd":
		return s.changeDirectory(args)
	case "pwd":
		fmt.Println(s.getCurrentDir())
		return nil
	case "alias":
		return s.handleAliasCommand(args)
	case "theme":
		return s.handleThemeCommand(args)
	case "git":
		return s.handleGitCommand(args)
	default:
		// 内蔵コマンドかチェック
		if s.isBuiltinCommand(command) {
			return s.executeBuiltinCommand(command, args)
		}
		return s.executeExternalCommand(command, args)
	}
}

func (s *Shell) handleAliasCommand(args []string) error {
	if s.config == nil {
		return fmt.Errorf("config not initialized")
	}

	if len(args) == 0 {
		// 全エイリアスを表示
		s.config.ListAliases()
		return nil
	}

	// alias name=command の形式で新しいエイリアスを設定
	aliasString := strings.Join(args, " ")
	return s.config.ParseAlias(aliasString)
}

func (s *Shell) handleThemeCommand(args []string) error {
	if len(args) == 0 {
		// 利用可能なテーマ一覧を表示
		themes.ListThemes()
		if s.config != nil {
			fmt.Printf("Current theme: %s\n", s.config.Theme)
		}
		return nil
	}

	themeName := args[0]
	if _, exists := themes.GetTheme(themeName); !exists {
		return fmt.Errorf("theme '%s' not found", themeName)
	}

	if s.config != nil {
		s.config.Theme = themeName
		fmt.Printf("Theme changed to: %s\n", themeName)
	}

	return nil
}

func (s *Shell) handleGitCommand(args []string) error {
	if len(args) == 0 {
		s.gitHelp()
		return nil
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "status":
		return s.gitStatus(subArgs)
	case "add":
		return s.gitAdd(subArgs)
	case "commit":
		return s.gitCommit(subArgs)
	case "push":
		return s.gitPush(subArgs)
	case "pull":
		return s.gitPull(subArgs)
	case "log":
		return s.gitLog(subArgs)
	case "clone":
		return s.gitClone(subArgs)
	case "help", "-h", "--help":
		s.gitHelp()
		return nil
	default:
		return fmt.Errorf("不明なGitコマンド: %s", subcommand)
	}
}
