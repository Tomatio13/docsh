package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"gopkg.in/yaml.v2"
)

// Localizer はメッセージの国際化を管理する構造体
type Localizer struct {
	language string
	messages map[string]string
}

// 現在のローカライザーインスタンス
var currentLocalizer *Localizer

// 組み込みメッセージ（フォールバック用）
var embeddedMessages = map[string]map[string]string{
	"en": {
		"app.title":                   "🌸 Cherry Shell v1.0.0 - Beautiful & Simple Shell 🌸",
		"app.description":             "Named after the cherry blossom shell (Sakura-gai) - small but beautiful",
		"app.exit_instruction":        "Type 'exit' to quit",
		"app.welcome":                 "Welcome to Cherry Shell! 🌸 Type 'exit' to quit.\n",
		"app.goodbye":                 "Goodbye!",
		"app.error":                   "Error: %v",
		"shell.runtime_info":          "=== 🌸 Cherry Shell 🌸 ===",
		"shell.runtime_os":            "Runtime OS: %s",
		"shell.runtime_arch":          "Runtime ARCH: %s",
		"shell.runtime_separator":     "==========================",
		"shell.config_load_warning":   "Warning: Could not load config file: %v",
		"config.not_initialized":      "configuration not initialized",
		"config.alias_created":        "Alias '%s' created",
		"config.alias_list_header":    "Current aliases:",
		"config.alias_parse_error":    "invalid alias format. Use: alias name=command",
		"config.parse_error":          "Config parse error at line %d: %v",
		"theme.current_theme":         "Current theme: %s",
		"theme.theme_changed":         "Theme changed to: %s",
		"theme.not_found":             "theme '%s' not found",
		"theme.list_header":           "Available themes:",
		"lang.current_language":       "Current language: %s",
		"lang.available_languages":    "Available languages:",
		"lang.invalid_language":       "invalid language '%s'",
		"lang.save_error":             "failed to save language setting: %v",
		"lang.init_error":             "failed to initialize language: %v",
		"lang.language_changed":       "Language changed to: %s",
		"lang.restart_notice":         "Note: Some messages may require shell restart to take effect",
		"git.status_header":           "Git Status:",
		"git.clean_working_directory": "Working directory is clean",
		"git.changed_files":           "Changed files:",
		"git.add_error":               "Error adding file %s: %v",
		"git.add_success":             "Added: %s",
		"git.commit_created":          "Commit created: %s",
		"git.commit_success":          "Committed: %s",
		"git.push_success":            "Pushed successfully",
		"git.pull_success":            "Pulled successfully",
		"git.already_up_to_date":      "Already up to date",
		"git.pull_completed":          "Pull completed",
		"git.cloning_repository":      "Cloning repository: %s → %s",
		"git.clone_completed":         "Clone completed: %s",
		"git.clone_success":           "Cloned to: %s",
		"git.unknown_command":         "unknown git command: %s",
		"git.help_title":              "Git Commands:",
		"git.help_header":             "Git Commands:",
		"git.help_help":               "  help - Show this help message",
		"git.help_status":             "  status - Show repository status",
		"git.help_add":                "  add <file> - Add file to staging",
		"git.help_commit":             "  commit -m <message> - Commit changes",
		"git.help_push":               "  push - Push to remote",
		"git.help_pull":               "  pull - Pull from remote",
		"git.help_log":                "  log - Show commit history",
		"git.help_clone":              "  clone <url> - Clone repository",
		"windows.cat_error":           "Error reading file %s: %v",
		"windows.copy_usage":          "Usage: copy <source> <destination>",
		"windows.copy_success":        "Copied %s to %s",
		"windows.copy_error":          "Error copying file: %v",
		"windows.move_usage":          "Usage: move <source> <destination>",
		"windows.move_success":        "Moved %s to %s",
		"windows.move_error":          "Error moving file: %v",
		"windows.delete_usage":        "Usage: del <file>",
		"windows.delete_success":      "Deleted: %s",
		"windows.delete_error":        "Error deleting file %s: %v",
		"windows.rm_error":            "Error deleting %s: %v",
		"windows.files_deleted":       "Deleted %d file(s)",
		"windows.files_copied":        "File(s) copied successfully",
		"windows.files_moved":         "File(s) moved successfully",
		"windows.mkdir_usage":         "Usage: mkdir <directory>",
		"windows.mkdir_success":       "Created directory: %s",
		"windows.mkdir_error":         "Error creating directory %s: %v",
		"windows.rmdir_usage":         "Usage: rmdir <directory>",
		"windows.rmdir_success":       "Removed directory: %s",
		"windows.rmdir_error":         "Error removing directory %s: %v",
		"windows.where_usage":         "Usage: where <command>",
		"windows.where_found":         "Found: %s",
		"windows.where_not_found":     "Command not found: %s",
		"docker.pull_image":           "Pulling image: %s",
		"docker.pull_success":         "Successfully pulled image: %s",
		"docker.pull_failed":          "Failed to pull image '%s': %v",
		"docker.start_container":      "Starting container: %s",
		"docker.start_success":        "Successfully started container: %s",
		"docker.start_failed":         "Failed to start container '%s': %v",
		"docker.start_already_running": "Container '%s' is already running",
		"docker.exec_command":         "Executing in container %s: %s",
		"docker.exec_failed":          "Failed to execute command in container '%s': %v",
		"docker.stop_container":       "Stopping container: %s",
		"docker.stop_success":         "Successfully stopped container: %s",
		"docker.stop_failed":          "Failed to stop container '%s': %v",
		"docker.stop_not_running":     "Container '%s' is not running",
		"docker.remove_container":     "Removing container: %s",
		"docker.remove_success":       "Successfully removed container: %s",
		"docker.remove_failed":        "Failed to remove container '%s': %v",
		"docker.remove_running_container": "Container '%s' is running. Stop it first or use --force",
		"docker.remove_image":         "Removing image: %s",
		"docker.remove_image_success": "Successfully removed image: %s",
		"docker.remove_image_failed":  "Failed to remove image '%s': %v",
		"docker.image_not_found":      "Image '%s' not found",
		"docker.container_not_found":  "Container '%s' not found",
		"docker.not_available":        "Docker is not available",
		"docker.image_name_required":  "Image name is required",
		"docker.container_name_required": "Container name is required",
		"docker.command_required":     "Command is required",
		"docker.error_checking_container": "Error checking container: %v",
		"docker.error_checking_container_status": "Error checking container status: %v",
		"docker.error_checking_image": "Error checking image: %v",
	},
	"ja": {
		"app.title":                   "🌸 Cherry Shell v1.0.0 - 美しくシンプルなシェル 🌸",
		"app.description":             "桜貝（Sakura-gai）にちなんで名付けられました - 小さくても美しい",
		"app.exit_instruction":        "終了するには 'exit' と入力してください",
		"app.welcome":                 "Cherry Shell へようこそ！ 🌸 終了するには 'exit' と入力してください。\n",
		"app.goodbye":                 "さようなら！",
		"app.error":                   "エラー: %v",
		"shell.runtime_info":          "=== 🌸 Cherry Shell 🌸 ===",
		"shell.runtime_os":            "実行OS: %s",
		"shell.runtime_arch":          "実行アーキテクチャ: %s",
		"shell.runtime_separator":     "==========================",
		"shell.config_load_warning":   "警告: 設定ファイルを読み込めませんでした: %v",
		"config.not_initialized":      "設定が初期化されていません",
		"config.alias_created":        "エイリアス '%s' を作成しました",
		"config.alias_list_header":    "現在のエイリアス:",
		"config.alias_parse_error":    "エイリアス形式が無効です。使用方法: alias name=command",
		"config.parse_error":          "設定ファイル %d 行目でエラー: %v",
		"theme.current_theme":         "現在のテーマ: %s",
		"theme.theme_changed":         "テーマを変更しました: %s",
		"theme.not_found":             "テーマ '%s' が見つかりません",
		"theme.list_header":           "利用可能なテーマ:",
		"lang.current_language":       "現在の言語: %s",
		"lang.available_languages":    "利用可能な言語:",
		"lang.invalid_language":       "無効な言語 '%s'",
		"lang.save_error":             "言語設定の保存に失敗しました: %v",
		"lang.init_error":             "言語の初期化に失敗しました: %v",
		"lang.language_changed":       "言語を変更しました: %s",
		"lang.restart_notice":         "注意: 一部のメッセージは次回起動時に反映されます",
		"git.status_header":           "Git ステータス:",
		"git.clean_working_directory": "ワーキングディレクトリはクリーンです",
		"git.changed_files":           "変更されたファイル:",
		"git.add_error":               "ファイル %s の追加エラー: %v",
		"git.add_success":             "追加しました: %s",
		"git.commit_created":          "コミットを作成しました: %s",
		"git.commit_success":          "コミットしました: %s",
		"git.push_success":            "プッシュが完了しました",
		"git.pull_success":            "プルが完了しました",
		"git.already_up_to_date":      "既に最新です",
		"git.pull_completed":          "プルが完了しました",
		"git.cloning_repository":      "リポジトリをクローンしています: %s → %s",
		"git.clone_completed":         "クローンが完了しました: %s",
		"git.clone_success":           "クローンしました: %s",
		"git.unknown_command":         "不明なgitコマンド: %s",
		"git.help_title":              "Gitコマンド:",
		"git.help_header":             "Gitコマンド:",
		"git.help_help":               "  help - このヘルプメッセージを表示",
		"git.help_status":             "  status - リポジトリの状態を表示",
		"git.help_add":                "  add <ファイル> - ファイルをステージングに追加",
		"git.help_commit":             "  commit -m <メッセージ> - 変更をコミット",
		"git.help_push":               "  push - リモートにプッシュ",
		"git.help_pull":               "  pull - リモートからプル",
		"git.help_log":                "  log - コミット履歴を表示",
		"git.help_clone":              "  clone <URL> - リポジトリをクローン",
		"windows.cat_error":           "ファイル %s の読み込みエラー: %v",
		"windows.copy_usage":          "使用方法: copy <コピー元> <コピー先>",
		"windows.copy_success":        "%s を %s にコピーしました",
		"windows.copy_error":          "ファイルのコピーエラー: %v",
		"windows.move_usage":          "使用方法: move <移動元> <移動先>",
		"windows.move_success":        "%s を %s に移動しました",
		"windows.move_error":          "ファイルの移動エラー: %v",
		"windows.delete_usage":        "使用方法: del <ファイル>",
		"windows.delete_success":      "削除しました: %s",
		"windows.delete_error":        "ファイル %s の削除エラー: %v",
		"windows.rm_error":            "%s の削除エラー: %v",
		"windows.files_deleted":       "%d 個のファイルを削除しました",
		"windows.files_copied":        "ファイルのコピーが完了しました",
		"windows.files_moved":         "ファイルの移動が完了しました",
		"windows.mkdir_usage":         "使用方法: mkdir <ディレクトリ>",
		"windows.mkdir_success":       "ディレクトリを作成しました: %s",
		"windows.mkdir_error":         "ディレクトリ %s の作成エラー: %v",
		"windows.rmdir_usage":         "使用方法: rmdir <ディレクトリ>",
		"windows.rmdir_success":       "ディレクトリを削除しました: %s",
		"windows.rmdir_error":         "ディレクトリ %s の削除エラー: %v",
		"windows.where_usage":         "使用方法: where <コマンド>",
		"windows.where_found":         "見つかりました: %s",
		"windows.where_not_found":     "コマンドが見つかりません: %s",
		"docker.pull_image":           "イメージをプル中: %s",
		"docker.pull_success":         "イメージのプルが完了しました: %s",
		"docker.pull_failed":          "イメージ '%s' のプルに失敗しました: %v",
		"docker.start_container":      "コンテナを開始中: %s",
		"docker.start_success":        "コンテナの開始が完了しました: %s",
		"docker.start_failed":         "コンテナ '%s' の開始に失敗しました: %v",
		"docker.start_already_running": "コンテナ '%s' は既に実行中です",
		"docker.exec_command":         "コンテナ %s でコマンドを実行中: %s",
		"docker.exec_failed":          "コンテナ '%s' でのコマンド実行に失敗しました: %v",
		"docker.stop_container":       "コンテナを停止中: %s",
		"docker.stop_success":         "コンテナの停止が完了しました: %s",
		"docker.stop_failed":          "コンテナ '%s' の停止に失敗しました: %v",
		"docker.stop_not_running":     "コンテナ '%s' は実行中ではありません",
		"docker.remove_container":     "コンテナを削除中: %s",
		"docker.remove_success":       "コンテナの削除が完了しました: %s",
		"docker.remove_failed":        "コンテナ '%s' の削除に失敗しました: %v",
		"docker.remove_running_container": "コンテナ '%s' は実行中です。まず停止するか --force を使用してください",
		"docker.remove_image":         "イメージを削除中: %s",
		"docker.remove_image_success": "イメージの削除が完了しました: %s",
		"docker.remove_image_failed":  "イメージ '%s' の削除に失敗しました: %v",
		"docker.image_not_found":      "イメージ '%s' が見つかりません",
		"docker.container_not_found":  "コンテナ '%s' が見つかりません",
		"docker.not_available":        "Docker が利用できません",
		"docker.image_name_required":  "イメージ名が必要です",
		"docker.container_name_required": "コンテナ名が必要です",
		"docker.command_required":     "コマンドが必要です",
		"docker.error_checking_container": "コンテナの確認中にエラーが発生しました: %v",
		"docker.error_checking_container_status": "コンテナの状態確認中にエラーが発生しました: %v",
		"docker.error_checking_image": "イメージの確認中にエラーが発生しました: %v",
	},
}

