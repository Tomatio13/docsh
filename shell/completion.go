package shell

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
)

// Completer はコマンドライン補完を提供します
func (s *Shell) Completer(d prompt.Document) []prompt.Suggest {
	// カーソル位置までのテキストを取得
	beforeCursor := d.TextBeforeCursor()
	words := strings.Fields(beforeCursor)

	// 何も入力されていない場合は補完候補を表示しない
	if len(words) == 0 && beforeCursor == "" {
		return []prompt.Suggest{}
	}

	// 空白のみの場合も補完候補を表示しない
	if strings.TrimSpace(beforeCursor) == "" {
		return []prompt.Suggest{}
	}

	// 1文字だけの入力では補完候補を表示しない（タイピング中の過度な表示を防ぐ）
	if len(strings.TrimSpace(beforeCursor)) == 1 {
		return []prompt.Suggest{}
	}

	if len(words) == 0 {
		// 何も入力されていない場合はコマンド補完
		return s.completeCommands("")
	}

	// 最初の単語（コマンド）の補完
	if len(words) == 1 && !strings.HasSuffix(beforeCursor, " ") {
		// 2文字以上入力された場合のみコマンド補完を表示
		if len(words[0]) < 2 {
			return []prompt.Suggest{}
		}
		return s.completeCommands(words[0])
	}

	// 引数の補完（主にファイル/ディレクトリ）
	command := words[0]

	// 現在入力中の引数を取得
	var currentArg string
	if strings.HasSuffix(beforeCursor, " ") {
		currentArg = ""
	} else if len(words) > 1 {
		currentArg = words[len(words)-1]
	}

	
	switch command {
	case "cd":
		// Docker専用シェルでは cd はコンテナへのログイン
		return s.completeDockerContainers(currentArg, true) // 実行中のコンテナのみ
	case "cat", "cp", "mv":
		return s.completeFiles(currentArg)
	case "rm":
		// Docker専用シェルでは rm はコンテナ削除
		return s.completeDockerContainers(currentArg, false)
	case "rmi":
		// Docker専用シェルでは rmi はイメージ削除
		return s.completeDockerImages(currentArg)
	case "start":
		// Docker専用シェルでは start はコンテナ開始
		return s.completeDockerContainers(currentArg, false)
	case "stop":
		// Docker専用シェルでは stop はコンテナ停止
		return s.completeDockerContainers(currentArg, true)
	case "exec":
		// Docker専用シェルでは exec はコンテナ内コマンド実行
		return s.completeDockerContainers(currentArg, true)
	case "pull":
		// Docker専用シェルでは pull はイメージのプル（補完無効）
		return []prompt.Suggest{}
	case "ps":
		// Docker専用シェルでは ps は docker ps（引数不要なので補完無効）
		return []prompt.Suggest{}
	case "ls":
		// Docker専用シェルでは ls は docker images にマッピングされるが、
		// ディレクトリ操作も必要なので引き続きファイル補完を提供
		return s.completeFilesAndDirectories(currentArg)
	case "kill":
		// Docker専用シェルでは kill は docker stop（コンテナ停止）
		return s.completeDockerContainers(currentArg, true)
	
	// ログ系コマンド - コンテナ名を指定
	case "tail", "head", "grep":
		// tail -> docker logs, head -> docker logs --tail, grep -> docker logs | grep
		return s.completeDockerContainers(currentArg, false) // 全てのコンテナ
	
	// ファイル操作系コマンド - コンテナ内実行なので実行中コンテナのみ
	case "vi", "nano", "mkdir", "find", "locate":
		// docker exec -it が必要なコマンド
		return s.completeDockerContainers(currentArg, true) // 実行中コンテナのみ
	
	// ネットワーク系コマンド
	case "netstat":
		// docker port コマンド（コンテナ名指定）
		return s.completeDockerContainers(currentArg, false) // 全てのコンテナ
	
	// システム情報系コマンド - 引数不要または補完無効
	case "free", "top", "htop", "uname":
		// docker stats, docker version などは引数不要
		return []prompt.Suggest{}
	case "df", "du":
		// docker system df は引数不要
		return []prompt.Suggest{}
	
	case "git":
		if len(words) == 2 && !strings.HasSuffix(beforeCursor, " ") {
			return s.completeGitSubcommands(currentArg)
		}
		return s.completeFilesAndDirectories(currentArg)
	case "docker":
		return s.completeDockerCommand(words, currentArg, beforeCursor)
	case "theme":
		return s.completeThemes(currentArg)
	case "lang":
		return s.completeLanguages(currentArg)
	default:
		return s.completeFilesAndDirectories(currentArg)
	}
}

