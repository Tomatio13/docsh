package executor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"docknaut/internal/engine"
	"docknaut/internal/parser"
)

// ExecutionResult represents the result of command execution
type ExecutionResult struct {
	Command    string
	Output     string
	Error      string
	ExitCode   int
	Duration   time.Duration
	Mapping    *engine.CommandMapping
}

// ShellExecutor defines the interface for command execution
type ShellExecutor interface {
	Execute(ctx context.Context, cmd *parser.ParsedCommand) (*ExecutionResult, error)
	ExecuteWithMapping(ctx context.Context, mapping *engine.CommandMapping, args []string) (*ExecutionResult, error)
	DryRun(cmd *parser.ParsedCommand) (string, error)
	IsDockerAvailable() bool
}

// DefaultShellExecutor is the default implementation of ShellExecutor
type DefaultShellExecutor struct {
	mappingEngine engine.MappingEngine
	dryRunMode    bool
}

// NewShellExecutor creates a new shell executor instance
func NewShellExecutor(mappingEngine engine.MappingEngine) ShellExecutor {
	return &DefaultShellExecutor{
		mappingEngine: mappingEngine,
		dryRunMode:    false,
	}
}

// Execute executes a parsed command (Docker-only mode)
func (executor *DefaultShellExecutor) Execute(ctx context.Context, cmd *parser.ParsedCommand) (*ExecutionResult, error) {
	start := time.Now()
	result := &ExecutionResult{
		Command: cmd.Command,
	}

	// Check if dry run mode is enabled
	if executor.dryRunMode {
		dryRunOutput, err := executor.DryRun(cmd)
		if err != nil {
			result.Error = err.Error()
			result.ExitCode = 1
		} else {
			result.Output = dryRunOutput
			result.ExitCode = 0
		}
		result.Duration = time.Since(start)
		return result, nil
	}

	// Handle different command types
	if cmd.IsBuiltin {
		return executor.executeBuiltinCommand(ctx, cmd)
	}

	if cmd.IsLinux {
		// Try to find mapping for Linux command with options first
		mapping, err := executor.mappingEngine.FindByLinuxCommandWithOptions(cmd.Command, cmd.Options)
		if err == nil {
			result.Mapping = mapping
			return executor.ExecuteWithMapping(ctx, mapping, cmd.Args)
		}
		
		// Fallback to exact command match
		mapping, err = executor.mappingEngine.FindByLinuxCommand(cmd.Command)
		if err == nil {
			result.Mapping = mapping
			return executor.ExecuteWithMapping(ctx, mapping, cmd.Args)
		}
		
		// Docker-only mode: reject unmapped Linux commands
		result.Error = fmt.Sprintf("Command '%s' is not supported in Docker-only mode. Use 'mapping search %s' to find available Docker commands.", cmd.Command, cmd.Command)
		result.ExitCode = 1
		result.Duration = time.Since(start)
		return result, fmt.Errorf(result.Error)
	}

	if cmd.IsDocker {
		return executor.executeDockerCommand(ctx, cmd)
	}

	// Docker-only mode: reject unknown commands
	result.Error = fmt.Sprintf("Command '%s' is not recognized. This is a Docker-only shell. Use 'help' to see available commands.", cmd.Command)
	result.ExitCode = 1
	result.Duration = time.Since(start)
	return result, fmt.Errorf(result.Error)
}

// ExecuteWithMapping executes a command using a specific mapping
func (executor *DefaultShellExecutor) ExecuteWithMapping(ctx context.Context, mapping *engine.CommandMapping, args []string) (*ExecutionResult, error) {
	start := time.Now()
	result := &ExecutionResult{
		Command: mapping.DockerCommand,
		Mapping: mapping,
	}

	// Check if Docker is available
	if !executor.IsDockerAvailable() {
		result.Error = "Docker is not available or not running"
		result.ExitCode = 1
		result.Duration = time.Since(start)
		return result, fmt.Errorf("docker is not available")
	}

	// Special handling for cd command (container entry)
	if mapping.LinuxCommand == "cd" && len(args) > 0 {
		return executor.executeContainerEntry(ctx, args[0], result)
	}

	// Parse the Docker command
	dockerCmd := strings.Fields(mapping.DockerCommand)
	if len(args) > 0 {
		dockerCmd = append(dockerCmd, args...)
	}

	// Check if this is a streaming command
	isStreaming := isStreamingCommand(dockerCmd)
	
	if isStreaming {
		return executor.executeStreamingCommand(ctx, dockerCmd, result)
	}
	
	// Execute the Docker command (non-streaming)
	cmd := exec.CommandContext(ctx, dockerCmd[0], dockerCmd[1:]...)
	output, err := cmd.CombinedOutput()

	result.Output = string(output)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
		}
		return result, err
	}

	result.ExitCode = 0
	return result, nil
}

