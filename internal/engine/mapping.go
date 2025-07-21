package engine

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// CommandMapping represents a Linux to Docker command mapping
type CommandMapping struct {
	ID            string            `json:"id" yaml:"id"`
	LinuxCommand  string            `json:"linux_command" yaml:"linux_command"`
	DockerCommand string            `json:"docker_command" yaml:"docker_command"`
	Category      string            `json:"category" yaml:"category"`
	Description   string            `json:"description" yaml:"description"`
	LinuxExample  string            `json:"linux_example" yaml:"linux_example"`
	DockerExample string            `json:"docker_example" yaml:"docker_example"`
	Notes         []string          `json:"notes" yaml:"notes"`
	Warnings      []string          `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	LocalizedDescription map[string]string `json:"localized_description,omitempty" yaml:"localized_description,omitempty"`
	LocalizedNotes       map[string][]string `json:"localized_notes,omitempty" yaml:"localized_notes,omitempty"`
}

// MappingEngine defines the interface for command mapping operations
type MappingEngine interface {
	LoadMappings() error
	FindByLinuxCommand(cmd string) (*CommandMapping, error)
	FindByLinuxCommandWithOptions(baseCmd string, options map[string]string) (*CommandMapping, error)
	FindByDockerCommand(cmd string) (*CommandMapping, error)
	ListByCategory(category string) ([]*CommandMapping, error)
	SearchCommands(query string) ([]*CommandMapping, error)
	GetAllMappings() []*CommandMapping
	GetCategories() []string
}

// DefaultMappingEngine is the default implementation of MappingEngine
type DefaultMappingEngine struct {
	mappings []CommandMapping
	dataPath string
}

// NewMappingEngine creates a new mapping engine instance
func NewMappingEngine(dataPath string) MappingEngine {
	return &DefaultMappingEngine{
		mappings: []CommandMapping{},
		dataPath: dataPath,
	}
}

// LoadMappings loads command mappings from the data file
func (engine *DefaultMappingEngine) LoadMappings() error {
	var mappingsData struct {
		Mappings []CommandMapping `yaml:"mappings"`
	}

	// Try to find the data file
	dataFile := filepath.Join(engine.dataPath, "mappings.yaml")
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		// Use default mappings if file doesn't exist
		engine.mappings = getDefaultMappings()
		return nil
	}

	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return fmt.Errorf("failed to read mappings file: %v", err)
	}

	err = yaml.Unmarshal(data, &mappingsData)
	if err != nil {
		return fmt.Errorf("failed to parse mappings file: %v", err)
	}

	engine.mappings = mappingsData.Mappings
	return nil
}

// FindByLinuxCommand finds a mapping by Linux command
func (engine *DefaultMappingEngine) FindByLinuxCommand(cmd string) (*CommandMapping, error) {
	// 完全一致を最初に試行
	for _, mapping := range engine.mappings {
		if mapping.LinuxCommand == cmd {
			return &mapping, nil
		}
	}
	return nil, fmt.Errorf("no mapping found for Linux command: %s", cmd)
}

// FindByLinuxCommandWithOptions finds a mapping by Linux command with options
func (engine *DefaultMappingEngine) FindByLinuxCommandWithOptions(baseCmd string, options map[string]string) (*CommandMapping, error) {
	// オプションを含む検索を試行
	for _, mapping := range engine.mappings {
		// 完全一致を最初に試行
		if mapping.LinuxCommand == baseCmd {
			return &mapping, nil
		}
		
		// オプション付きのコマンドを試行
		if strings.HasPrefix(mapping.LinuxCommand, baseCmd+" ") {
			cmdParts := strings.Fields(mapping.LinuxCommand)
			if len(cmdParts) >= 2 {
				// 最初の部分がベースコマンドと一致し、オプションが含まれている場合
				if cmdParts[0] == baseCmd {
					// オプションをチェック
					for _, part := range cmdParts[1:] {
						if strings.HasPrefix(part, "-") {
							// オプションの部分を抽出
							optionKey := strings.TrimPrefix(part, "-")
							optionKey = strings.TrimPrefix(optionKey, "-") // --option の場合
							if _, exists := options[optionKey]; exists {
								return &mapping, nil
							}
						}
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("no mapping found for Linux command: %s with options", baseCmd)
}

// FindByDockerCommand finds a mapping by Docker command
func (engine *DefaultMappingEngine) FindByDockerCommand(cmd string) (*CommandMapping, error) {
	for _, mapping := range engine.mappings {
		if strings.HasPrefix(mapping.DockerCommand, cmd) {
			return &mapping, nil
		}
	}
	return nil, fmt.Errorf("no mapping found for Docker command: %s", cmd)
}

// ListByCategory returns all mappings in a specific category
func (engine *DefaultMappingEngine) ListByCategory(category string) ([]*CommandMapping, error) {
	var results []*CommandMapping
	for _, mapping := range engine.mappings {
		if mapping.Category == category {
			results = append(results, &mapping)
		}
	}
	return results, nil
}

// SearchCommands searches for mappings by query string
func (engine *DefaultMappingEngine) SearchCommands(query string) ([]*CommandMapping, error) {
	var results []*CommandMapping
	query = strings.ToLower(query)

	for _, mapping := range engine.mappings {
		if strings.Contains(strings.ToLower(mapping.LinuxCommand), query) ||
			strings.Contains(strings.ToLower(mapping.DockerCommand), query) ||
			strings.Contains(strings.ToLower(mapping.Description), query) {
			results = append(results, &mapping)
		}
	}
	return results, nil
}

// GetAllMappings returns all loaded mappings
func (engine *DefaultMappingEngine) GetAllMappings() []*CommandMapping {
	var results []*CommandMapping
	for i := range engine.mappings {
		results = append(results, &engine.mappings[i])
	}
	return results
}

// GetCategories returns all unique categories
func (engine *DefaultMappingEngine) GetCategories() []string {
	categories := make(map[string]bool)
	for _, mapping := range engine.mappings {
		categories[mapping.Category] = true
	}

	var result []string
	for category := range categories {
		result = append(result, category)
	}
	return result
}

// getDefaultMappings returns default command mappings
func getDefaultMappings() []CommandMapping {
	return []CommandMapping{
		{
			ID:            "ls-docker-images",
			LinuxCommand:  "ls",
			DockerCommand: "docker images",
			Category:      "list-operations",
			Description:   "リスト表示 - 利用可能なイメージを表示",
			LinuxExample:  "ls -la",
			DockerExample: "docker images -a",
			Notes:         []string{"docker imagesはDockerイメージを表示", "-aオプションで中間イメージも表示"},
			LocalizedDescription: map[string]string{
				"en": "List display - Show available images",
				"ja": "リスト表示 - 利用可能なイメージを表示",
			},
			LocalizedNotes: map[string][]string{
				"en": {"docker images shows Docker images", "-a option shows intermediate images too"},
				"ja": {"docker imagesはDockerイメージを表示", "-aオプションで中間イメージも表示"},
			},
		},
		{
			ID:            "ps-docker-ps",
			LinuxCommand:  "ps",
			DockerCommand: "docker ps",
			Category:      "process-management",
			Description:   "プロセス一覧表示",
			LinuxExample:  "ps aux",
			DockerExample: "docker ps -a",
			Notes:         []string{"psは全プロセス、docker psはコンテナのみ"},
			LocalizedDescription: map[string]string{
				"en": "Process list display",
				"ja": "プロセス一覧表示",
			},
			LocalizedNotes: map[string][]string{
				"en": {"ps shows all processes, docker ps shows containers only"},
				"ja": {"psは全プロセス、docker psはコンテナのみ"},
			},
		},
		{
			ID:            "kill-docker-stop",
			LinuxCommand:  "kill",
			DockerCommand: "docker stop",
			Category:      "process-management",
			Description:   "プロセス停止",
			LinuxExample:  "kill 1234",
			DockerExample: "docker stop container_name",
			Notes:         []string{"killはPID、docker stopはコンテナ名またはID"},
			LocalizedDescription: map[string]string{
				"en": "Stop process",
				"ja": "プロセス停止",
			},
			LocalizedNotes: map[string][]string{
				"en": {"kill uses PID, docker stop uses container name or ID"},
				"ja": {"killはPID、docker stopはコンテナ名またはID"},
			},
		},
		{
			ID:            "rm-docker-rm",
			LinuxCommand:  "rm",
			DockerCommand: "docker rm",
			Category:      "file-operations",
			Description:   "ファイル/コンテナ削除",
			LinuxExample:  "rm file.txt",
			DockerExample: "docker rm container_name",
			Notes:         []string{"rmはファイル削除、docker rmはコンテナ削除"},
			LocalizedDescription: map[string]string{
				"en": "Remove files/containers",
				"ja": "ファイル/コンテナ削除",
			},
			LocalizedNotes: map[string][]string{
				"en": {"rm removes files, docker rm removes containers"},
				"ja": {"rmはファイル削除、docker rmはコンテナ削除"},
			},
		},
		{
			ID:            "tail-docker-logs",
			LinuxCommand:  "tail",
			DockerCommand: "docker logs",
			Category:      "logs-monitoring",
			Description:   "ログ表示",
			LinuxExample:  "tail -f /var/log/app.log",
			DockerExample: "docker logs -f container_name",
			Notes:         []string{"tailはファイル、docker logsはコンテナのログ"},
			LocalizedDescription: map[string]string{
				"en": "Display logs",
				"ja": "ログ表示",
			},
			LocalizedNotes: map[string][]string{
				"en": {"tail shows file content, docker logs shows container logs"},
				"ja": {"tailはファイル、docker logsはコンテナのログ"},
			},
		},
		{
			ID:            "cp-docker-cp",
			LinuxCommand:  "cp",
			DockerCommand: "docker cp",
			Category:      "file-operations",
			Description:   "ファイルコピー",
			LinuxExample:  "cp file.txt /dest/",
			DockerExample: "docker cp container_name:/file.txt /dest/",
			Notes:         []string{"docker cpはコンテナとホスト間でファイル転送"},
			LocalizedDescription: map[string]string{
				"en": "Copy files",
				"ja": "ファイルコピー",
			},
			LocalizedNotes: map[string][]string{
				"en": {"docker cp transfers files between container and host"},
				"ja": {"docker cpはコンテナとホスト間でファイル転送"},
			},
		},
		{
			ID:            "df-docker-system-df",
			LinuxCommand:  "df",
			DockerCommand: "docker system df",
			Category:      "system-information",
			Description:   "ディスク使用量表示",
			LinuxExample:  "df -h",
			DockerExample: "docker system df",
			Notes:         []string{"docker system dfはDockerのディスク使用量を表示"},
			LocalizedDescription: map[string]string{
				"en": "Display disk usage",
				"ja": "ディスク使用量表示",
			},
			LocalizedNotes: map[string][]string{
				"en": {"docker system df shows Docker disk usage"},
				"ja": {"docker system dfはDockerのディスク使用量を表示"},
			},
		},
		{
			ID:            "free-docker-stats-no-stream",
			LinuxCommand:  "free",
			DockerCommand: "docker stats --no-stream",
			Category:      "system-information",
			Description:   "メモリ使用量表示",
			LinuxExample:  "free -h",
			DockerExample: "docker stats --no-stream",
			Notes:         []string{"docker statsでコンテナのメモリ使用量を表示"},
			LocalizedDescription: map[string]string{
				"en": "Display memory usage",
				"ja": "メモリ使用量表示",
			},
			LocalizedNotes: map[string][]string{
				"en": {"docker stats shows container memory usage"},
				"ja": {"docker statsでコンテナのメモリ使用量を表示"},
			},
		},
		{
			ID:            "top-docker-stats",
			LinuxCommand:  "top",
			DockerCommand: "docker stats",
			Category:      "system-information",
			Description:   "リアルタイムシステム情報",
			LinuxExample:  "top",
			DockerExample: "docker stats",
			Notes:         []string{"topはシステム全体、docker statsはコンテナの統計情報"},
			LocalizedDescription: map[string]string{
				"en": "Real-time system information",
				"ja": "リアルタイムシステム情報",
			},
			LocalizedNotes: map[string][]string{
				"en": {"top shows system-wide info, docker stats shows container statistics"},
				"ja": {"topはシステム全体、docker statsはコンテナの統計情報"},
			},
		},
	}
}

// SaveMappings saves the current mappings to a file
func (engine *DefaultMappingEngine) SaveMappings() error {
	mappingsData := struct {
		Mappings []CommandMapping `yaml:"mappings"`
	}{
		Mappings: engine.mappings,
	}

	data, err := yaml.Marshal(mappingsData)
	if err != nil {
		return fmt.Errorf("failed to marshal mappings: %v", err)
	}

	dataFile := filepath.Join(engine.dataPath, "mappings.yaml")
	err = ioutil.WriteFile(dataFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write mappings file: %v", err)
	}

	return nil
}