// completeCommands はコマンド名の補完を提供します
func (s *Shell) completeCommands(prefix string) []prompt.Suggest {
	suggests := []prompt.Suggest{
		// 内蔵コマンド
		{Text: "cd", Description: "コンテナにログイン"},
		{Text: "pwd", Description: "現在のディレクトリを表示"},
		{Text: "ls", Description: "ディレクトリの内容を表示"},
		{Text: "cat", Description: "ファイルの内容を表示"},
		{Text: "cp", Description: "ファイルをコピー"},
		{Text: "mv", Description: "ファイルを移動"},
		{Text: "mkdir", Description: "ディレクトリを作成"},
		{Text: "rmdir", Description: "ディレクトリを削除"},
		{Text: "touch", Description: "ファイルを作成"},
		{Text: "echo", Description: "文字列を表示"},
		{Text: "clear", Description: "画面をクリア"},
		{Text: "exit", Description: "シェルを終了"},

		// Docker専用コマンド（このシェルでは Linux コマンドが Docker コマンドにマッピング）
		{Text: "ps", Description: "コンテナ一覧表示 (docker ps)"},
		{Text: "kill", Description: "コンテナ停止 (docker stop)"},
		{Text: "rm", Description: "Dockerコンテナを削除"},
		{Text: "rmi", Description: "Dockerイメージを削除"},
		{Text: "start", Description: "Dockerコンテナを開始"},
		{Text: "stop", Description: "Dockerコンテナを停止"},
		{Text: "exec", Description: "Dockerコンテナ内でコマンド実行"},
		{Text: "pull", Description: "Dockerイメージをプル"},
		
		// ログ系コマンド
		{Text: "tail", Description: "ログ表示 (docker logs)"},
		{Text: "head", Description: "ログ先頭表示 (docker logs --tail)"},
		{Text: "grep", Description: "ログ検索 (docker logs | grep)"},
		
		// ファイル操作系コマンド（コンテナ内実行）
		{Text: "vi", Description: "ファイル編集 (docker exec vi)"},
		{Text: "nano", Description: "ファイル編集 (docker exec nano)"},
		{Text: "mkdir", Description: "ディレクトリ作成 (docker exec mkdir)"},
		{Text: "find", Description: "ファイル検索 (docker exec find)"},
		{Text: "locate", Description: "ファイル検索 (docker exec find)"},
		
		// ネットワーク系コマンド
		{Text: "netstat", Description: "ポート表示 (docker port)"},
		
		// システム情報系コマンド
		{Text: "free", Description: "メモリ使用量 (docker stats)"},
		{Text: "top", Description: "リアルタイム統計 (docker stats)"},
		{Text: "htop", Description: "リアルタイム統計 (docker stats)"},
		{Text: "df", Description: "ディスク使用量 (docker system df)"},
		{Text: "du", Description: "ディスク使用量詳細 (docker system df)"},
		{Text: "uname", Description: "システム情報 (docker version)"},

		// システムコマンド
		{Text: "git", Description: "Git バージョン管理"},
		{Text: "docker", Description: "Docker コンテナ管理"},
		{Text: "theme", Description: "テーマを変更"},
		{Text: "lang", Description: "言語を変更"},
		{Text: "alias", Description: "エイリアスを管理"},
		{Text: "config", Description: "設定を表示"},
	}

	// エイリアスを追加
	if s.config != nil {
		for alias := range s.config.Aliases {
			suggests = append(suggests, prompt.Suggest{
				Text:        alias,
				Description: "エイリアス: " + s.config.Aliases[alias],
			})
		}
	}

	return prompt.FilterHasPrefix(suggests, prefix, true)
}