// DryRun shows what command would be executed without actually executing it
func (executor *DefaultShellExecutor) DryRun(cmd *parser.ParsedCommand) (string, error) {
	if cmd.IsLinux {
		mapping, err := executor.mappingEngine.FindByLinuxCommand(cmd.Command)
		if err == nil {
			dockerCmd := mapping.DockerCommand
			if len(cmd.Args) > 0 {
				dockerCmd = dockerCmd + " " + strings.Join(cmd.Args, " ")
			}
			return fmt.Sprintf("Would execute: %s\n\nMapping: %s -> %s\nDescription: %s\n",
				dockerCmd, mapping.LinuxCommand, mapping.DockerCommand, mapping.Description), nil
		}
	}

	commandLine := cmd.Command
	if len(cmd.Args) > 0 {
		commandLine = commandLine + " " + strings.Join(cmd.Args, " ")
	}
	return fmt.Sprintf("Would execute: %s", commandLine), nil
}

// IsDockerAvailable checks if Docker is available and running
func (executor *DefaultShellExecutor) IsDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	err := cmd.Run()
	return err == nil
}

// executeBuiltinCommand executes builtin commands
func (executor *DefaultShellExecutor) executeBuiltinCommand(ctx context.Context, cmd *parser.ParsedCommand) (*ExecutionResult, error) {
	start := time.Now()
	result := &ExecutionResult{
		Command: cmd.Command,
	}

	switch cmd.Command {
	case "help":
		result.Output = executor.getHelpText()
	case "version":
		result.Output = "Docknaut version 1.0.0"
	case "mapping":
		output, err := executor.handleMappingCommand(cmd.Args)
		if err != nil {
			result.Error = err.Error()
			result.ExitCode = 1
		} else {
			result.Output = output
		}
	default:
		result.Error = fmt.Sprintf("Unknown builtin command: %s", cmd.Command)
		result.ExitCode = 1
	}

	result.Duration = time.Since(start)
	return result, nil
}

// executeSystemCommand executes system commands
func (executor *DefaultShellExecutor) executeSystemCommand(ctx context.Context, cmd *parser.ParsedCommand) (*ExecutionResult, error) {
	start := time.Now()
	result := &ExecutionResult{
		Command: cmd.Command,
	}

	execCmd := exec.CommandContext(ctx, cmd.Command, cmd.Args...)
	output, err := execCmd.CombinedOutput()

	result.Output = string(output)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
		}
		return result, err
	}

	result.ExitCode = 0
	return result, nil
}

// executeDockerCommand executes Docker commands
func (executor *DefaultShellExecutor) executeDockerCommand(ctx context.Context, cmd *parser.ParsedCommand) (*ExecutionResult, error) {
	start := time.Now()
	result := &ExecutionResult{
		Command: cmd.Command,
	}

	// Check if Docker is available
	if !executor.IsDockerAvailable() {
		result.Error = "Docker is not available or not running"
		result.ExitCode = 1
		result.Duration = time.Since(start)
		return result, fmt.Errorf("docker is not available")
	}

	// Prepend "docker" if not already present
	args := cmd.Args
	if cmd.Command != "docker" {
		args = append([]string{cmd.Command}, args...)
	}

	execCmd := exec.CommandContext(ctx, "docker", args...)
	output, err := execCmd.CombinedOutput()

	result.Output = string(output)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
		}
		return result, err
	}

	result.ExitCode = 0
	return result, nil
}

// handleMappingCommand handles mapping-related commands
func (executor *DefaultShellExecutor) handleMappingCommand(args []string) (string, error) {
	if len(args) == 0 {
		return executor.listAllMappings(), nil
	}

	switch args[0] {
	case "list":
		if len(args) > 1 {
			return executor.listMappingsByCategory(args[1])
		}
		return executor.listAllMappings(), nil
	case "search":
		if len(args) > 1 {
			return executor.searchMappings(strings.Join(args[1:], " "))
		}
		return "", fmt.Errorf("search requires a query")
	case "show":
		if len(args) > 1 {
			return executor.showMapping(args[1])
		}
		return "", fmt.Errorf("show requires a command name")
	default:
		return "", fmt.Errorf("unknown mapping command: %s", args[0])
	}
}

