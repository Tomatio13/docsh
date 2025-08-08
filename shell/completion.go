package shell

import (
	"os"
	"path/filepath"
	"strings"

	"docknaut/i18n"

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
		// 一般的なディレクトリ変更
		return s.completeDirectories(currentArg)
	case "login":
		// login はコンテナへのログイン
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
		{Text: "cd", Description: i18n.T("completion.descriptions.cd")},
		{Text: "login", Description: i18n.T("completion.descriptions.login")},
		{Text: "pwd", Description: i18n.T("completion.descriptions.pwd")},
		{Text: "ls", Description: i18n.T("completion.descriptions.ls")},
		{Text: "cat", Description: i18n.T("completion.descriptions.cat")},
		{Text: "cp", Description: i18n.T("completion.descriptions.cp")},
		{Text: "mv", Description: i18n.T("completion.descriptions.mv")},
		{Text: "mkdir", Description: i18n.T("completion.descriptions.mkdir")},
		{Text: "rmdir", Description: i18n.T("completion.descriptions.rmdir")},
		{Text: "touch", Description: i18n.T("completion.descriptions.touch")},
		{Text: "echo", Description: i18n.T("completion.descriptions.echo")},
		{Text: "clear", Description: i18n.T("completion.descriptions.clear")},
		{Text: "exit", Description: i18n.T("completion.descriptions.exit")},

		// Docker専用コマンド（このシェルでは Linux コマンドが Docker コマンドにマッピング）
		{Text: "ps", Description: i18n.T("completion.descriptions.ps")},
		{Text: "kill", Description: i18n.T("completion.descriptions.kill")},
		{Text: "rm", Description: i18n.T("completion.descriptions.rm")},
		{Text: "rmi", Description: i18n.T("completion.descriptions.rmi")},
		{Text: "start", Description: i18n.T("completion.descriptions.start")},
		{Text: "stop", Description: i18n.T("completion.descriptions.stop")},
		{Text: "exec", Description: i18n.T("completion.descriptions.exec")},
		{Text: "pull", Description: i18n.T("completion.descriptions.pull")},

		// ログ系コマンド
		{Text: "tail", Description: i18n.T("completion.descriptions.tail")},
		{Text: "head", Description: i18n.T("completion.descriptions.head")},
		{Text: "grep", Description: i18n.T("completion.descriptions.grep")},

		// ファイル操作系コマンド（コンテナ内実行）
		{Text: "vi", Description: i18n.T("completion.descriptions.vi")},
		{Text: "nano", Description: i18n.T("completion.descriptions.nano")},
		{Text: "mkdir", Description: i18n.T("completion.descriptions.mkdir")},
		{Text: "find", Description: i18n.T("completion.descriptions.find")},
		{Text: "locate", Description: i18n.T("completion.descriptions.locate")},

		// ネットワーク系コマンド
		{Text: "netstat", Description: i18n.T("completion.descriptions.netstat")},

		// システム情報系コマンド
		{Text: "free", Description: i18n.T("completion.descriptions.free")},
		{Text: "top", Description: i18n.T("completion.descriptions.top")},
		{Text: "htop", Description: i18n.T("completion.descriptions.htop")},
		{Text: "df", Description: i18n.T("completion.descriptions.df")},
		{Text: "du", Description: i18n.T("completion.descriptions.du")},
		{Text: "uname", Description: i18n.T("completion.descriptions.uname")},

		// システムコマンド

		{Text: "docker", Description: i18n.T("completion.descriptions.docker")},
		{Text: "theme", Description: i18n.T("completion.descriptions.theme")},
		{Text: "lang", Description: i18n.T("completion.descriptions.lang")},
		{Text: "alias", Description: i18n.T("completion.descriptions.alias")},
		{Text: "config", Description: i18n.T("completion.descriptions.config")},
	}

	// エイリアスを追加
	if s.config != nil {
		for alias := range s.config.Aliases {
			suggests = append(suggests, prompt.Suggest{
				Text:        alias,
				Description: i18n.T("completion.alias_value", s.config.Aliases[alias]),
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
					Description: i18n.T("completion.entry_directory"),
				})
			}
		} else {
			// ファイルの場合
			if includeFiles {
				suggests = append(suggests, prompt.Suggest{
					Text:        fullPath,
					Description: i18n.T("completion.entry_file"),
				})
			}
		}
	}

	return suggests
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
		{Text: "ps", Description: i18n.T("completion.docker_subcommands.ps")},
		{Text: "images", Description: i18n.T("completion.docker_subcommands.images")},
		{Text: "run", Description: i18n.T("completion.docker_subcommands.run")},
		{Text: "exec", Description: i18n.T("completion.docker_subcommands.exec")},
		{Text: "start", Description: i18n.T("completion.docker_subcommands.start")},
		{Text: "stop", Description: i18n.T("completion.docker_subcommands.stop")},
		{Text: "restart", Description: i18n.T("completion.docker_subcommands.restart")},
		{Text: "rm", Description: i18n.T("completion.docker_subcommands.rm")},
		{Text: "rmi", Description: i18n.T("completion.docker_subcommands.rmi")},
		{Text: "pull", Description: i18n.T("completion.docker_subcommands.pull")},
		{Text: "push", Description: i18n.T("completion.docker_subcommands.push")},
		{Text: "build", Description: i18n.T("completion.docker_subcommands.build")},
		{Text: "logs", Description: i18n.T("completion.docker_subcommands.logs")},
		{Text: "inspect", Description: i18n.T("completion.docker_subcommands.inspect")},
		{Text: "network", Description: i18n.T("completion.docker_subcommands.network")},
		{Text: "volume", Description: i18n.T("completion.docker_subcommands.volume")},
		{Text: "system", Description: i18n.T("completion.docker_subcommands.system")},
		{Text: "version", Description: i18n.T("completion.docker_subcommands.version")},
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
				{Text: "ls", Description: i18n.T("completion.docker_network_subcommands.ls")},
				{Text: "create", Description: i18n.T("completion.docker_network_subcommands.create")},
				{Text: "rm", Description: i18n.T("completion.docker_network_subcommands.rm")},
				{Text: "inspect", Description: i18n.T("completion.docker_network_subcommands.inspect")},
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
				{Text: "ls", Description: i18n.T("completion.docker_volume_subcommands.ls")},
				{Text: "create", Description: i18n.T("completion.docker_volume_subcommands.create")},
				{Text: "rm", Description: i18n.T("completion.docker_volume_subcommands.rm")},
				{Text: "inspect", Description: i18n.T("completion.docker_volume_subcommands.inspect")},
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
