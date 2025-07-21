package parser

import (
	"strings"
)

// ParsedCommand represents a parsed command with its components
type ParsedCommand struct {
	Command   string
	Args      []string
	Options   map[string]string
	IsDocker  bool
	IsLinux   bool
	IsBuiltin bool
}

// CommandParser defines the interface for command parsing
type CommandParser interface {
	ParseCommand(input string) (*ParsedCommand, error)
	IsLinuxCommand(cmd string) bool
	IsDockerCommand(cmd string) bool
	IsBuiltinCommand(cmd string) bool
}

// DefaultCommandParser is the default implementation of CommandParser
type DefaultCommandParser struct {
	linuxCommands   []string
	dockerCommands  []string
	builtinCommands []string
}

// NewCommandParser creates a new command parser instance
func NewCommandParser() CommandParser {
	return &DefaultCommandParser{
		linuxCommands: []string{
			"ls", "ps", "kill", "rm", "cp", "mv", "cd", "pwd", "cat", "grep", "find",
			"tail", "head", "top", "df", "du", "free", "which", "whoami", "uname",
			"chmod", "chown", "mkdir", "rmdir", "touch", "ln", "tar", "gzip", "gunzip",
			"wget", "curl", "ping", "netstat", "ssh", "scp", "rsync", "cron", "history",
		},
		dockerCommands: []string{
			"docker", "run", "build", "pull", "push", "ps", "images", "rmi", "rm",
			"start", "stop", "restart", "kill", "logs", "exec", "cp", "commit",
			"create", "pause", "unpause", "wait", "export", "import", "save", "load",
			"tag", "inspect", "stats", "top", "port", "network", "volume", "system",
		},
		builtinCommands: []string{
			"exit", "quit", "help", "version", "alias", "unalias", "theme", "lang",
			"config", "mapping", "search", "list", "pwd", "clear", "cls",
		},
	}
}

// ParseCommand parses a command string and returns a ParsedCommand
func (parser *DefaultCommandParser) ParseCommand(input string) (*ParsedCommand, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	// Split the input into parts
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, nil
	}

	command := parts[0]
	args := parts[1:]

	// Parse options
	options := make(map[string]string)
	var filteredArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			// Long option
			if strings.Contains(arg, "=") {
				keyValue := strings.SplitN(arg[2:], "=", 2)
				options[keyValue[0]] = keyValue[1]
			} else {
				options[arg[2:]] = "true"
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			// Short option
			optionKey := arg[1:]
			// Check if this is a flag-like option (no value expected)
			if isFlagOption(optionKey) {
				options[optionKey] = "true"
			} else if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				options[optionKey] = args[i+1]
				i++ // Skip the next argument
			} else {
				options[optionKey] = "true"
			}
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	parsed := &ParsedCommand{
		Command:   command,
		Args:      filteredArgs,
		Options:   options,
		IsDocker:  parser.IsDockerCommand(command),
		IsLinux:   parser.IsLinuxCommand(command),
		IsBuiltin: parser.IsBuiltinCommand(command),
	}

	return parsed, nil
}

// IsLinuxCommand checks if a command is a Linux command
func (parser *DefaultCommandParser) IsLinuxCommand(cmd string) bool {
	for _, linuxCmd := range parser.linuxCommands {
		if cmd == linuxCmd {
			return true
		}
	}
	return false
}

// IsDockerCommand checks if a command is a Docker command
func (parser *DefaultCommandParser) IsDockerCommand(cmd string) bool {
	if cmd == "docker" {
		return true
	}
	for _, dockerCmd := range parser.dockerCommands {
		if cmd == dockerCmd {
			return true
		}
	}
	return false
}

// IsBuiltinCommand checks if a command is a builtin command
func (parser *DefaultCommandParser) IsBuiltinCommand(cmd string) bool {
	for _, builtinCmd := range parser.builtinCommands {
		if cmd == builtinCmd {
			return true
		}
	}
	return false
}

// GetCommandType returns the type of command (linux, docker, builtin, or unknown)
func (parser *DefaultCommandParser) GetCommandType(cmd string) string {
	if parser.IsBuiltinCommand(cmd) {
		return "builtin"
	}
	if parser.IsDockerCommand(cmd) {
		return "docker"
	}
	if parser.IsLinuxCommand(cmd) {
		return "linux"
	}
	return "unknown"
}

// SuggestCommands suggests similar commands based on input
func (parser *DefaultCommandParser) SuggestCommands(input string) []string {
	input = strings.ToLower(input)
	var suggestions []string

	// Check builtin commands first
	for _, cmd := range parser.builtinCommands {
		if strings.Contains(strings.ToLower(cmd), input) {
			suggestions = append(suggestions, cmd)
		}
	}

	// Check Linux commands
	for _, cmd := range parser.linuxCommands {
		if strings.Contains(strings.ToLower(cmd), input) {
			suggestions = append(suggestions, cmd)
		}
	}

	// Check Docker commands
	for _, cmd := range parser.dockerCommands {
		if strings.Contains(strings.ToLower(cmd), input) {
			suggestions = append(suggestions, cmd)
		}
	}

	// Limit suggestions to prevent overwhelming output
	if len(suggestions) > 10 {
		suggestions = suggestions[:10]
	}

	return suggestions
}

// isFlagOption checks if an option is a flag (doesn't take a value)
func isFlagOption(option string) bool {
	flagOptions := []string{
		"f", "follow", // for tail -f
		"a", "all",    // for ls -a, ps -a
		"l", "long",   // for ls -l
		"h", "help",   // for --help
		"v", "verbose", // for --verbose
		"q", "quiet",  // for --quiet
		"r", "recursive", // for -r
		"i", "interactive", // for -i
		"force", // for --force
		"dry-run", // for --dry-run
	}
	
	for _, flag := range flagOptions {
		if option == flag {
			return true
		}
	}
	return false
}