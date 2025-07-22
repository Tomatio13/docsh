package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	
	"docknaut/i18n"
)

// isDockerContainer checks if the given name or ID is a Docker container
func (s *Shell) isDockerContainer(nameOrID string) (bool, error) {
	if !s.shellExecutor.IsDockerAvailable() {
		return false, fmt.Errorf("Docker is not available")
	}

	// Check if container exists (by name or ID) - running containers
	cmd := exec.Command("docker", "ps", "--format", "{{.ID}}\t{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			containerID := parts[0]
			containerName := parts[1]
			// Check if nameOrID matches either ID or name (or partial ID)
			if containerID == nameOrID || containerName == nameOrID || strings.HasPrefix(containerID, nameOrID) {
				return true, nil
			}
		}
	}

	// Also check stopped containers
	cmd = exec.Command("docker", "ps", "-a", "--format", "{{.ID}}\t{{.Names}}")
	output, err = cmd.Output()
	if err != nil {
		return false, err
	}

	lines = strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			containerID := parts[0]
			containerName := parts[1]
			// Check if nameOrID matches either ID or name (or partial ID)
			if containerID == nameOrID || containerName == nameOrID || strings.HasPrefix(containerID, nameOrID) {
				return true, nil
			}
		}
	}

	return false, nil
}

// isContainerRunning checks if a container is currently running
func (s *Shell) isContainerRunning(nameOrID string) (bool, error) {
	if !s.shellExecutor.IsDockerAvailable() {
		return false, fmt.Errorf("Docker is not available")
	}

	cmd := exec.Command("docker", "ps", "--format", "{{.ID}}\t{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			containerID := parts[0]
			containerName := parts[1]
			// Check if nameOrID matches either ID or name (or partial ID)
			if containerID == nameOrID || containerName == nameOrID || strings.HasPrefix(containerID, nameOrID) {
				return true, nil
			}
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

// Docker lifecycle management functions

// pullImage pulls a Docker image
func (s *Shell) pullImage(imageName string) error {
	if !s.shellExecutor.IsDockerAvailable() {
		return fmt.Errorf(i18n.T("docker.not_available"))
	}

	if imageName == "" {
		return fmt.Errorf(i18n.T("docker.image_name_required"))
	}

	fmt.Printf(i18n.T("docker.pull_image")+"\n", imageName)
	cmd := exec.Command("docker", "pull", imageName)
	cmd.Stdout = s.getStdout()
	cmd.Stderr = s.getStderr()

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf(i18n.T("docker.pull_failed"), imageName, err)
	}

	fmt.Printf(i18n.T("docker.pull_success")+"\n", imageName)
	return nil
}

// startContainer starts a stopped container
func (s *Shell) startContainer(containerName string) error {
	if !s.shellExecutor.IsDockerAvailable() {
		return fmt.Errorf(i18n.T("docker.not_available"))
	}

	if containerName == "" {
		return fmt.Errorf(i18n.T("docker.container_name_required"))
	}

	// Check if container exists
	exists, err := s.isDockerContainer(containerName)
	if err != nil {
		return fmt.Errorf(i18n.T("docker.error_checking_container"), err)
	}
	if !exists {
		return fmt.Errorf(i18n.T("docker.container_not_found"), containerName)
	}

	// Check if container is already running
	running, err := s.isContainerRunning(containerName)
	if err != nil {
		return fmt.Errorf(i18n.T("docker.error_checking_container_status"), err)
	}
	if running {
		return fmt.Errorf(i18n.T("docker.start_already_running"), containerName)
	}

	fmt.Printf(i18n.T("docker.start_container")+"\n", containerName)
	cmd := exec.Command("docker", "start", containerName)
	cmd.Stdout = s.getStdout()
	cmd.Stderr = s.getStderr()

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(i18n.T("docker.start_failed"), containerName, err)
	}

	fmt.Printf(i18n.T("docker.start_success")+"\n", containerName)
	return nil
}

// execInContainer executes a command in a running container
func (s *Shell) execInContainer(containerName string, command []string) error {
	if !s.shellExecutor.IsDockerAvailable() {
		return fmt.Errorf(i18n.T("docker.not_available"))
	}

	if containerName == "" {
		return fmt.Errorf(i18n.T("docker.container_name_required"))
	}

	if len(command) == 0 {
		return fmt.Errorf(i18n.T("docker.command_required"))
	}

	// Check if container exists
	exists, err := s.isDockerContainer(containerName)
	if err != nil {
		return fmt.Errorf(i18n.T("docker.error_checking_container"), err)
	}
	if !exists {
		return fmt.Errorf(i18n.T("docker.container_not_found"), containerName)
	}

	// Check if container is running
	running, err := s.isContainerRunning(containerName)
	if err != nil {
		return fmt.Errorf(i18n.T("docker.error_checking_container_status"), err)
	}
	if !running {
		return fmt.Errorf(i18n.T("docker.stop_not_running"), containerName)
	}

	// Build docker exec command
	dockerCmd := []string{"docker", "exec", "-it", containerName}
	dockerCmd = append(dockerCmd, command...)

	fmt.Printf(i18n.T("docker.exec_command")+"\n", containerName, strings.Join(command, " "))
	
	cmd := exec.Command(dockerCmd[0], dockerCmd[1:]...)
	cmd.Stdin = s.getStdin()
	cmd.Stdout = s.getStdout()
	cmd.Stderr = s.getStderr()

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(i18n.T("docker.exec_failed"), containerName, err)
	}

	return nil
}

