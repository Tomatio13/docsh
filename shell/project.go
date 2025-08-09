package shell

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// composeLabels used by Docker Compose
const (
	composeProjectLabel    = "com.docker.compose.project"
	composeWorkingDirLabel = "com.docker.compose.project.working_dir"
	composeServiceLabel    = "com.docker.compose.service"
)

type containerInfo struct {
	ID         string
	Names      string
	Status     string
	Ports      string
	Project    string
	WorkingDir string
	Service    string
}

type projectService struct {
	ServiceName string
	Container   containerInfo
}

type projectGroup struct {
	ProjectName string
	WorkingDir  string
	Services    []projectService
}

// handleProjectCommand implements: project <name> [ps|logs <svc>|restart <svc>|stop]
func (s *Shell) handleProjectCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("project name required")
	}
	// support: project ps  -> list all projects
	if args[0] == "ps" && len(args) == 1 {
		return s.psByProject()
	}

	project := args[0]
	action := "ps"
	rest := []string{}
	if len(args) > 1 {
		action = args[1]
		rest = args[2:]
	}

	// collect containers belonging to projects
	groups, err := s.collectProjects()
	if err != nil {
		return err
	}
	var pg *projectGroup
	for i := range groups {
		if groups[i].ProjectName == project {
			pg = &groups[i]
			break
		}
	}

	switch action {
	case "ps":
		printProjectPS(pg)
		return nil
	case "logs":
		// 2系統の入力を許容する:
		// 1) 正規: project <project> logs <service> [options]
		// 2) 省略: project <service> logs [options]
		if len(rest) == 0 || strings.HasPrefix(rest[0], "-") {
			// 省略系: 第一引数はサービス名、rest はオプション
			service := project
			options := rest
			// サービス名から一意のコンテナを特定
			var (
				foundContainer string
				foundProjects  []string
			)
			for i := range groups {
				g := &groups[i]
				for _, svc := range g.Services {
					if svc.ServiceName == service || svc.Container.Names == service {
						foundProjects = append(foundProjects, g.ProjectName)
						foundContainer = svc.Container.Names
					}
				}
			}
			if len(foundProjects) == 0 {
				return fmt.Errorf("service not found: %s", service)
			}
			if len(foundProjects) > 1 {
				return fmt.Errorf("service '%s' is ambiguous across projects: %s", service, strings.Join(foundProjects, ", "))
			}
			return s.execDocker("logs", append(options, foundContainer)...)
		}
		// 正規系: pg が必要
		if pg == nil {
			return fmt.Errorf("project not found: %s", project)
		}
		// docker logs は [OPTIONS] CONTAINER 順
		container := serviceContainerName(pg, rest[0])
		options := rest[1:]
		return s.execDocker("logs", append(options, container)...)
	case "start":
		composeFile := filepath.Join(pg.WorkingDir, "docker-compose.yml")
		if len(rest) == 0 {
			// start whole project (prefer docker compose)
			if fileExists(composeFile) {
				return s.execDocker("compose", "-f", composeFile, "up", "-d")
			}
			for _, svc := range pg.Services {
				_ = s.execDocker("start", svc.Container.Names)
			}
			return nil
		}
		// single service
		if fileExists(composeFile) {
			return s.execDocker("compose", "-f", composeFile, "start", rest[0])
		}
		return s.execDocker("start", serviceContainerName(pg, rest[0]))
	case "restart":
		composeFile := filepath.Join(pg.WorkingDir, "docker-compose.yml")
		if len(rest) == 0 {
			if fileExists(composeFile) {
				return s.execDocker("compose", "-f", composeFile, "restart")
			}
			for _, svc := range pg.Services {
				_ = s.execDocker("restart", svc.Container.Names)
			}
			return nil
		}
		if fileExists(composeFile) {
			return s.execDocker("compose", "-f", composeFile, "restart", rest[0])
		}
		return s.execDocker("restart", serviceContainerName(pg, rest[0]))
	case "stop":
		composeFile := filepath.Join(pg.WorkingDir, "docker-compose.yml")
		if len(rest) == 0 {
			if fileExists(composeFile) {
				return s.execDocker("compose", "-f", composeFile, "stop")
			}
			for _, svc := range pg.Services {
				_ = s.execDocker("stop", svc.Container.Names)
			}
			return nil
		}
		if fileExists(composeFile) {
			return s.execDocker("compose", "-f", composeFile, "stop", rest[0])
		}
		return s.execDocker("stop", serviceContainerName(pg, rest[0]))
	default:
		return fmt.Errorf("unknown project action: %s", action)
	}
}

