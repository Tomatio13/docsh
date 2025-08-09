package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// YAMLConfig represents the YAML configuration structure
type YAMLConfig struct {
	Shell struct {
		Prompt       string `yaml:"prompt"`
		HistorySize  int    `yaml:"history_size"`
		AutoComplete bool   `yaml:"auto_complete"`
		DryRunMode   bool   `yaml:"dry_run_mode"`
		ShowMappings bool   `yaml:"show_mappings"`
	} `yaml:"shell"`

	// Banner section
	Banner struct {
		Enabled bool   `yaml:"enabled"`
		Style   string `yaml:"style"`
	} `yaml:"banner"`

	Mapping struct {
		DataFile     string `yaml:"data_file"`
		CacheEnabled bool   `yaml:"cache_enabled"`
		AutoSuggest  bool   `yaml:"auto_suggest"`
	} `yaml:"mapping"`

	Docker struct {
		DefaultOptions []string `yaml:"default_options"`
		Timeout        int      `yaml:"timeout"`
		AutoDetect     bool     `yaml:"auto_detect"`
	} `yaml:"docker"`

	Display struct {
		ShowWarnings     bool `yaml:"show_warnings"`
		ColorOutput      bool `yaml:"color_output"`
		VerboseMode      bool `yaml:"verbose_mode"`
		ShowExamples     bool `yaml:"show_examples"`
		ShowDescriptions bool `yaml:"show_descriptions"`
	} `yaml:"display"`

	I18n struct {
		DefaultLanguage    string   `yaml:"default_language"`
		SupportedLanguages []string `yaml:"supported_languages"`
		LocaleDir          string   `yaml:"locale_dir"`
		FallbackLanguage   string   `yaml:"fallback_language"`
	} `yaml:"i18n"`

	Features struct {
		Aliases           bool `yaml:"aliases"`
		ContextManagement bool `yaml:"context_management"`
		History           bool `yaml:"history"`
		Completion        bool `yaml:"completion"`
		CommandMapping    bool `yaml:"command_mapping"`
		GitIntegration    bool `yaml:"git_integration"`
	} `yaml:"features"`

	Aliases map[string]string `yaml:"aliases"`

	Context struct {
		CurrentContainer string   `yaml:"current_container"`
		RecentContainers []string `yaml:"recent_containers"`
		AutoSwitch       bool     `yaml:"auto_switch"`
		ShowInPrompt     bool     `yaml:"show_in_prompt"`
	} `yaml:"context"`

	History struct {
		MaxEntries        int    `yaml:"max_entries"`
		SaveToFile        bool   `yaml:"save_to_file"`
		SearchEnabled     bool   `yaml:"search_enabled"`
		DuplicateHandling string `yaml:"duplicate_handling"`
	} `yaml:"history"`

	Completion struct {
		Enabled        bool `yaml:"enabled"`
		ContainerNames bool `yaml:"container_names"`
		ImageNames     bool `yaml:"image_names"`
		CommandOptions bool `yaml:"command_options"`
		FilePaths      bool `yaml:"file_paths"`
		MaxSuggestions int  `yaml:"max_suggestions"`
	} `yaml:"completion"`

	Themes struct {
		Default   string `yaml:"default"`
		Available []struct {
			Name   string            `yaml:"name"`
			Prompt string            `yaml:"prompt"`
			Colors map[string]string `yaml:"colors"`
		} `yaml:"available"`
	} `yaml:"themes"`
}

// LoadYAMLConfig loads configuration from YAML file
func (c *Config) LoadYAMLConfig(dataPath string) error {
	configPath := filepath.Join(dataPath, "config.yaml")

	// Check if YAML config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // No YAML config file, continue with existing config
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read YAML config file: %v", err)
	}

	var yamlConfig YAMLConfig
	err = yaml.Unmarshal(data, &yamlConfig)
	if err != nil {
		return fmt.Errorf("failed to parse YAML config file: %v", err)
	}

	// Merge YAML config with existing config
	c.mergeYAMLConfig(&yamlConfig)

	return nil
}

