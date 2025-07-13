package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Aliases    map[string]string
	Prompt     string
	Theme      string
	Variables  map[string]string
}

func NewConfig() *Config {
	return &Config{
		Aliases:   make(map[string]string),
		Variables: make(map[string]string),
		Prompt:    "cherry:%s$ ",
		Theme:     "robbyrussell",
	}
}

func (c *Config) LoadConfigFile() error {
	configPaths := []string{
		".cherryshrc",
		filepath.Join(os.Getenv("HOME"), ".cherryshrc"),
		filepath.Join(os.Getenv("USERPROFILE"), ".cherryshrc"), // Windows
	}
	
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return c.parseConfigFile(path)
		}
	}
	
	// 設定ファイルが見つからない場合はデフォルトを作成
	return c.createDefaultConfig()
}

func (c *Config) parseConfigFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// 空行やコメント行をスキップ
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		if err := c.parseLine(line); err != nil {
			fmt.Printf("Warning: Error parsing line %d: %v\n", lineNum, err)
		}
	}
	
	return scanner.Err()
}

func (c *Config) parseLine(line string) error {
	// aliasコマンドの解析
	if strings.HasPrefix(line, "alias ") {
		return c.parseAlias(line[6:]) // "alias "を除去
	}
	
	// 環境変数の設定
	if strings.Contains(line, "=") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			
			switch key {
			case "PROMPT":
				c.Prompt = value
			case "THEME":
				c.Theme = value
			default:
				c.Variables[key] = value
				os.Setenv(key, value) // 環境変数として設定
			}
		}
	}
	
	return nil
}

func (c *Config) parseAlias(aliasDef string) error {
	// alias name='command' または alias name=command の形式を解析
	parts := strings.SplitN(aliasDef, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid alias format: %s", aliasDef)
	}
	
	name := strings.TrimSpace(parts[0])
	command := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
	
	c.Aliases[name] = command
	return nil
}

func (c *Config) createDefaultConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	
	configPath := filepath.Join(homeDir, ".cherryshrc")
	
	defaultConfig := `# Cherry Shell Configuration File
# 🌸 Cherry Shell - Beautiful & Simple Shell 🌸
# 
# Prompt configuration
PROMPT="cherry:%s$ "

# Theme setting
THEME="robbyrussell"

# Aliases
alias ll='ls -la'
alias la='ls -la'
alias l='ls -l'
alias grep='grep --color=auto'
alias ..='cd ..'
alias ...='cd ../..'

# Custom environment variables
# EDITOR="vim"
# BROWSER="firefox"
`
	
	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}

func (c *Config) GetAlias(name string) (string, bool) {
	command, exists := c.Aliases[name]
	return command, exists
}

func (c *Config) ParseAlias(aliasDef string) error {
	return c.parseAlias(aliasDef)
}