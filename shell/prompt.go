package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"cherrysh/themes"
)

func (s *Shell) showPrompt() {
	prompt := s.buildPrompt()
	fmt.Print(prompt)
}

func (s *Shell) buildPrompt() string {
	var prompt string

	// テーマが設定されている場合はテーマのプロンプトを使用
	if s.config != nil && s.config.Theme != "" {
		if theme, exists := themes.GetTheme(s.config.Theme); exists {
			prompt = theme.GetPrompt()
		}
	}

	// カスタムプロンプトが設定されている場合はそれを優先
	if s.config != nil && s.config.Prompt != "" && s.config.Prompt != "cherry:%s$ " {
		prompt = s.config.Prompt
	}

	// プロンプトが設定されていない場合はデフォルト
	if prompt == "" {
		prompt = "cherry:%s$ "
	}

	// 変数展開とクリーンアップ
	result := s.expandPromptVariables(prompt)
	return s.cleanPrompt(result)
}

func (s *Shell) cleanPrompt(prompt string) string {
	result := prompt

	// 連続する空白を単一の空白に変換
	spaceRegex := regexp.MustCompile(`\s{2,}`)
	result = spaceRegex.ReplaceAllString(result, " ")

	// 先頭の空白を除去（ただし末尾のスペースは意図的な場合があるので保持）
	result = strings.TrimLeft(result, " \t")

	// ANSIエスケープシーケンスの直後の余分な空白を除去
	// \033[XXXm の後に続く空白を除去（ただし最後のスペースは保持）
	ansiSpaceRegex := regexp.MustCompile(`(\033\[[0-9;]*m)\s+`)
	result = ansiSpaceRegex.ReplaceAllString(result, "$1")

	// エスケープシーケンス間の余分な空白を除去
	multiAnsiRegex := regexp.MustCompile(`(\033\[[0-9;]*m)\s+(\033\[[0-9;]*m)`)
	result = multiAnsiRegex.ReplaceAllString(result, "$1$2")

	return result
}

func (s *Shell) expandPromptVariables(prompt string) string {
	// まず \n と \t をプレースホルダーとして扱い、
	// 変数展開前に処理することで、パス文字列 (例: "C:\\temp") 内の
	// 文字列が誤って置換される問題を防ぐ。
	result := strings.ReplaceAll(prompt, "\\n", "\n")
	result = strings.ReplaceAll(result, "\\t", "\t")

	// 変数プレースホルダーを安定した順序で置換する
	replacements := map[string]string{
		"%s": s.getShortPath(),  // 現在のディレクトリ (短縮)
		"%S": s.getCurrentDir(), // 現在のディレクトリ (フルパス)
		"%u": s.getUsername(),   // ユーザー名
		"%h": s.getHostname(),   // ホスト名
		"%t": s.getTime(),       // 現在時刻
		"%d": s.getDate(),       // 現在日付
		"%w": s.getWeekday(),    // 曜日
		"%%": "%",               // エスケープされた%
	}

	// 置換順序を固定することで map 順序の非決定性を排除
	order := []string{"%s", "%S", "%u", "%h", "%t", "%d", "%w", "%%"}
	for _, placeholder := range order {
		result = strings.ReplaceAll(result, placeholder, replacements[placeholder])
	}

	return result
}

func (s *Shell) getShortPath() string {
	current := s.getCurrentDir()

	// ホームディレクトリの場合は~で表示
	if home, err := os.UserHomeDir(); err == nil {
		if strings.HasPrefix(current, home) {
			return "~" + current[len(home):]
		}
	}

	// 長いパスの場合は末尾のディレクトリ名のみ表示
	if len(current) > 30 {
		return "..." + filepath.Base(current)
	}

	return current
}

func (s *Shell) getUsername() string {
	if username := os.Getenv("USER"); username != "" {
		return username
	}
	if username := os.Getenv("USERNAME"); username != "" { // Windows
		return username
	}
	return "unknown"
}

func (s *Shell) getHostname() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return "localhost"
}

func (s *Shell) getTime() string {
	return time.Now().Format("15:04:05")
}

func (s *Shell) getDate() string {
	return time.Now().Format("2006-01-02")
}

func (s *Shell) getWeekday() string {
	return time.Now().Format("Mon")
}

func (s *Shell) getOSInfo() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

// Git情報を取得（将来の拡張用）
func (s *Shell) getGitBranch() string {
	// TODO: gitブランチ情報の取得実装
	return ""
}

func (s *Shell) getGitStatus() string {
	// TODO: git状態情報の取得実装
	return ""
}
