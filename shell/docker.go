package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// isDockerContainer checks if the given name is a Docker container
func (s *Shell) isDockerContainer(name string) (bool, error) {
	if !s.shellExecutor.IsDockerAvailable() {
		return false, fmt.Errorf("Docker is not available")
	}

	// Check if container exists and is running
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}", "--filter", fmt.Sprintf("name=%s", name))
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// Check if the container name appears in the output
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
func (s *Shell) isContainerRunning(name string) (bool, error) {
	if !s.shellExecutor.IsDockerAvailable() {
		return false, fmt.Errorf("Docker is not available")
	}

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

// enterContainer executes docker exec -it <container> /bin/bash
func (s *Shell) enterContainer(containerName string) error {
	if !s.shellExecutor.IsDockerAvailable() {
		return fmt.Errorf("Docker is not available")
	}

	// Check if container exists
	exists, err := s.isDockerContainer(containerName)
	if err != nil {
		return fmt.Errorf("error checking container: %v", err)
	}
	if !exists {
		return fmt.Errorf("container '%s' not found", containerName)
	}

	// Check if container is running
	running, err := s.isContainerRunning(containerName)
	if err != nil {
		return fmt.Errorf("error checking container status: %v", err)
	}
	if !running {
		return fmt.Errorf("container '%s' is not running", containerName)
	}

	// Try different shells in order of preference
	shells := []string{"/bin/bash", "/bin/sh", "/bin/ash"}
	
	for _, shell := range shells {
		cmd := exec.Command("docker", "exec", "-it", containerName, shell)
		cmd.Stdin = s.getStdin()
		cmd.Stdout = s.getStdout()
		cmd.Stderr = s.getStderr()
		
		err := cmd.Run()
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to enter container '%s': no available shell", containerName)
}

// getStdin returns the standard input
func (s *Shell) getStdin() *os.File {
	return os.Stdin
}

// getStdout returns the standard output
func (s *Shell) getStdout() *os.File {
	return os.Stdout
}

// getStderr returns the standard error
func (s *Shell) getStderr() *os.File {
	return os.Stderr
}