// listAllMappings lists all available mappings
func (executor *DefaultShellExecutor) listAllMappings() string {
	var output strings.Builder

	output.WriteString("Available Command Mappings:\n\n")
	
	categories := executor.mappingEngine.GetCategories()
	for _, category := range categories {
		output.WriteString(fmt.Sprintf("=== %s ===\n", strings.ToUpper(category)))
		categoryMappings, _ := executor.mappingEngine.ListByCategory(category)
		for _, mapping := range categoryMappings {
			output.WriteString(fmt.Sprintf("  %s -> %s\n", mapping.LinuxCommand, mapping.DockerCommand))
			output.WriteString(fmt.Sprintf("    %s\n", mapping.Description))
		}
		output.WriteString("\n")
	}

	return output.String()
}

// listMappingsByCategory lists mappings for a specific category
func (executor *DefaultShellExecutor) listMappingsByCategory(category string) (string, error) {
	mappings, err := executor.mappingEngine.ListByCategory(category)
	if err != nil {
		return "", err
	}

	if len(mappings) == 0 {
		return fmt.Sprintf("No mappings found for category: %s", category), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Mappings for category '%s':\n\n", category))

	for _, mapping := range mappings {
		output.WriteString(fmt.Sprintf("%s -> %s\n", mapping.LinuxCommand, mapping.DockerCommand))
		output.WriteString(fmt.Sprintf("  Description: %s\n", mapping.Description))
		output.WriteString(fmt.Sprintf("  Example: %s\n", mapping.DockerExample))
		if len(mapping.Notes) > 0 {
			output.WriteString(fmt.Sprintf("  Notes: %s\n", strings.Join(mapping.Notes, ", ")))
		}
		output.WriteString("\n")
	}

	return output.String(), nil
}

// searchMappings searches for mappings by query
func (executor *DefaultShellExecutor) searchMappings(query string) (string, error) {
	mappings, err := executor.mappingEngine.SearchCommands(query)
	if err != nil {
		return "", err
	}

	if len(mappings) == 0 {
		return fmt.Sprintf("No mappings found for query: %s", query), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Search results for '%s':\n\n", query))

	for _, mapping := range mappings {
		output.WriteString(fmt.Sprintf("%s -> %s\n", mapping.LinuxCommand, mapping.DockerCommand))
		output.WriteString(fmt.Sprintf("  Category: %s\n", mapping.Category))
		output.WriteString(fmt.Sprintf("  Description: %s\n", mapping.Description))
		output.WriteString("\n")
	}

	return output.String(), nil
}

// showMapping shows details for a specific mapping
func (executor *DefaultShellExecutor) showMapping(command string) (string, error) {
	mapping, err := executor.mappingEngine.FindByLinuxCommand(command)
	if err != nil {
		mapping, err = executor.mappingEngine.FindByDockerCommand(command)
		if err != nil {
			return "", fmt.Errorf("no mapping found for command: %s", command)
		}
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Mapping Details for '%s':\n\n", command))
	output.WriteString(fmt.Sprintf("Linux Command: %s\n", mapping.LinuxCommand))
	output.WriteString(fmt.Sprintf("Docker Command: %s\n", mapping.DockerCommand))
	output.WriteString(fmt.Sprintf("Category: %s\n", mapping.Category))
	output.WriteString(fmt.Sprintf("Description: %s\n", mapping.Description))
	output.WriteString(fmt.Sprintf("Linux Example: %s\n", mapping.LinuxExample))
	output.WriteString(fmt.Sprintf("Docker Example: %s\n", mapping.DockerExample))

	if len(mapping.Notes) > 0 {
		output.WriteString("\nNotes:\n")
		for _, note := range mapping.Notes {
			output.WriteString(fmt.Sprintf("  - %s\n", note))
		}
	}

	if len(mapping.Warnings) > 0 {
		output.WriteString("\nWarnings:\n")
		for _, warning := range mapping.Warnings {
			output.WriteString(fmt.Sprintf("  ⚠️  %s\n", warning))
		}
	}

	return output.String(), nil
}

// getHelpText returns help text for the shell
func (executor *DefaultShellExecutor) getHelpText() string {
	return `Docknaut - Docker Command Mapping Shell

Available commands:
  help                    Show this help message
  version                 Show version information
  mapping [list|search|show] <args>  Manage command mappings
  exit                    Exit the shell

Command mapping examples:
  ls                      -> docker ps
  ps                      -> docker ps
  kill <container>        -> docker stop <container>
  rm <container>          -> docker rm <container>
  tail -f <container>     -> docker logs -f <container>

For more information about specific mappings, use:
  mapping show <command>
`
}

// executeStreamingCommand executes streaming commands like docker logs -f with real-time output
func (executor *DefaultShellExecutor) executeStreamingCommand(ctx context.Context, dockerCmd []string, result *ExecutionResult) (*ExecutionResult, error) {
	start := time.Now()
	
	// Create the command with its own process group
	cmd := exec.CommandContext(ctx, dockerCmd[0], dockerCmd[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	
	// Setup pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		result.Error = err.Error()
		result.ExitCode = 1
		result.Duration = time.Since(start)
		return result, err
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		result.Error = err.Error()
		result.ExitCode = 1
		result.Duration = time.Since(start)
		return result, err
	}
	
	// Start the command
	if err := cmd.Start(); err != nil {
		result.Error = err.Error()
		result.ExitCode = 1
		result.Duration = time.Since(start)
		return result, err
	}
	
	// Setup dedicated signal handling for this process
	sigChan := make(chan os.Signal, 1)
	// Stop any existing signal notifications to avoid conflicts with go-prompt
	signal.Stop(sigChan)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	// Channel to signal when command is done
	done := make(chan error, 1)
	outputDone := make(chan bool, 2)
	
	// Goroutine to handle stdout
	go func() {
		defer func() { outputDone <- true }()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Println(scanner.Text())
			}
		}
	}()
	
	// Goroutine to handle stderr
	go func() {
		defer func() { outputDone <- true }()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Fprintln(os.Stderr, scanner.Text())
			}
		}
	}()
	
	// Goroutine to wait for command completion
	go func() {
		done <- cmd.Wait()
	}()
	
	// Wait for either completion or cancellation
	select {
	case <-ctx.Done():
		// Context was cancelled
		executor.terminateCommand(cmd)
		result.Error = "Command interrupted by user"
		result.ExitCode = 130
		result.Duration = time.Since(start)
		return result, fmt.Errorf("command interrupted")
		
	case err := <-done:
		// Command completed normally
		signal.Stop(sigChan) // Clean up signal handling
		result.Duration = time.Since(start)
		
		if err != nil {
			result.Error = err.Error()
			if exitError, ok := err.(*exec.ExitError); ok {
				result.ExitCode = exitError.ExitCode()
			} else {
				result.ExitCode = 1
			}
			return result, err
		}
		
		result.ExitCode = 0
		return result, nil
		
	case <-sigChan:
		// Signal received directly
		fmt.Println("\n^C")
		executor.terminateCommand(cmd)
		signal.Stop(sigChan) // Clean up signal handling
		result.Error = "Command interrupted by signal"
		result.ExitCode = 130
		result.Duration = time.Since(start)
		return result, fmt.Errorf("command interrupted by signal")
	}
}