// Init は指定された言語でローカライザーを初期化する
func Init(language string) error {
	localizer := &Localizer{
		language: language,
		messages: make(map[string]string),
	}

	// メッセージファイルを読み込む（失敗時は組み込みメッセージを使用）
	if err := localizer.loadMessages(); err != nil {
		// フォールバック: 組み込みメッセージを使用
		if embeddedMsgs, exists := embeddedMessages[language]; exists {
			localizer.messages = embeddedMsgs
		} else {
			// 指定された言語がない場合は英語を使用
			localizer.messages = embeddedMessages["en"]
		}
	}

	currentLocalizer = localizer
	return nil
}

// T はメッセージキーを翻訳する
func T(key string, args ...interface{}) string {
	if currentLocalizer == nil {
		// フォールバック: 英語で初期化を試行
		if err := Init("en"); err != nil {
			return key
		}
	}

	message, exists := currentLocalizer.messages[key]
	if !exists {
		return key
	}

	// 引数がある場合はフォーマット
	if len(args) > 0 {
		return fmt.Sprintf(message, args...)
	}

	return message
}

// GetCurrentLanguage は現在の言語を返す
func GetCurrentLanguage() string {
	if currentLocalizer == nil {
		return "en"
	}
	return currentLocalizer.language
}

// DetectLanguage はコマンドライン引数と環境変数から言語を検出する
func DetectLanguage(args []string) string {
	// コマンドライン引数から検出
	for i, arg := range args {
		if arg == "--lang" && i+1 < len(args) {
			return args[i+1]
		}
	}

	// 環境変数から検出
	if lang := os.Getenv("CHERRYSH_LANG"); lang != "" {
		return lang
	}

	// システムロケールから検出
	if lang := os.Getenv("LANG"); lang != "" {
		if strings.Contains(lang, "ja") {
			return "ja"
		}
	}

	// デフォルトは英語
	return "en"
}

