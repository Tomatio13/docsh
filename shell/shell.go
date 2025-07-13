package shell

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"cherrysh/config"
	"cherrysh/i18n"
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

	// Windows環境の初期化
	shell.initializeWindowsEnvironment()

	// Windows ANSI サポートを有効化
	enableWindowsAnsiSupport()

	// 設定ファイルを読み込み
	if err := cfg.LoadConfigFile(); err != nil {
		fmt.Printf(i18n.T("shell.config_load_warning")+"\n", err)
	}

	return shell
}

func (s *Shell) Start() error {
	fmt.Printf(i18n.T("app.welcome"))

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
			fmt.Println(i18n.T("app.goodbye"))
			break
		}

		if err := s.executeCommand(input); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("app.error")+"\n", err)
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
	case "lang":
		return s.handleLangCommand(args)
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
		return fmt.Errorf(i18n.T("config.not_initialized"))
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
			fmt.Printf(i18n.T("theme.current_theme")+"\n", s.config.Theme)
		}
		return nil
	}

	themeName := args[0]
	if _, exists := themes.GetTheme(themeName); !exists {
		return fmt.Errorf(i18n.T("theme.not_found"), themeName)
	}

	if s.config != nil {
		s.config.Theme = themeName
		fmt.Printf(i18n.T("theme.theme_changed")+"\n", themeName)
	}

	return nil
}

func (s *Shell) handleLangCommand(args []string) error {
	if len(args) == 0 {
		// 現在の言語設定を表示
		currentLang := i18n.GetCurrentLanguage()
		availableLangs := i18n.GetAvailableLanguages()

		fmt.Printf(i18n.T("lang.current_language")+"\n", currentLang)
		fmt.Printf(i18n.T("lang.available_languages") + "\n")
		for _, lang := range availableLangs {
			fmt.Printf("  %s\n", lang)
		}
		return nil
	}

	newLang := args[0]
	availableLangs := i18n.GetAvailableLanguages()

	// 指定された言語が利用可能かチェック
	isValid := false
	for _, lang := range availableLangs {
		if lang == newLang {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf(i18n.T("lang.invalid_language"), newLang)
	}

	// 設定ファイルに言語設定を保存
	if s.config != nil {
		if err := s.config.SetLanguage(newLang); err != nil {
			return fmt.Errorf(i18n.T("lang.save_error"), err)
		}
	}

	// i18nを再初期化
	if err := i18n.Init(newLang); err != nil {
		return fmt.Errorf(i18n.T("lang.init_error"), err)
	}

	fmt.Printf(i18n.T("lang.language_changed")+"\n", newLang)
	fmt.Printf(i18n.T("lang.restart_notice") + "\n")

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
		return fmt.Errorf(i18n.T("git.unknown_command"), subcommand)
	}
}