// terminateCommand forcibly terminates a command and its process group
func (executor *DefaultShellExecutor) terminateCommand(cmd *exec.Cmd) {
	if cmd.Process != nil {
		// First try to terminate the process group
		if cmd.Process.Pid > 0 {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		}
		
		// Give it a moment to terminate gracefully
		time.Sleep(100 * time.Millisecond)
		
		// If still running, force kill the process group
		if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
			if cmd.Process.Pid > 0 {
				syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			}
			cmd.Process.Kill()
		}
	}
}

// isStreamingCommand checks if a Docker command is a streaming command
func isStreamingCommand(dockerCmd []string) bool {
	if len(dockerCmd) < 2 {
		return false
	}
	
	// Check for docker logs -f
	if dockerCmd[0] == "docker" && dockerCmd[1] == "logs" {
		for _, arg := range dockerCmd[2:] {
			if arg == "-f" || arg == "--follow" {
				return true
			}
		}
	}
	
	// Check for docker attach
	if dockerCmd[0] == "docker" && dockerCmd[1] == "attach" {
		return true
	}
	
	// Check for docker exec with interactive/tty flags
	if dockerCmd[0] == "docker" && dockerCmd[1] == "exec" {
		for _, arg := range dockerCmd[2:] {
			if arg == "-it" || (arg == "-i" || arg == "-t") {
				return true
			}
		}
	}
	
	// Check for docker stats (enhanced logic for --no-stream)
	if dockerCmd[0] == "docker" && dockerCmd[1] == "stats" {
		// Check if --no-stream is present
		for _, arg := range dockerCmd[2:] {
			if arg == "--no-stream" {
				return false // Not streaming if --no-stream is present
			}
		}
		return true // Default docker stats is streaming
	}
	
	return false
}