func (l *Localizer) loadMessages() error {
	// 実行ファイルのディレクトリを取得
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execDir := filepath.Dir(execPath)

	// メッセージファイルのパスを構築
	messageFile := filepath.Join(execDir, "data", "locales", l.language+".yaml")

	// ファイルが存在しない場合は、プロジェクトディレクトリから読み込む
	if _, err := os.Stat(messageFile); os.IsNotExist(err) {
		// 開発環境用のパス
		messageFile = filepath.Join("data", "locales", l.language+".yaml")
	}

	// ファイルを読み込む
	data, err := os.ReadFile(messageFile)
	if err != nil {
		return fmt.Errorf("failed to read message file %s: %w", messageFile, err)
	}

	// YAMLをパース
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return fmt.Errorf("failed to parse message file %s: %w", messageFile, err)
	}

	// 階層的なYAMLデータをフラットなマップに変換
	l.messages = make(map[string]string)
	
	// 各トップレベルセクションを処理
	for sectionName, sectionData := range yamlData {
		if sectionMap, ok := sectionData.(map[interface{}]interface{}); ok {
			flattenYAMLInterface(sectionMap, sectionName, l.messages)
		} else if sectionMap, ok := sectionData.(map[string]interface{}); ok {
			flattenYAML(sectionMap, sectionName, l.messages)
		}
	}

	return nil
}

// flattenYAML は階層的なYAMLデータをフラットなマップに変換する
func flattenYAML(data map[string]interface{}, prefix string, result map[string]string) {
	for key, value := range data {
		newKey := key
		if prefix != "" {
			newKey = prefix + "." + key
		}
		
		switch v := value.(type) {
		case map[string]interface{}:
			flattenYAML(v, newKey, result)
		case string:
			result[newKey] = v
		default:
			result[newKey] = fmt.Sprintf("%v", v)
		}
	}
}

// flattenYAMLInterface は interface{}キーの階層的なYAMLデータをフラットなマップに変換する
func flattenYAMLInterface(data map[interface{}]interface{}, prefix string, result map[string]string) {
	for key, value := range data {
		keyStr := fmt.Sprintf("%v", key)
		newKey := keyStr
		if prefix != "" {
			newKey = prefix + "." + keyStr
		}
		
		switch v := value.(type) {
		case map[interface{}]interface{}:
			flattenYAMLInterface(v, newKey, result)
		case string:
			result[newKey] = v
		default:
			result[newKey] = fmt.Sprintf("%v", v)
		}
	}
}

// GetAvailableLanguages は利用可能な言語のリストを返す
func GetAvailableLanguages() []string {
	return []string{"en", "ja"}
}
