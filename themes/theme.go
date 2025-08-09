package themes

import (
	"fmt"
	"strings"

	"docsh/i18n"
)

// Theme はプロンプトテーマの定義
type Theme struct {
	Name   string
	Prompt string
	Colors map[string]string
}

// 利用可能なテーマ
var themes = map[string]Theme{
	"default": {
		Name:   "Default",
		Prompt: "🐳 %s $ ",
		Colors: map[string]string{
			"directory": "cyan",
			"prompt":    "green",
			"error":     "red",
		},
	},
	"minimal": {
		Name:   "Minimal",
		Prompt: "%s > ",
		Colors: map[string]string{
			"directory": "blue",
			"prompt":    "white",
			"error":     "red",
		},
	},
	"robbyrussell": {
		Name:   "Robbyrussell",
		Prompt: "➜ %s ",
		Colors: map[string]string{
			"directory": "cyan",
			"prompt":    "green",
			"error":     "red",
		},
	},
	"agnoster": {
		Name:   "Agnoster",
		Prompt: "⚡ %s ➤ ",
		Colors: map[string]string{
			"directory": "blue",
			"prompt":    "yellow",
			"error":     "red",
		},
	},
	"pure": {
		Name:   "Pure",
		Prompt: "%s ❯ ",
		Colors: map[string]string{
			"directory": "blue",
			"prompt":    "magenta",
			"error":     "red",
		},
	},
}

// GetTheme は指定されたテーマを取得
func GetTheme(name string) (Theme, bool) {
	theme, exists := themes[name]
	return theme, exists
}

// GetThemePrompt はテーマのプロンプトを取得
func GetThemePrompt(themeName string) string {
	if theme, exists := themes[themeName]; exists {
		return theme.Prompt
	}
	return themes["default"].Prompt
}

// GetThemeColor はテーマの色設定を取得
func GetThemeColor(themeName, colorType string) string {
	if theme, exists := themes[themeName]; exists {
		if color, exists := theme.Colors[colorType]; exists {
			return color
		}
	}
	return themes["default"].Colors[colorType]
}

// colorizeText は文字列に色を適用
func colorizeText(text, color string) string {
	colorCodes := map[string]string{
		"black":   "30",
		"red":     "31",
		"green":   "32",
		"yellow":  "33",
		"blue":    "34",
		"magenta": "35",
		"cyan":    "36",
		"white":   "37",
	}

	if code, exists := colorCodes[color]; exists {
		return fmt.Sprintf("\033[%sm%s\033[0m", code, text)
	}
	return text
}

// ApplyThemeColors はテーマの色設定を適用
func ApplyThemeColors(themeName, text string) string {
	theme, exists := themes[themeName]
	if !exists {
		return text
	}

	result := text
	for colorName, colorValue := range theme.Colors {
		placeholder := fmt.Sprintf("$fg[%s]", colorName)
		result = strings.ReplaceAll(result, placeholder, colorizeText("", colorValue))

		placeholder = fmt.Sprintf("%%{$fg[%s]%%}", colorName)
		result = strings.ReplaceAll(result, placeholder, colorizeText("", colorValue))
	}

	return result
}

// GetAvailableThemes は利用可能なテーマのリストを取得
func GetAvailableThemes() []string {
	var themeNames []string
	for name := range themes {
		themeNames = append(themeNames, name)
	}
	return themeNames
}

// AddTheme は新しいテーマを追加
func AddTheme(name string, theme Theme) {
	themes[name] = theme
}

// RemoveTheme はテーマを削除
func RemoveTheme(name string) bool {
	if name == "default" {
		return false // デフォルトテーマは削除不可
	}

	if _, exists := themes[name]; exists {
		delete(themes, name)
		return true
	}
	return false
}

// ValidateTheme はテーマの妥当性をチェック
func ValidateTheme(theme Theme) error {
	if theme.Name == "" {
		return fmt.Errorf("theme name cannot be empty")
	}

	if theme.Prompt == "" {
		return fmt.Errorf("theme prompt cannot be empty")
	}

	// 必要な色設定がすべて存在するかチェック
	requiredColors := []string{"directory", "prompt", "error"}
	for _, colorType := range requiredColors {
		if _, exists := theme.Colors[colorType]; !exists {
			return fmt.Errorf("missing required color: %s", colorType)
		}
	}

	return nil
}

// ListThemes は利用可能なテーマを表示
func ListThemes() {
	fmt.Println(i18n.T("theme.available_themes"))
	for name, theme := range themes {
		fmt.Printf("  %s - %s\n", name, theme.Name)
	}
}