// completeFiles はファイル名の補完を提供します
func (s *Shell) completeFiles(prefix string) []prompt.Suggest {
	return s.completeFileSystem(prefix, false, true)
}

// completeDirectories はディレクトリ名の補完を提供します
func (s *Shell) completeDirectories(prefix string) []prompt.Suggest {
	return s.completeFileSystem(prefix, true, false)
}

// completeFilesAndDirectories はファイルとディレクトリの補完を提供します
func (s *Shell) completeFilesAndDirectories(prefix string) []prompt.Suggest {
	return s.completeFileSystem(prefix, true, true)
}

// completeFileSystem はファイルシステムの補完を提供します
func (s *Shell) completeFileSystem(prefix string, includeDirs, includeFiles bool) []prompt.Suggest {
	var suggests []prompt.Suggest

	// パスを解析
	dir := filepath.Dir(prefix)
	base := filepath.Base(prefix)

	// 相対パスの場合は現在のディレクトリを基準にする
	if !filepath.IsAbs(dir) {
		if dir == "." || prefix == "" || !strings.Contains(prefix, string(filepath.Separator)) {
			dir = s.getCurrentDir()
			if prefix != "" && !strings.Contains(prefix, string(filepath.Separator)) {
				base = prefix
			} else {
				base = ""
			}
		} else {
			dir = filepath.Join(s.getCurrentDir(), dir)
		}
	}

	// ディレクトリの内容を読み取り
	entries, err := os.ReadDir(dir)
	if err != nil {
		return suggests
	}

	for _, entry := range entries {
		name := entry.Name()

		// 隠しファイルは . で始まる場合のみ表示
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(base, ".") {
			continue
		}

		// フィルタリング
		if !strings.HasPrefix(name, base) {
			continue
		}

		// パスを構築
		var fullPath string
		if strings.Contains(prefix, string(filepath.Separator)) {
			fullPath = filepath.Join(filepath.Dir(prefix), name)
		} else {
			fullPath = name
		}

		// ディレクトリの場合
		if entry.IsDir() {
			if includeDirs {
				suggests = append(suggests, prompt.Suggest{
					Text:        fullPath + string(filepath.Separator),
					Description: "ディレクトリ",
				})
			}
		} else {
			// ファイルの場合
			if includeFiles {
				suggests = append(suggests, prompt.Suggest{
					Text:        fullPath,
					Description: "ファイル",
				})
			}
		}
	}

	return suggests
}

// completeGitSubcommands はGitサブコマンドの補完を提供します
func (s *Shell) completeGitSubcommands(prefix string) []prompt.Suggest {
	suggests := []prompt.Suggest{
		{Text: "status", Description: "作業ディレクトリの状態を表示"},
		{Text: "add", Description: "ファイルをステージングエリアに追加"},
		{Text: "commit", Description: "変更をコミット"},
		{Text: "push", Description: "リモートリポジトリにプッシュ"},
		{Text: "pull", Description: "リモートリポジトリからプル"},
		{Text: "log", Description: "コミット履歴を表示"},
		{Text: "clone", Description: "リポジトリをクローン"},
		{Text: "help", Description: "ヘルプを表示"},
	}

	return prompt.FilterHasPrefix(suggests, prefix, true)
}

// completeThemes はテーマの補完を提供します
func (s *Shell) completeThemes(prefix string) []prompt.Suggest {
	suggests := []prompt.Suggest{
		{Text: "default", Description: "デフォルトテーマ"},
		{Text: "minimal", Description: "ミニマルテーマ"},
		{Text: "robbyrussell", Description: "Robby Russell テーマ"},
		{Text: "agnoster", Description: "Agnoster テーマ"},
		{Text: "pure", Description: "Pure テーマ"},
	}

	return prompt.FilterHasPrefix(suggests, prefix, true)
}