// executeContainerEntry handles cd command to enter containers
func (executor *DefaultShellExecutor) executeContainerEntry(ctx context.Context, containerName string, result *ExecutionResult) (*ExecutionResult, error) {
	start := time.Now()
	
	// Check if container exists and is running
	isContainer, err := executor.isDockerContainer(containerName)
	if err != nil {
		result.Error = fmt.Sprintf("Error checking container: %v", err)
		result.ExitCode = 1
		result.Duration = time.Since(start)
		return result, err
	}
	
	if !isContainer {
		result.Error = fmt.Sprintf("Container '%s' not found", containerName)
		result.ExitCode = 1
		result.Duration = time.Since(start)
		return result, fmt.Errorf("container not found")
	}
	
	// Check if container is running
	running, err := executor.isContainerRunning(containerName)
	if err != nil {
		result.Error = fmt.Sprintf("Error checking container status: %v", err)
		result.ExitCode = 1
		result.Duration = time.Since(start)
		return result, err
	}
	
	if !running {
		result.Error = fmt.Sprintf("Container '%s' is not running", containerName)
		result.ExitCode = 1
		result.Duration = time.Since(start)
		return result, fmt.Errorf("container not running")
	}
	
	// Try different shells in order of preference
	shells := []string{"/bin/bash", "/bin/sh", "/bin/ash"}
	
	// Check if TTY is available
	var dockerFlags []string
	if executor.isTTYAvailable() {
		dockerFlags = []string{"exec", "-it"}
	} else {
		dockerFlags = []string{"exec", "-i"}
	}
	
	for _, shell := range shells {
		args := append(dockerFlags, containerName, shell)
		cmd := exec.CommandContext(ctx, "docker", args...)
		
		// In non-TTY environment, don't try to actually execute the command
		if !executor.isTTYAvailable() {
			result.Output = fmt.Sprintf("Would enter container: %s (non-interactive mode)", containerName)
			result.ExitCode = 0
			result.Duration = time.Since(start)
			return result, nil
		}
		
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		err := cmd.Run()
		if err == nil {
			result.Output = fmt.Sprintf("Entered container: %s", containerName)
			result.ExitCode = 0
			result.Duration = time.Since(start)
			return result, nil
		}
	}
	
	result.Error = fmt.Sprintf("Failed to enter container '%s': no available shell", containerName)
	result.ExitCode = 1
	result.Duration = time.Since(start)
	return result, fmt.Errorf("no available shell")
}

// isDockerContainer checks if the given name is a Docker container
func (executor *DefaultShellExecutor) isDockerContainer(name string) (bool, error) {
	// Check if container exists (running)
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}", "--filter", fmt.Sprintf("name=%s", name))
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	
	containerNames := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, containerName := range containerNames {
		if containerName == name {
			return true, nil
		}
	}
	
	// Also check stopped containers
	cmd = exec.Command("docker", "ps", "-a", "--format", "{{.Names}}", "--filter", fmt.Sprintf("name=%s", name))
	output, err = cmd.Output()
	if err != nil {
		return false, err
	}
	
	containerNames = strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, containerName := range containerNames {
		if containerName == name {
			return true, nil
		}
	}
	
	return false, nil
}

// isContainerRunning checks if a container is currently running
func (executor *DefaultShellExecutor) isContainerRunning(name string) (bool, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}", "--filter", fmt.Sprintf("name=%s", name))
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	
	containerNames := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, containerName := range containerNames {
		if containerName == name {
			return true, nil
		}
	}
	
	return false, nil
}

// SetDryRunMode enables or disables dry run mode
func (executor *DefaultShellExecutor) SetDryRunMode(enabled bool) {
	executor.dryRunMode = enabled
}

// isTTYAvailable checks if TTY is available for interactive commands
func (executor *DefaultShellExecutor) isTTYAvailable() bool {
	// Check if stdin is a terminal
	if fileInfo, err := os.Stdin.Stat(); err == nil {
		// Check if it's a character device (TTY)
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	}
	return false
}