func serviceContainerName(pg *projectGroup, service string) string {
	for _, svc := range pg.Services {
		if svc.ServiceName == service {
			return svc.Container.Names
		}
	}
	return service
}

func printProjectPS(pg *projectGroup) {
	fmt.Printf("Project: %s (%s)\n", pg.ProjectName, pg.WorkingDir)
	fmt.Printf("Services:\n")
	for _, svc := range pg.Services {
		status := strings.TrimSpace(svc.Container.Status)
		ports := strings.TrimSpace(svc.Container.Ports)
		if ports != "" && ports != "-" {
			fmt.Printf("  %-10s %-8s %s\n", svc.ServiceName, status, ports)
		} else {
			fmt.Printf("  %-10s %s\n", svc.ServiceName, status)
		}
	}
	// Ensure prompt does not overwrite the last service line
	fmt.Println()
}

// psByProject implements: ps --by-project
func (s *Shell) psByProject() error {
	groups, err := s.collectProjects()
	if err != nil {
		return err
	}
	if len(groups) == 0 {
		// fallback: just run docker ps -a formatting handled upstream
		return s.execDocker("ps", "-a")
	}
	if os.Getenv("DOCSH_DEBUG") == "1" {
		fmt.Println("[DEBUG] project groups built:")
		for _, g := range groups {
			fmt.Printf("  [DEBUG] %s (%s) services=%d\n", g.ProjectName, g.WorkingDir, len(g.Services))
			for _, svc := range g.Services {
				fmt.Printf("    [DEBUG] svc=%s name=%s status=%s\n", svc.ServiceName, svc.Container.Names, svc.Container.Status)
			}
		}
	}

	for i := range groups {
		pg := &groups[i]
		fmt.Printf("\n📦 %s (%s)\n", pg.ProjectName, pg.WorkingDir)
		for _, svc := range pg.Services {
			status := strings.TrimSpace(svc.Container.Status)
			ports := strings.TrimSpace(svc.Container.Ports)
			if ports != "" && ports != "-" {
				fmt.Printf("  %-10s %-8s %s\n", svc.ServiceName, status, ports)
			} else {
				fmt.Printf("  %-10s %s\n", svc.ServiceName, status)
			}
		}
	}
	// 末尾に空行を出して、REPL のプロンプトが直前行を上書きしないようにする
	fmt.Println()
	return nil
}

// collectProjects scans docker ps -a and groups by compose project label
func (s *Shell) collectProjects() ([]projectGroup, error) {
	if !s.shellExecutor.IsDockerAvailable() {
		return nil, fmt.Errorf("Docker is not available")
	}
	// First list all container IDs
	idsOut, err := exec.Command("docker", "ps", "-a", "-q").Output()
	if err != nil {
		return nil, err
	}
	ids := strings.Fields(strings.TrimSpace(string(idsOut)))
	var all []containerInfo
	for _, id := range ids {
		// Inspect each container to get reliable fields
		// Ports as JSON to reconstruct mapping
		// Use custom delimiter to avoid accidental splitting by tabs inside values
		delim := "::DOCSH::"
		format := "{{.Id}}" + delim + "{{.Name}}" + delim + "{{.State.Status}}" + delim + "{{json .NetworkSettings.Ports}}" + delim + "{{index .Config.Labels \"com.docker.compose.project\"}}" + delim + "{{index .Config.Labels \"com.docker.compose.project.working_dir\"}}" + delim + "{{index .Config.Labels \"com.docker.compose.service\"}}"
		out, err := exec.Command("docker", "inspect", "-f", format, id).Output()
		if err != nil {
			continue
		}
		fields := strings.Split(strings.TrimSpace(string(out)), delim)
		if len(fields) < 4 {
			continue
		}
		name := strings.TrimPrefix(fields[1], "/")
		ci := containerInfo{
			ID:     fields[0],
			Names:  name,
			Status: fields[2],
			Ports:  formatPortsFromJSON(fields[3]),
		}
		if len(fields) >= 5 {
			ci.Project = normalizeVal(fields[4])
		}
		if len(fields) >= 6 {
			ci.WorkingDir = normalizeVal(fields[5])
		}
		if len(fields) >= 7 {
			ci.Service = normalizeVal(fields[6])
		}
		// 補助: service が空で、Names が "<project>-<service>-N" または "<service>" 形式なら補完
		if ci.Service == "" && ci.Project != "" {
			n := ci.Names
			base := n
			if strings.HasPrefix(n, ci.Project+"-") {
				// hyphen 区切り: project-service-index
				parts := strings.Split(n, "-")
				if len(parts) >= 2 {
					base = parts[1]
				}
			} else if strings.HasPrefix(n, ci.Project+"_") {
				// underscore 区切り: project_service_index
				parts := strings.Split(n, "_")
				if len(parts) >= 2 {
					base = parts[1]
				}
			}
			// index サフィックス除去
			base = strings.TrimSuffix(base, "-1")
			base = strings.TrimSuffix(base, "_1")
			if base != "" && base != ci.Project {
				ci.Service = base
			}
		}

		if os.Getenv("DOCSH_DEBUG") == "1" {
			fmt.Printf("[DEBUG] id=%s name=%s project=%s service=%s workdir=%s\n", ci.ID[:12], ci.Names, ci.Project, ci.Service, ci.WorkingDir)
		}

		all = append(all, ci)
	}

	// (no longer used; we extract fields directly)

	// build map
	m := map[string]*projectGroup{}
	for _, c := range all {
		p := strings.TrimSpace(c.Project)
		if p == "" {
			continue
		}
		if _, exists := m[p]; !exists {
			m[p] = &projectGroup{ProjectName: p, WorkingDir: strings.TrimSpace(c.WorkingDir)}
		}
		svc := strings.TrimSpace(c.Service)
		if svc == "" {
			svc = c.Names
		}
		// 重複排除（同名サービスが複数回追加されないように）
		exists := false
		for _, s := range m[p].Services {
			if s.ServiceName == svc {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		m[p].Services = append(m[p].Services, projectService{ServiceName: svc, Container: c})
	}

	// sort services and collect
	var groups []projectGroup
	for _, v := range m {
		sort.Slice(v.Services, func(i, j int) bool { return v.Services[i].ServiceName < v.Services[j].ServiceName })
		groups = append(groups, *v)
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].ProjectName < groups[j].ProjectName })
	return groups, nil
}