// mergeYAMLConfig merges YAML configuration with existing config
func (c *Config) mergeYAMLConfig(yamlConfig *YAMLConfig) {
	// Override language if not set in traditional config
	if c.Language == "" && yamlConfig.I18n.DefaultLanguage != "" {
		c.Language = yamlConfig.I18n.DefaultLanguage
	}

	// Override theme if not set in traditional config
	if c.Theme == "default" && yamlConfig.Themes.Default != "" {
		c.Theme = yamlConfig.Themes.Default
	}

	// Merge aliases
	if yamlConfig.Aliases != nil {
		for name, command := range yamlConfig.Aliases {
			// Only add if not already defined in traditional config
			if _, exists := c.Aliases[name]; !exists {
				c.Aliases[name] = command
			}
		}
	}

	// Banner settings
	if yamlConfig.Banner.Enabled {
		c.BannerEnabled = true
	}
	if yamlConfig.Banner.Style != "" {
		c.BannerStyle = yamlConfig.Banner.Style
	}
}

// SaveYAMLConfig saves current configuration to YAML file
func (c *Config) SaveYAMLConfig(dataPath string) error {
	configPath := filepath.Join(dataPath, "config.yaml")

	// Create a YAML config structure from current config
	yamlConfig := YAMLConfig{}

	// Basic shell configuration
	yamlConfig.Shell.Prompt = "🐳 docsh> "
	yamlConfig.Shell.HistorySize = 1000
	yamlConfig.Shell.AutoComplete = true
	yamlConfig.Shell.DryRunMode = false
	yamlConfig.Shell.ShowMappings = true

	// Banner configuration
	yamlConfig.Banner.Enabled = c.BannerEnabled
	yamlConfig.Banner.Style = c.BannerStyle

	// Mapping configuration
	yamlConfig.Mapping.DataFile = "data/mappings.yaml"
	yamlConfig.Mapping.CacheEnabled = true
	yamlConfig.Mapping.AutoSuggest = true

	// Docker configuration
	yamlConfig.Docker.DefaultOptions = []string{}
	yamlConfig.Docker.Timeout = 30
	yamlConfig.Docker.AutoDetect = true

	// Display configuration
	yamlConfig.Display.ShowWarnings = true
	yamlConfig.Display.ColorOutput = true
	yamlConfig.Display.VerboseMode = false
	yamlConfig.Display.ShowExamples = true
	yamlConfig.Display.ShowDescriptions = true

	// Internationalization
	yamlConfig.I18n.DefaultLanguage = c.Language
	yamlConfig.I18n.SupportedLanguages = []string{"ja", "en"}
	yamlConfig.I18n.LocaleDir = "data/locales"
	yamlConfig.I18n.FallbackLanguage = "en"

	// Features
	yamlConfig.Features.Aliases = true
	yamlConfig.Features.ContextManagement = true
	yamlConfig.Features.History = true
	yamlConfig.Features.Completion = true
	yamlConfig.Features.CommandMapping = true
	yamlConfig.Features.GitIntegration = true

	// Aliases
	yamlConfig.Aliases = c.Aliases

	// Context
	yamlConfig.Context.CurrentContainer = ""
	yamlConfig.Context.RecentContainers = []string{}
	yamlConfig.Context.AutoSwitch = true
	yamlConfig.Context.ShowInPrompt = true

	// History
	yamlConfig.History.MaxEntries = 1000
	yamlConfig.History.SaveToFile = true
	yamlConfig.History.SearchEnabled = true
	yamlConfig.History.DuplicateHandling = "ignore"

	// Completion
	yamlConfig.Completion.Enabled = true
	yamlConfig.Completion.ContainerNames = true
	yamlConfig.Completion.ImageNames = true
	yamlConfig.Completion.CommandOptions = true
	yamlConfig.Completion.FilePaths = true
	yamlConfig.Completion.MaxSuggestions = 16

	// Themes
	yamlConfig.Themes.Default = c.Theme

	// Marshal to YAML
	data, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML config: %v", err)
	}

	// Write to file
	err = ioutil.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write YAML config file: %v", err)
	}

	return nil
}

// GetYAMLConfigPath returns the path to the YAML configuration file
func (c *Config) GetYAMLConfigPath(dataPath string) string {
	return filepath.Join(dataPath, "config.yaml")
}
