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
		return s.completeDirectories(currentArg)
	case "cat", "rm", "cp", "mv":
		return s.completeFiles(currentArg)
	case "ls":
		return s.completeFilesAndDirectories(currentArg)
	case "git":
		if len(words) == 2 && !strings.HasSuffix(beforeCursor, " ") {
			return s.completeGitSubcommands(currentArg)
		}
		return s.completeFilesAndDirectories(currentArg)
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
		{Text: "cd", Description: "ディレクトリを変更"},
		{Text: "pwd", Description: "現在のディレクトリを表示"},
		{Text: "ls", Description: "ディレクトリの内容を表示"},
		{Text: "cat", Description: "ファイルの内容を表示"},
		{Text: "cp", Description: "ファイルをコピー"},
		{Text: "mv", Description: "ファイルを移動"},
		{Text: "rm", Description: "ファイルを削除"},
		{Text: "mkdir", Description: "ディレクトリを作成"},
		{Text: "rmdir", Description: "ディレクトリを削除"},
		{Text: "touch", Description: "ファイルを作成"},
		{Text: "echo", Description: "文字列を表示"},
		{Text: "clear", Description: "画面をクリア"},
		{Text: "exit", Description: "シェルを終了"},

		// システムコマンド
		{Text: "git", Description: "Git バージョン管理"},
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
