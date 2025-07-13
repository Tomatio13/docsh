package config

import (
	"fmt"
	"strings"

	"cherrysh/i18n"
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
		fmt.Println(i18n.T("alias.no_aliases"))
		return
	}

	fmt.Println(i18n.T("alias.defined_aliases"))
	for name, command := range c.Aliases {
		fmt.Printf(i18n.T("alias.alias_format")+"\n", name, command)
	}
}

func (c *Config) ParseAlias(aliasDef string) error {
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
