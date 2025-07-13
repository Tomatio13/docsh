package config

import (
	"fmt"
	"strings"
)

func (c *Config) AddAlias(name, command string) {
	c.Aliases[name] = command
}

func (c *Config) RemoveAlias(name string) bool {
	if _, exists := c.Aliases[name]; exists {
		delete(c.Aliases, name)
		return true
	}
	return false
}

func (c *Config) ListAliases() {
	if len(c.Aliases) == 0 {
		fmt.Println("No aliases defined")
		return
	}
	
	fmt.Println("Defined aliases:")
	for name, command := range c.Aliases {
		fmt.Printf("  %s='%s'\n", name, command)
	}
}

func (c *Config) ExpandAlias(input string) string {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return input
	}
	
	command := parts[0]
	if aliasCommand, exists := c.Aliases[command]; exists {
		// エイリアスを展開
		expandedParts := strings.Fields(aliasCommand)
		if len(parts) > 1 {
			// 残りの引数を追加
			expandedParts = append(expandedParts, parts[1:]...)
		}
		return strings.Join(expandedParts, " ")
	}
	
	return input
}