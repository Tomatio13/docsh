package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cherrysh/i18n"
)

type Config struct {
	Aliases     map[string]string
	Theme       string
	Language    string
	GitHubToken string
	GitHubUser  string
}

func NewConfig() *Config {
	return &Config{
		Aliases:  make(map[string]string),
		Theme:    "default",
		Language: "", // 空の場合は自動検出
	}
}

func (c *Config) LoadConfigFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".cherryshrc")
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 設定ファイルが存在しない場合は無視
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 空行とコメント行をスキップ
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 設定行の処理を統一
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "LANG":
					c.Language = strings.Trim(value, "\"'")
				case "THEME":
					c.Theme = strings.Trim(value, "\"'")
				case "GITHUB_TOKEN":
					c.GitHubToken = strings.Trim(value, "\"'")
				case "GITHUB_USER":
					c.GitHubUser = strings.Trim(value, "\"'")
				}
			}
		}

		// エイリアス行の処理
		if strings.HasPrefix(line, "alias ") {
			aliasLine := strings.TrimPrefix(line, "alias ")
			if err := c.ParseAlias(aliasLine); err != nil {
				fmt.Printf(i18n.T("config.parse_error")+"\n", lineNum, err)
			}
		}

		// 旧形式のテーマ設定もサポート
		if strings.HasPrefix(line, "theme ") {
			themeLine := strings.TrimPrefix(line, "theme ")
			c.Theme = strings.TrimSpace(themeLine)
		}
	}

	return scanner.Err()
}

func (c *Config) SaveConfigFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".cherryshrc")
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// ヘッダーコメントを追加
	fmt.Fprintln(file, "# Cherry Shell Configuration File")
	fmt.Fprintln(file, "# 🌸 Cherry Shell - Beautiful & Simple Shell 🌸")
	fmt.Fprintln(file, "")

	// 言語設定を保存
	if c.Language != "" {
		fmt.Fprintf(file, "LANG=%s\n", c.Language)
		fmt.Fprintln(file, "")
	}

	// テーマ設定を保存
	if c.Theme != "default" {
		fmt.Fprintf(file, "theme %s\n", c.Theme)
		fmt.Fprintln(file, "")
	}

	// エイリアス設定を保存
	if len(c.Aliases) > 0 {
		fmt.Fprintln(file, "# エイリアス設定")
		for name, command := range c.Aliases {
			fmt.Fprintf(file, "alias %s=\"%s\"\n", name, command)
		}
	}

	return nil
}

// GetLanguage は設定ファイルから言語設定を取得し、未設定の場合は自動検出を行う
func (c *Config) GetLanguage(args []string) string {
	// 設定ファイルに言語設定がある場合はそれを使用
	if c.Language != "" {
		return c.Language
	}

	// 設定ファイルに言語設定がない場合は従来の自動検出を使用
	return i18n.DetectLanguage(args)
}

// SetLanguage は言語設定を変更し、設定ファイルに保存する
func (c *Config) SetLanguage(language string) error {
	c.Language = language
	return c.SaveConfigFile()
}