// completeLanguages は言語の補完を提供します
func (s *Shell) completeLanguages(prefix string) []prompt.Suggest {
	suggests := []prompt.Suggest{
		{Text: "en", Description: "English"},
		{Text: "ja", Description: "日本語"},
	}

	return prompt.FilterHasPrefix(suggests, prefix, true)
}

// Docker補完関数群

// completeDockerContainers はDockerコンテナ名の補完を提供します
func (s *Shell) completeDockerContainers(prefix string, running bool) []prompt.Suggest {
	containers := s.getDockerContainers(running)
	
	// Docker から取得できなかった場合は補完無効
	if len(containers) == 0 {
		return []prompt.Suggest{}
	}
	
	var suggests []prompt.Suggest
	for _, container := range containers {
		var description string
		if running {
			description = "実行中のコンテナ"
		} else {
			description = "コンテナ"
		}
		suggests = append(suggests, prompt.Suggest{
			Text:        container,
			Description: description,
		})
	}

	return prompt.FilterHasPrefix(suggests, prefix, true)
}

// completeDockerImages はDockerイメージ名の補完を提供します
func (s *Shell) completeDockerImages(prefix string) []prompt.Suggest {
	images := s.getDockerImages()
	
	// Docker から取得できなかった場合は補完無効
	if len(images) == 0 {
		return []prompt.Suggest{}
	}
	
	var suggests []prompt.Suggest
	for _, image := range images {
		suggests = append(suggests, prompt.Suggest{
			Text:        image,
			Description: "Dockerイメージ",
		})
	}

	return prompt.FilterHasPrefix(suggests, prefix, true)
}

// completeDockerNetworks はDockerネットワーク名の補完を提供します
func (s *Shell) completeDockerNetworks(prefix string) []prompt.Suggest {
	networks := s.getDockerNetworks()
	
	// Docker から取得できなかった場合は補完無効
	if len(networks) == 0 {
		return []prompt.Suggest{}
	}
	
	var suggests []prompt.Suggest
	for _, network := range networks {
		suggests = append(suggests, prompt.Suggest{
			Text:        network,
			Description: "Dockerネットワーク",
		})
	}

	return prompt.FilterHasPrefix(suggests, prefix, true)
}

// completeDockerVolumes はDockerボリューム名の補完を提供します
func (s *Shell) completeDockerVolumes(prefix string) []prompt.Suggest {
	volumes := s.getDockerVolumes()
	
	// Docker から取得できなかった場合は補完無効
	if len(volumes) == 0 {
		return []prompt.Suggest{}
	}
	
	var suggests []prompt.Suggest
	for _, volume := range volumes {
		suggests = append(suggests, prompt.Suggest{
			Text:        volume,
			Description: "Dockerボリューム",
		})
	}

	return prompt.FilterHasPrefix(suggests, prefix, true)
}

// completeDockerSubcommands はDockerサブコマンドの補完を提供します
func (s *Shell) completeDockerSubcommands(prefix string) []prompt.Suggest {
	suggests := []prompt.Suggest{
		{Text: "ps", Description: "実行中のコンテナを表示"},
		{Text: "images", Description: "イメージ一覧を表示"},
		{Text: "run", Description: "新しいコンテナを実行"},
		{Text: "exec", Description: "実行中のコンテナでコマンドを実行"},
		{Text: "start", Description: "停止中のコンテナを開始"},
		{Text: "stop", Description: "実行中のコンテナを停止"},
		{Text: "restart", Description: "コンテナを再起動"},
		{Text: "rm", Description: "コンテナを削除"},
		{Text: "rmi", Description: "イメージを削除"},
		{Text: "pull", Description: "イメージをダウンロード"},
		{Text: "push", Description: "イメージをアップロード"},
		{Text: "build", Description: "Dockerイメージをビルド"},
		{Text: "logs", Description: "コンテナのログを表示"},
		{Text: "inspect", Description: "詳細情報を表示"},
		{Text: "network", Description: "ネットワークを管理"},
		{Text: "volume", Description: "ボリュームを管理"},
		{Text: "system", Description: "システム情報を表示"},
		{Text: "version", Description: "バージョンを表示"},
	}

	return prompt.FilterHasPrefix(suggests, prefix, true)
}