// fetchComposeLabels gets compose labels via docker inspect for a container id
func fetchComposeLabels(containerID string) (project string, workingDir string, service string) {
	// Use Go template to print exactly the three labels as tsv
	format := "{{index .Config.Labels \"com.docker.compose.project\"}}\t{{index .Config.Labels \"com.docker.compose.project.working_dir\"}}\t{{index .Config.Labels \"com.docker.compose.service\"}}"
	cmd := exec.Command("docker", "inspect", "-f", format, containerID)
	out, err := cmd.Output()
	if err != nil {
		return "", "", ""
	}
	fields := strings.Split(strings.TrimSpace(string(out)), "\t")
	if len(fields) >= 1 {
		project = fields[0]
	}
	if len(fields) >= 2 {
		workingDir = fields[1]
	}
	if len(fields) >= 3 {
		service = fields[2]
	}
	return strings.TrimSpace(project), strings.TrimSpace(workingDir), strings.TrimSpace(service)
}

// formatPortsFromJSON converts docker inspect Ports JSON to human-friendly mapping
func formatPortsFromJSON(portsJSON string) string {
	if strings.TrimSpace(portsJSON) == "null" || strings.TrimSpace(portsJSON) == "" {
		return ""
	}
	type binding struct {
		HostIp   string `json:"HostIp"`
		HostPort string `json:"HostPort"`
	}
	var data map[string][]binding
	if err := json.Unmarshal([]byte(portsJSON), &data); err != nil {
		return ""
	}
	var parts []string
	for containerPort, binds := range data {
		if len(binds) == 0 {
			parts = append(parts, containerPort)
			continue
		}
		for _, b := range binds {
			host := b.HostIp
			if host == "" {
				host = "0.0.0.0"
			}
			parts = append(parts, fmt.Sprintf("%s:%s->%s", host, b.HostPort, containerPort))
		}
	}
	return strings.Join(parts, ", ")
}

// normalizeVal converts docker template "<no value>" to empty string and trims spaces
func normalizeVal(s string) string {
	s = strings.TrimSpace(s)
	if s == "<no value>" {
		return ""
	}
	return s
}

// execDocker is a tiny helper to run docker subcommands directly with passthrough IO
func (s *Shell) execDocker(subcmd string, args ...string) error {
	if !s.shellExecutor.IsDockerAvailable() {
		return fmt.Errorf("Docker is not available")
	}
	full := append([]string{subcmd}, args...)
	cmd := exec.Command("docker", full...)
	cmd.Stdout = s.getStdout()
	cmd.Stderr = s.getStderr()
	cmd.Stdin = s.getStdin()
	return cmd.Run()
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