// stopContainer stops a running container
func (s *Shell) stopContainer(containerName string) error {
	if !s.shellExecutor.IsDockerAvailable() {
		return fmt.Errorf(i18n.T("docker.not_available"))
	}

	if containerName == "" {
		return fmt.Errorf(i18n.T("docker.container_name_required"))
	}

	// Check if container exists
	exists, err := s.isDockerContainer(containerName)
	if err != nil {
		return fmt.Errorf(i18n.T("docker.error_checking_container"), err)
	}
	if !exists {
		return fmt.Errorf(i18n.T("docker.container_not_found"), containerName)
	}

	// Check if container is running
	running, err := s.isContainerRunning(containerName)
	if err != nil {
		return fmt.Errorf(i18n.T("docker.error_checking_container_status"), err)
	}
	if !running {
		return fmt.Errorf(i18n.T("docker.stop_not_running"), containerName)
	}

	fmt.Printf(i18n.T("docker.stop_container")+"\n", containerName)
	cmd := exec.Command("docker", "stop", containerName)
	cmd.Stdout = s.getStdout()
	cmd.Stderr = s.getStderr()

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(i18n.T("docker.stop_failed"), containerName, err)
	}

	fmt.Printf(i18n.T("docker.stop_success")+"\n", containerName)
	return nil
}

// removeContainer removes a container (must be stopped first)
func (s *Shell) removeContainer(containerName string, force bool) error {
	if !s.shellExecutor.IsDockerAvailable() {
		return fmt.Errorf(i18n.T("docker.not_available"))
	}

	if containerName == "" {
		return fmt.Errorf(i18n.T("docker.container_name_required"))
	}

	// Check if container exists
	exists, err := s.isDockerContainer(containerName)
	if err != nil {
		return fmt.Errorf(i18n.T("docker.error_checking_container"), err)
	}
	if !exists {
		return fmt.Errorf(i18n.T("docker.container_not_found"), containerName)
	}

	// If not forcing, check if container is running
	if !force {
		running, err := s.isContainerRunning(containerName)
		if err != nil {
			return fmt.Errorf(i18n.T("docker.error_checking_container_status"), err)
		}
		if running {
			return fmt.Errorf(i18n.T("docker.remove_running_container"), containerName)
		}
	}

	fmt.Printf(i18n.T("docker.remove_container")+"\n", containerName)
	
	var cmd *exec.Cmd
	if force {
		cmd = exec.Command("docker", "rm", "-f", containerName)
	} else {
		cmd = exec.Command("docker", "rm", containerName)
	}
	
	cmd.Stdout = s.getStdout()
	cmd.Stderr = s.getStderr()

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(i18n.T("docker.remove_failed"), containerName, err)
	}

	fmt.Printf(i18n.T("docker.remove_success")+"\n", containerName)
	return nil
}

// removeImage removes a Docker image
func (s *Shell) removeImage(imageName string, force bool) error {
	if !s.shellExecutor.IsDockerAvailable() {
		return fmt.Errorf(i18n.T("docker.not_available"))
	}

	if imageName == "" {
		return fmt.Errorf(i18n.T("docker.image_name_required"))
	}

	// Check if image exists
	cmd := exec.Command("docker", "images", "-q", imageName)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(i18n.T("docker.error_checking_image"), err)
	}

	if strings.TrimSpace(string(output)) == "" {
		return fmt.Errorf(i18n.T("docker.image_not_found"), imageName)
	}

	fmt.Printf(i18n.T("docker.remove_image")+"\n", imageName)
	
	var removeCmd *exec.Cmd
	if force {
		removeCmd = exec.Command("docker", "rmi", "-f", imageName)
	} else {
		removeCmd = exec.Command("docker", "rmi", imageName)
	}
	
	removeCmd.Stdout = s.getStdout()
	removeCmd.Stderr = s.getStderr()

	err = removeCmd.Run()
	if err != nil {
		return fmt.Errorf(i18n.T("docker.remove_image_failed"), imageName, err)
	}

	fmt.Printf(i18n.T("docker.remove_image_success")+"\n", imageName)
	return nil
}

// Docker補完関数群

// getDockerContainers は全てのDockerコンテナ（実行中・停止中）を取得します
func (s *Shell) getDockerContainers(running bool) []string {
	if !s.shellExecutor.IsDockerAvailable() {
		return []string{}
	}

	var cmd *exec.Cmd
	if running {
		// 実行中のコンテナのみ
		cmd = exec.Command("docker", "ps", "--format", "{{.Names}}")
	} else {
		// 全てのコンテナ（実行中・停止中）
		cmd = exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
	}

	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var containers []string
	for _, line := range lines {
		if line != "" {
			containers = append(containers, line)
		}
	}

	return containers
}

// getDockerImages はDockerイメージ一覧を取得します
func (s *Shell) getDockerImages() []string {
	if !s.shellExecutor.IsDockerAvailable() {
		return []string{}
	}

	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var images []string
	for _, line := range lines {
		if line != "" && !strings.Contains(line, "<none>") {
			images = append(images, line)
		}
	}

	return images
}

// getDockerNetworks はDockerネットワーク一覧を取得します
func (s *Shell) getDockerNetworks() []string {
	if !s.shellExecutor.IsDockerAvailable() {
		return []string{}
	}

	cmd := exec.Command("docker", "network", "ls", "--format", "{{.Name}}")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var networks []string
	for _, line := range lines {
		if line != "" {
			networks = append(networks, line)
		}
	}

	return networks
}

// getDockerVolumes はDockerボリューム一覧を取得します
func (s *Shell) getDockerVolumes() []string {
	if !s.shellExecutor.IsDockerAvailable() {
		return []string{}
	}

	cmd := exec.Command("docker", "volume", "ls", "--format", "{{.Name}}")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var volumes []string
	for _, line := range lines {
		if line != "" {
			volumes = append(volumes, line)
		}
	}

	return volumes
}