// completeDockerCommand はDockerコマンド全体の補完ロジックを制御します
func (s *Shell) completeDockerCommand(words []string, currentArg, beforeCursor string) []prompt.Suggest {
	// サブコマンドが入力されていない場合はサブコマンド補完
	if len(words) == 2 && !strings.HasSuffix(beforeCursor, " ") {
		return s.completeDockerSubcommands(currentArg)
	}

	if len(words) < 2 {
		return []prompt.Suggest{}
	}

	subcommand := words[1]

	switch subcommand {
	// コンテナを削除するコマンド
	case "rm":
		if len(words) >= 3 && (len(words) > 3 || strings.HasSuffix(beforeCursor, " ")) {
			return s.completeDockerContainers(currentArg, false) // 全てのコンテナ（停止中も含む）
		}
		return s.completeDockerContainers(currentArg, false)

	// イメージを削除するコマンド
	case "rmi":
		return s.completeDockerImages(currentArg)

	// 実行中のコンテナに対するコマンド
	case "stop", "restart", "exec", "logs", "inspect":
		return s.completeDockerContainers(currentArg, true) // 実行中のコンテナのみ

	// 停止中のコンテナを開始するコマンド
	case "start":
		// 停止中のコンテナを取得するために、全コンテナから実行中を除外
		// ここでは簡単のために全コンテナを表示
		return s.completeDockerContainers(currentArg, false)

	// イメージに関するコマンド
	case "run", "push":
		return s.completeDockerImages(currentArg)

	// イメージをプルする際の候補は取得困難なので補完無効
	case "pull":
		return []prompt.Suggest{}

	// ネットワーク管理
	case "network":
		if len(words) == 3 && !strings.HasSuffix(beforeCursor, " ") {
			// network サブコマンドの補完
			suggests := []prompt.Suggest{
				{Text: "ls", Description: "ネットワーク一覧を表示"},
				{Text: "create", Description: "ネットワークを作成"},
				{Text: "rm", Description: "ネットワークを削除"},
				{Text: "inspect", Description: "ネットワーク詳細を表示"},
			}
			return prompt.FilterHasPrefix(suggests, currentArg, true)
		}
		if len(words) >= 4 && (words[2] == "rm" || words[2] == "inspect") {
			return s.completeDockerNetworks(currentArg)
		}
		return []prompt.Suggest{}

	// ボリューム管理
	case "volume":
		if len(words) == 3 && !strings.HasSuffix(beforeCursor, " ") {
			// volume サブコマンドの補完
			suggests := []prompt.Suggest{
				{Text: "ls", Description: "ボリューム一覧を表示"},
				{Text: "create", Description: "ボリュームを作成"},
				{Text: "rm", Description: "ボリュームを削除"},
				{Text: "inspect", Description: "ボリューム詳細を表示"},
			}
			return prompt.FilterHasPrefix(suggests, currentArg, true)
		}
		if len(words) >= 4 && (words[2] == "rm" || words[2] == "inspect") {
			return s.completeDockerVolumes(currentArg)
		}
		return []prompt.Suggest{}

	// buildコマンドはDockerfileのあるディレクトリを指定するのでディレクトリ補完
	case "build":
		return s.completeDirectories(currentArg)

	// その他のコマンドはファイルシステム補完にフォールバック
	default:
		return s.completeFilesAndDirectories(currentArg)
	}
}
