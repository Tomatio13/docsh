package shell

import (
	"fmt"
	"os"
	"strings"

	"cherrysh/config"
	"cherrysh/i18n"
	"cherrysh/themes"

	"github.com/c-bata/go-prompt"
)

type Shell struct {
	cwd     string
	config  *config.Config
	history []string
}

func NewShell(cfg *config.Config) *Shell {
	cwd, _ := os.Getwd()

	shell := &Shell{
		cwd:     cwd,
		config:  cfg,
		history: []string{},
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
	// 設定ファイルを再読み込み
	if err := s.config.LoadConfigFile(); err != nil {
		fmt.Printf(i18n.T("shell.config_load_warning"), err)
	}

	fmt.Print(i18n.T("app.welcome"))

	// go-promptを使用したインタラクティブプロンプト
	p := prompt.New(
		s.executor,
		s.Completer,
		prompt.OptionTitle("🌸 Cherry Shell"),
		prompt.OptionHistory(s.history),
		prompt.OptionLivePrefix(s.getLivePrefix),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionDescriptionBGColor(prompt.Black),
		prompt.OptionDescriptionTextColor(prompt.White),
		prompt.OptionScrollbarThumbColor(prompt.DarkGray),
		prompt.OptionScrollbarBGColor(prompt.Black),
		prompt.OptionMaxSuggestion(16),
	)
	p.Run()

	return nil
}

// executor はコマンド実行を処理します
func (s *Shell) executor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	// 履歴に追加（重複を避ける）
	if len(s.history) == 0 || s.history[len(s.history)-1] != input {
		s.history = append(s.history, input)
		// 履歴の上限を設定（例：1000件）
		if len(s.history) > 1000 {
			s.history = s.history[1:]
		}
	}

	if input == "exit" {
		fmt.Println(i18n.T("app.goodbye"))
		os.Exit(0)
	}

	if err := s.executeCommand(input); err != nil {
		fmt.Printf(i18n.T("app.error")+"\n", err)
	}
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
	case "config":
		if len(args) > 0 {
			switch args[0] {
			case "show":
				s.showConfig()
			default:
				fmt.Printf(i18n.T("config.unknown_command")+"\n", args[0])
			}
		} else {
			fmt.Println(i18n.T("config.usage"))
		}
		return nil
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

func (s *Shell) showConfig() {
	fmt.Println(i18n.T("config.show_header"))
	fmt.Printf(i18n.T("config.show_theme")+"\n", s.config.Theme)
	fmt.Printf(i18n.T("config.show_language")+"\n", s.config.Language)

	if s.config.GitHubUser != "" {
		fmt.Printf(i18n.T("config.show_github_user")+"\n", s.config.GitHubUser)
	}

	if s.config.GitHubToken != "" {
		fmt.Printf(i18n.T("config.show_github_token")+"\n", s.config.GitHubToken[:10])
	} else {
		fmt.Println(i18n.T("config.show_github_token_not_set"))
	}

	fmt.Printf(i18n.T("config.show_aliases_count")+"\n", len(s.config.Aliases))
	if len(s.config.Aliases) > 0 {
		fmt.Println(i18n.T("config.show_aliases_header"))
		for name, command := range s.config.Aliases {
			fmt.Printf(i18n.T("config.show_alias_item")+"\n", name, command)
		}
	}
}

// getLivePrefix は動的なプロンプトプレフィックスを返します
func (s *Shell) getLivePrefix() (string, bool) {
	return s.buildPrompt(), true
}
