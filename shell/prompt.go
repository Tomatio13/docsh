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
			prompt = theme.Prompt
		}
	}

	// プロンプトが設定されていない場合はデフォルト
	if prompt == "" {
		prompt = "🌸 %s $ "
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

	// 改行文字を削除
	result = strings.ReplaceAll(result, "\n", "")
	result = strings.ReplaceAll(result, "\r", "")

	// 末尾の空白を削除
	result = strings.TrimRight(result, " ")

	// プロンプトの末尾にスペースを追加（入力との区切り）
	if !strings.HasSuffix(result, " ") {
		result += " "
	}

	return result
}

func (s *Shell) expandPromptVariables(prompt string) string {
	result := prompt

	// 基本的な変数展開
	variables := map[string]string{
		"%s":   s.getShortPath(),
		"%d":   s.getCurrentDir(),
		"%u":   s.getUsername(),
		"%h":   s.getHostname(),
		"%t":   s.getTime(),
		"%D":   s.getDate(),
		"%w":   s.getWeekday(),
		"%os":  s.getOSInfo(),
		"%git": s.getGitBranch(),
		"%gs":  s.getGitStatus(),
	}

	for placeholder, value := range variables {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// テーマの色設定を適用
	if s.config != nil && s.config.Theme != "" {
		result = themes.ApplyThemeColors(s.config.Theme, result)
	}

	return result
}

func (s *Shell) getShortPath() string {
	currentDir := s.getCurrentDir()
	homeDir, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(currentDir, homeDir) {
		return "~" + currentDir[len(homeDir):]
	}

	// パスが長い場合は短縮
	if len(currentDir) > 50 {
		parts := strings.Split(currentDir, string(filepath.Separator))
		if len(parts) > 3 {
			return "..." + string(filepath.Separator) + strings.Join(parts[len(parts)-2:], string(filepath.Separator))
		}
	}

	return currentDir
}

func (s *Shell) getUsername() string {
	if username := os.Getenv("USER"); username != "" {
		return username
	}
	if username := os.Getenv("USERNAME"); username != "" {
		return username
	}
	return "user"
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
	return time.Now().Format("Monday")
}

func (s *Shell) getOSInfo() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

func (s *Shell) getGitBranch() string {
	// 簡単なgitブランチ検出（実装は簡略化）
	return ""
}

func (s *Shell) getGitStatus() string {
	// 簡単なgitステータス検出（実装は簡略化）
	return ""
}
