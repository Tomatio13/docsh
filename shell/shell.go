package shell

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"docknaut/config"
	"docknaut/i18n"
	"docknaut/internal/engine"
	"docknaut/internal/executor"
	"docknaut/internal/parser"
	"docknaut/themes"

	"github.com/c-bata/go-prompt"
)

type Shell struct {
	cwd           string
	config        *config.Config
	history       []string
	mappingEngine engine.MappingEngine
	commandParser parser.CommandParser
	shellExecutor executor.ShellExecutor
	dataPath      string
}

func NewShell(cfg *config.Config, dataPath string) *Shell {
	cwd, _ := os.Getwd()

	// データパスを設定（デフォルトは ./data）
	if dataPath == "" {
		dataPath = "data"
	}

	// 新しいコンポーネントを初期化
	mappingEngine := engine.NewMappingEngine(dataPath)
	commandParser := parser.NewCommandParser()
	shellExecutor := executor.NewShellExecutor(mappingEngine)

	// マッピングデータを読み込み
	if err := mappingEngine.LoadMappings(); err != nil {
		fmt.Printf(i18n.T("shell.config_load_warning")+"\n", err)
	}

	shell := &Shell{
		cwd:           cwd,
		config:        cfg,
		history:       []string{},
		mappingEngine: mappingEngine,
		commandParser: commandParser,
		shellExecutor: shellExecutor,
		dataPath:      dataPath,
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

	// 起動バナーを表示
	if s.config.BannerEnabled {
		banner := themes.RenderBanner(s.config.BannerStyle)
		fmt.Print(banner)
	}

	fmt.Print(i18n.T("app.docker_only_welcome"))

	// go-promptを使用したインタラクティブプロンプト
	p := prompt.New(
		s.executor,
		s.Completer,
		prompt.OptionTitle("🐳 Docsh (Docker-Only)"),
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

// ExecuteCommand exposes the executeCommand method for external use
func (s *Shell) ExecuteCommand(input string) error {
	return s.executeCommand(input)
}

func (s *Shell) executeCommand(input string) error {
	// エイリアス展開
	if s.config != nil {
		input = s.config.ExpandAlias(input)
	}

	// コマンドをパース
	parsedCmd, err := s.commandParser.ParseCommand(input)
	if err != nil {
		return err
	}
	if parsedCmd == nil {
		return nil
	}

	command := parsedCmd.Command
	args := parsedCmd.Args

	// Docker専用シェルの内蔵コマンドのみ処理
	switch command {
	case "pwd":
		fmt.Println(s.getCurrentDir())
		return nil
	case "alias":
		return s.handleAliasCommand(args)
	case "theme":
		return s.handleThemeCommand(args)
	case "lang":
		return s.handleLangCommand(args)
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
	case "mapping":
		return s.handleMappingCommand(args)
	case "help":
		s.showHelp()
		return nil
	case "version":
		fmt.Println(i18n.T("app.docker_only_version"))
		return nil
	case "clear", "cls":
		fmt.Print("\033[2J\033[H")
		return nil
	// Docker lifecycle commands
	case "pull":
		if len(args) == 0 {
			return fmt.Errorf(i18n.T("docker.image_name_required"))
		}
		return s.pullImage(args[0])
	case "start":
		if len(args) == 0 {
			return fmt.Errorf(i18n.T("docker.container_name_required"))
		}
		return s.startContainer(args[0])
	case "exec":
		if len(args) < 2 {
			return fmt.Errorf(i18n.T("docker.container_name_required") + " and " + i18n.T("docker.command_required"))
		}
		return s.execInContainer(args[0], args[1:])
	case "stop":
		if len(args) == 0 {
			return fmt.Errorf(i18n.T("docker.container_name_required"))
		}
		return s.stopContainer(args[0])
	case "rm":
		if len(args) == 0 {
			return fmt.Errorf(i18n.T("docker.container_name_required"))
		}
		// Check for --force flag
		force := false
		containerName := args[0]
		if len(args) > 1 {
			for _, arg := range args[1:] {
				if arg == "--force" || arg == "-f" {
					force = true
					break
				}
			}
		}
		return s.removeContainer(containerName, force)
	case "rmi":
		if len(args) == 0 {
			return fmt.Errorf(i18n.T("docker.image_name_required"))
		}
		// Check for --force flag
		force := false
		imageName := args[0]
		if len(args) > 1 {
			for _, arg := range args[1:] {
				if arg == "--force" || arg == "-f" {
					force = true
					break
				}
			}
		}
		return s.removeImage(imageName, force)
	default:
		// Docker専用シェルモードでコマンド実行
		// ストリーミングコマンドの場合は特別な処理を行う
		isStreaming := isStreamingCommand(parsedCmd)
		if isStreaming {
			return s.executeStreamingCommandDirectly(parsedCmd)
		}

		// 通常のコマンド実行
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := s.shellExecutor.Execute(ctx, parsedCmd)
		if err != nil {
			// Docker専用シェルのエラーメッセージを表示
			fmt.Printf("❌ %s\n", result.Error)
			fmt.Println(i18n.T("app.docker_only_available_commands"))
			fmt.Println(i18n.T("app.docker_only_commands_list"))
			fmt.Println(i18n.T("app.docker_only_mapping_help"))
			return nil
		}

		// 結果を表示
		if result.Output != "" {
			fmt.Print(result.Output)
		}

		// マッピング情報を表示
		if result.Mapping != nil {
			fmt.Printf("✅ %s -> %s\n", result.Mapping.LinuxCommand, result.Mapping.DockerCommand)
		}

		return nil
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

	// Docker の状態を表示
	fmt.Println()
	fmt.Println(i18n.T("config.docker_only_mode_enabled"))
	if s.shellExecutor.IsDockerAvailable() {
		fmt.Println(i18n.T("config.docker_available"))
	} else {
		fmt.Println(i18n.T("config.docker_not_available"))
	}

	// マッピング統計を表示
	mappings := s.mappingEngine.GetAllMappings()
	fmt.Printf("\n"+i18n.T("config.total_mappings")+"\n", len(mappings))
	categories := s.mappingEngine.GetCategories()
	fmt.Printf(i18n.T("config.available_categories")+"\n", strings.Join(categories, ", "))
	fmt.Println("\n" + i18n.T("config.linux_commands_disabled"))
}

// getLivePrefix は動的なプロンプトプレフィックスを返します
func (s *Shell) getLivePrefix() (string, bool) {
	return s.buildPrompt(), true
}

// handleMappingCommand は mapping コマンドを処理します
func (s *Shell) handleMappingCommand(args []string) error {
	if len(args) == 0 {
		// 全マッピング一覧を表示
		s.listAllMappings()
		return nil
	}

	switch args[0] {
	case "list":
		if len(args) > 1 {
			return s.listMappingsByCategory(args[1])
		}
		s.listAllMappings()
		return nil
	case "search":
		if len(args) > 1 {
			return s.searchMappings(strings.Join(args[1:], " "))
		}
		return fmt.Errorf(i18n.T("mappings.search_no_results"), "")
	case "show":
		if len(args) > 1 {
			return s.showMapping(args[1])
		}
		return fmt.Errorf("show requires a command name")
	default:
		return fmt.Errorf("unknown mapping command: %s", args[0])
	}
}

// listAllMappings は全マッピングを一覧表示します
func (s *Shell) listAllMappings() {
	fmt.Println(i18n.T("commands.mapping_help"))
	fmt.Println()

	categories := s.mappingEngine.GetCategories()
	for _, category := range categories {
		fmt.Printf("=== %s ===\n", i18n.T("categories."+category))
		categoryMappings, _ := s.mappingEngine.ListByCategory(category)
		for _, mapping := range categoryMappings {
			fmt.Printf("  %s -> %s\n", mapping.LinuxCommand, mapping.DockerCommand)
			if mapping.Description != "" {
				fmt.Printf("    %s\n", mapping.Description)
			}
		}
		fmt.Println()
	}
}

// listMappingsByCategory はカテゴリ別マッピングを表示します
func (s *Shell) listMappingsByCategory(category string) error {
	mappings, err := s.mappingEngine.ListByCategory(category)
	if err != nil {
		return err
	}

	if len(mappings) == 0 {
		fmt.Printf(i18n.T("mappings.category_not_found")+"\n", category)
		return nil
	}

	fmt.Printf("Mappings for category '%s':\n\n", i18n.T("categories."+category))
	for _, mapping := range mappings {
		fmt.Printf("%s -> %s\n", mapping.LinuxCommand, mapping.DockerCommand)
		fmt.Printf("  %s: %s\n", i18n.T("help.description"), mapping.Description)
		fmt.Printf("  %s: %s\n", i18n.T("help.examples"), mapping.DockerExample)
		if len(mapping.Notes) > 0 {
			fmt.Printf("  %s: %s\n", i18n.T("help.notes"), strings.Join(mapping.Notes, ", "))
		}
		fmt.Println()
	}
	return nil
}

// searchMappings はマッピングを検索します
func (s *Shell) searchMappings(query string) error {
	mappings, err := s.mappingEngine.SearchCommands(query)
	if err != nil {
		return err
	}

	if len(mappings) == 0 {
		fmt.Printf(i18n.T("mappings.search_no_results")+"\n", query)
		return nil
	}

	fmt.Printf("Search results for '%s':\n\n", query)
	for _, mapping := range mappings {
		fmt.Printf("%s -> %s\n", mapping.LinuxCommand, mapping.DockerCommand)
		fmt.Printf("  Category: %s\n", i18n.T("categories."+mapping.Category))
		fmt.Printf("  %s: %s\n", i18n.T("help.description"), mapping.Description)
		fmt.Println()
	}
	return nil
}

// showMapping は特定のマッピング詳細を表示します
func (s *Shell) showMapping(command string) error {
	mapping, err := s.mappingEngine.FindByLinuxCommand(command)
	if err != nil {
		mapping, err = s.mappingEngine.FindByDockerCommand(command)
		if err != nil {
			return fmt.Errorf(i18n.T("mappings.not_found"), command)
		}
	}

	fmt.Printf("Mapping Details for '%s':\n\n", command)
	fmt.Printf("Linux Command: %s\n", mapping.LinuxCommand)
	fmt.Printf("Docker Command: %s\n", mapping.DockerCommand)
	fmt.Printf("Category: %s\n", i18n.T("categories."+mapping.Category))
	fmt.Printf("%s: %s\n", i18n.T("help.description"), mapping.Description)
	fmt.Printf("Linux Example: %s\n", mapping.LinuxExample)
	fmt.Printf("Docker Example: %s\n", mapping.DockerExample)

	if len(mapping.Notes) > 0 {
		fmt.Printf("\n%s:\n", i18n.T("help.notes"))
		for _, note := range mapping.Notes {
			fmt.Printf("  - %s\n", note)
		}
	}

	if len(mapping.Warnings) > 0 {
		fmt.Printf("\n%s:\n", i18n.T("help.warnings"))
		for _, warning := range mapping.Warnings {
			fmt.Printf("  ⚠️  %s\n", warning)
		}
	}

	return nil
}

// isStreamingCommand ストリーミングコマンドかどうかを判定します
func isStreamingCommand(parsedCmd *parser.ParsedCommand) bool {
	// 最初にLinuxコマンド名で判定（最優先）
	switch parsedCmd.Command {
	case "free":
		// free は docker stats --no-stream にマッピングされる（非ストリーミング）
		return false
	case "top":
		// top は docker stats にマッピングされる（ストリーミング）
		return true
	}

	// ParsedCommandをDockerコマンド配列に変換
	var dockerCmd []string

	// tail -f などのLinuxコマンドをDockerコマンドに変換してチェック
	if parsedCmd.Command == "tail" {
		if _, hasF := parsedCmd.Options["f"]; hasF {
			// tail -f は docker logs -f にマッピングされる
			dockerCmd = []string{"docker", "logs", "-f"}
			dockerCmd = append(dockerCmd, parsedCmd.Args...)
		}
	} else if parsedCmd.Command == "docker" {
		// 既にDockerコマンドの場合
		dockerCmd = append([]string{"docker"}, parsedCmd.Args...)
	} else {
		// その他のコマンドの場合、基本的にストリーミングではない
		return false
	}

	// executor packageのisStreamingCommand関数を利用
	if len(dockerCmd) >= 2 {
		// docker logs -f のチェック
		if dockerCmd[1] == "logs" {
			for _, arg := range dockerCmd[2:] {
				if arg == "-f" || arg == "--follow" {
					return true
				}
			}
		}

		// docker attach のチェック
		if dockerCmd[1] == "attach" {
			return true
		}

		// docker exec with interactive/tty flags のチェック
		if dockerCmd[1] == "exec" {
			for _, arg := range dockerCmd[2:] {
				if arg == "-it" || arg == "-i" || arg == "-t" {
					return true
				}
			}
		}

		// docker stats のチェック (--no-stream がない場合はストリーミング)
		if dockerCmd[1] == "stats" {
			for _, arg := range dockerCmd[2:] {
				if arg == "--no-stream" {
					return false // --no-stream があればストリーミングではない
				}
			}
			return true // デフォルトのdocker statsはストリーミング
		}
	}

	return false
}

// showHelp はヘルプを表示します
func (s *Shell) showHelp() {
	fmt.Println(i18n.T("commands.docker_only_help_title"))
	fmt.Println(i18n.T("commands.docker_only_help_description"))
	fmt.Println()
	fmt.Println(i18n.T("commands.docker_only_mappings_title"))
	fmt.Println(i18n.T("commands.examples_header"))
	fmt.Println(i18n.T("commands.examples_ls"))
	fmt.Println(i18n.T("commands.examples_ps"))
	fmt.Println(i18n.T("commands.examples_kill"))
	fmt.Println(i18n.T("commands.examples_rm"))
	fmt.Println(i18n.T("commands.examples_tail"))
	fmt.Println(i18n.T("commands.examples_cd"))
	fmt.Println()
	fmt.Println(i18n.T("commands.docker_only_docker_commands_title"))
	fmt.Println(i18n.T("commands.docker_commands_note"))
	fmt.Println()
	fmt.Println(i18n.T("commands.lifecycle_header"))
	fmt.Println(i18n.T("commands.lifecycle_pull"))
	fmt.Println(i18n.T("commands.lifecycle_start"))
	fmt.Println(i18n.T("commands.lifecycle_stop"))
	fmt.Println(i18n.T("commands.lifecycle_exec"))
	fmt.Println(i18n.T("commands.lifecycle_rm"))
	fmt.Println(i18n.T("commands.lifecycle_rmi"))
	fmt.Println()
	fmt.Println(i18n.T("commands.docker_only_builtin_commands_title"))
	fmt.Println("  help                    " + i18n.T("help.usage"))
	fmt.Println("  mapping [list|search|show] " + i18n.T("commands.mapping_help"))
	fmt.Println("  alias <name>=<command>  " + i18n.T("commands.alias_help"))
	fmt.Println("  theme [name]            " + i18n.T("commands.theme_help"))
	fmt.Println("  config [show]           " + i18n.T("commands.config_help"))
	fmt.Println("  exit                    " + i18n.T("commands.exit_help"))
	fmt.Println()
	fmt.Println(i18n.T("commands.docker_only_more_info_title"))
	fmt.Println("  mapping list            " + i18n.T("commands.mapping_list"))
	fmt.Println("  mapping search <query>  " + i18n.T("commands.mapping_search"))
	fmt.Println("  mapping show <command>  " + i18n.T("commands.mapping_show"))
	fmt.Println()
	fmt.Println(i18n.T("commands.docker_only_note_title") + " " + i18n.T("commands.docker_only_note_message"))
}

// executeStreamingCommandDirectly はgo-promptをバイパスしてストリーミングコマンドを直接実行します
func (s *Shell) executeStreamingCommandDirectly(parsedCmd *parser.ParsedCommand) error {
	// マッピングを解決
	var dockerCmd []string
	var mapping *engine.CommandMapping

	if parsedCmd.Command == "tail" && parsedCmd.Options["f"] == "true" {
		// tail -f をdocker logs -f にマッピング
		var err error
		mapping, err = s.mappingEngine.FindByLinuxCommand("tail -f")
		if err != nil {
			fmt.Printf("❌ tail -f mapping not found: %s\n", err.Error())
			return err
		}
		dockerCmd = strings.Fields(mapping.DockerCommand)
		dockerCmd = append(dockerCmd, parsedCmd.Args...)
	} else if parsedCmd.Command == "docker" {
		// 直接Dockerコマンド
		dockerCmd = append([]string{"docker"}, parsedCmd.Args...)
	} else {
		return fmt.Errorf("unsupported streaming command: %s", parsedCmd.Command)
	}

	fmt.Printf(i18n.T("app.executing")+"\n", strings.Join(dockerCmd, " "))
	if mapping != nil {
		fmt.Printf(i18n.T("app.mapping_applied")+"\n", mapping.LinuxCommand, mapping.DockerCommand)
	}
	fmt.Println(i18n.T("app.stream_stop_tip"))

	// コンテキストでgoroutineの協調的終了を制御
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Dockerコマンドを作成
	cmd := exec.CommandContext(ctx, dockerCmd[0], dockerCmd[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// パイプを作成してstdin/stdout/stderrを制御
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// コマンドを開始
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// 複数の終了監視を並行実行
	terminationChan := make(chan string, 5)

	// 1. 標準的なシグナル処理
	go s.watchForSignals(ctx, terminationChan)

	// 2. プロセス完了監視
	go func() {
		err := cmd.Wait()
		if err != nil {
			terminationChan <- fmt.Sprintf("process_error:%v", err)
		} else {
			terminationChan <- "process_completed"
		}
	}()

	// 3. 標準入力監視（exit コマンド用）
	go s.watchForStdinExit(ctx, terminationChan)

	// 4. 緊急時のプロセス監視
	go s.emergencyProcessMonitor(ctx, cmd, terminationChan)

	// 5. タイムアウト監視（極端に長い場合の保護）
	go func() {
		select {
		case <-time.After(30 * time.Minute):
			terminationChan <- "timeout"
		case <-ctx.Done():
			return
		}
	}()

	// 終了理由を待機
	reason := <-terminationChan

	// 全てのgoroutineを停止
	cancel()

	// プロセス終了処理
	s.cleanupProcess(cmd, stdin)

	switch {
	case strings.HasPrefix(reason, "signal"):
		fmt.Println(i18n.T("app.command_stopped_signal"))
	case reason == "stdin_exit":
		fmt.Println(i18n.T("app.command_stopped"))
	case reason == "stdin_force_kill":
		fmt.Println(i18n.T("app.command_force_killed"))
	case reason == "stdin_stop":
		fmt.Println(i18n.T("app.command_stopped_manual"))
	case reason == "process_completed":
		fmt.Println(i18n.T("app.command_completed"))
	case strings.HasPrefix(reason, "process_error"):
		fmt.Printf(i18n.T("app.command_failed_reason")+"\n", strings.TrimPrefix(reason, "process_error:"))
	case reason == "timeout":
		fmt.Println(i18n.T("app.command_timed_out"))
	case reason == "emergency":
		fmt.Println(i18n.T("app.command_stopped_alert"))
	case reason == "emergency_auto_terminate":
		fmt.Println(i18n.T("app.command_auto_terminated"))
	case reason == "process_already_exited":
		fmt.Println(i18n.T("app.command_exited"))
	}

	return nil
}

// watchForSignals は様々な方法でシグナルを監視します
func (s *Shell) watchForSignals(ctx context.Context, terminationChan chan string) {

	// 複数のシグナルチャンネルを作成
	sigChan1 := make(chan os.Signal, 5)
	sigChan2 := make(chan os.Signal, 5)
	sigChan3 := make(chan os.Signal, 5)

	// より多くのシグナルタイプを監視
	signal.Notify(sigChan1, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(sigChan2, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	signal.Notify(sigChan3, syscall.SIGTERM, os.Interrupt, syscall.SIGINT, syscall.SIGKILL)

	// コンテキストキャンセル時のクリーンアップ
	defer func() {
		signal.Stop(sigChan1)
		signal.Stop(sigChan2)
		signal.Stop(sigChan3)
	}()

	// 複数の監視goroutineを起動
	go func() {
		for {
			select {
			case sig := <-sigChan1:
				terminationChan <- fmt.Sprintf("signal_1:%v", sig)
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case sig := <-sigChan2:
				terminationChan <- fmt.Sprintf("signal_2:%v", sig)
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case sig := <-sigChan3:
				terminationChan <- fmt.Sprintf("signal_3:%v", sig)
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	// キーボード割り込み検出の別アプローチ
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// /proc/self/status で割り込み状態をチェック（Linuxのみ）
				if data, err := os.ReadFile("/proc/self/status"); err == nil {
					status := string(data)
					if strings.Contains(status, "SigPnd:") && !strings.Contains(status, "SigPnd:\t0000000000000000") {
						terminationChan <- "signal_proc_check"
						return
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// 最後の手段：定期的にプロセスの親プロセスをチェック
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		originalPpid := os.Getppid()

		for {
			select {
			case <-ticker.C:
				currentPpid := os.Getppid()
				if currentPpid != originalPpid {
					terminationChan <- "signal_orphan"
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// watchForStdinExit は標準入力から exit コマンドを監視します
func (s *Shell) watchForStdinExit(ctx context.Context, terminationChan chan string) {

	// バッファリングして文字列を組み立て
	go func() {
		inputBuffer := ""
		buffer := make([]byte, 1)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				// 1文字ずつ読み取り（非ブロッキング読み取りの試行）
				n, err := os.Stdin.Read(buffer)
				if err != nil {
					time.Sleep(10 * time.Millisecond) // CPU使用率を下げる
					continue
				}

				if n > 0 {
					char := string(buffer[0])

					// Enterキー（改行）の処理
					if char == "\n" || char == "\r" {
						if inputBuffer == "exit" || inputBuffer == "quit" || inputBuffer == "q" {
							terminationChan <- "stdin_exit"
							return
						} else if inputBuffer == "KILL" || inputBuffer == "kill" {
							terminationChan <- "stdin_force_kill"
							return
						} else if inputBuffer == "STOP" || inputBuffer == "stop" {
							terminationChan <- "stdin_stop"
							return
						}
						inputBuffer = "" // リセット
					} else if char >= " " && char <= "~" { // 印刷可能文字のみ
						inputBuffer += char
					}
				}
			}
		}
	}()
}

// emergencyProcessMonitor は緊急時のプロセス監視を行います
func (s *Shell) emergencyProcessMonitor(ctx context.Context, cmd *exec.Cmd, terminationChan chan string) {
	ticker := time.NewTicker(30 * time.Second) // 30秒間隔に変更
	defer ticker.Stop()

	emergencyCount := 0
	for {
		select {
		case <-ticker.C:
			emergencyCount++

			// プロセスがゾンビ化していないかチェック
			if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
				terminationChan <- "emergency"
				return
			}

			// 1分後に一度だけ代替手段を提示
			if emergencyCount == 2 { // 1分経過
				fmt.Println("\n💡 Alternative commands: 'exit', 'stop', or 'kill' + Enter")
			}

			// 長時間動作している場合の自動終了
			if emergencyCount >= 10 { // 5分経過
				terminationChan <- "emergency_auto_terminate"
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// cleanupProcess はプロセスを確実に終了させます
func (s *Shell) cleanupProcess(cmd *exec.Cmd, stdin interface{}) {
	if stdin != nil {
		if closer, ok := stdin.(io.WriteCloser); ok {
			closer.Close()
		}
	}

	if cmd.Process == nil {
		return
	}

	pid := cmd.Process.Pid

	// 段階的終了
	steps := []struct {
		name   string
		signal os.Signal
		wait   time.Duration
	}{
		{"SIGTERM to process group", syscall.SIGTERM, 200 * time.Millisecond},
		{"SIGKILL to process group", syscall.SIGKILL, 100 * time.Millisecond},
	}

	for _, step := range steps {
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			break
		}

		// プロセスグループに送信
		if err := syscall.Kill(-pid, step.signal.(syscall.Signal)); err != nil {
			// 個別プロセスに送信
			if step.signal == syscall.SIGKILL {
				cmd.Process.Kill()
			} else {
				cmd.Process.Signal(step.signal)
			}
		}

		time.Sleep(step.wait)
	}
}
