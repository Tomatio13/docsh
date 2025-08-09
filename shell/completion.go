package shell

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"docknaut/i18n"
)

// Suggest は簡易サジェスト型
type Suggest struct {
	Text        string
	Description string
}

// filterHasPrefix は前方一致でフィルタリングします（case-sensitive, leading match）
func filterHasPrefix(items []Suggest, prefix string, includeDescription bool) []Suggest {
	if prefix == "" {
		return items
	}
	var out []Suggest
	for _, it := range items {
		if strings.HasPrefix(it.Text, prefix) || (includeDescription && strings.HasPrefix(it.Description, prefix)) {
			out = append(out, it)
		}
	}
	return out
}

// Complete は現在の入力文字列に対する補完を返します
func (s *Shell) Complete(line string) []Suggest {
	beforeCursor := line
	words := strings.Fields(beforeCursor)

	if len(words) == 0 && beforeCursor == "" {
		return []Suggest{}
	}
	if strings.TrimSpace(beforeCursor) == "" {
		return []Suggest{}
	}
	if len(strings.TrimSpace(beforeCursor)) == 1 {
		return []Suggest{}
	}

	if len(words) == 0 {
		return s.completeCommands("")
	}

	if len(words) == 1 && !strings.HasSuffix(beforeCursor, " ") {
		// 1文字でもコマンド候補を早めに出す
		if words[0] == "project" {
			return s.completeProjectCommand([]string{"project"}, "", beforeCursor)
		} else if strings.HasPrefix("project", words[0]) {
			return s.completeProjectTopLevel(words[0])
		}
		return s.completeCommands(words[0])
	}

	command := words[0]
	var currentArg string
	if strings.HasSuffix(beforeCursor, " ") {
		currentArg = ""
	} else if len(words) > 1 {
		currentArg = words[len(words)-1]
	}

	switch command {
	case "cd":
		return s.completeDirectories(currentArg)
	case "login":
		return s.completeDockerContainers(currentArg, true)
	case "cat", "cp", "mv":
		return s.completeFiles(currentArg)
	case "rm":
		return s.completeDockerContainers(currentArg, false)
	case "rmi":
		return s.completeDockerImages(currentArg)
	case "start":
		return s.completeDockerContainers(currentArg, false)
	case "stop":
		return s.completeDockerContainers(currentArg, true)
	case "exec":
		return s.completeDockerContainers(currentArg, true)
	case "pull":
		return []Suggest{}
	case "ps":
		return []Suggest{}
	case "ls":
		return s.completeFilesAndDirectories(currentArg)
	case "kill":
		return s.completeDockerContainers(currentArg, true)
	case "tail", "head", "grep":
		return s.completeDockerContainers(currentArg, false)
	case "vi", "nano", "mkdir", "find", "locate":
		return s.completeDockerContainers(currentArg, true)
	case "netstat":
		return s.completeDockerContainers(currentArg, false)
	case "free", "top", "htop", "uname":
		return []Suggest{}
	case "df", "du":
		return []Suggest{}
	case "docker":
		return s.completeDockerCommand(words, currentArg, beforeCursor)
	case "project":
		return s.completeProjectCommand(words, currentArg, beforeCursor)
	case "theme":
		return s.completeThemes(currentArg)
	case "lang":
		// lang 単体でも言語候補を提示したいので currentArg をそのまま
		return s.completeLanguages(currentArg)
	default:
		return s.completeFilesAndDirectories(currentArg)
	}
}

// completeCommands はコマンド名の補完を提供します
func (s *Shell) completeCommands(prefix string) []Suggest {
	suggests := []Suggest{
		{Text: "cd", Description: i18n.T("completion.descriptions.cd")},
		{Text: "login", Description: i18n.T("completion.descriptions.login")},
		{Text: "pwd", Description: i18n.T("completion.descriptions.pwd")},
		{Text: "ls", Description: i18n.T("completion.descriptions.ls")},
		{Text: "cat", Description: i18n.T("completion.descriptions.cat")},
		{Text: "cp", Description: i18n.T("completion.descriptions.cp")},
		{Text: "mv", Description: i18n.T("completion.descriptions.mv")},
		{Text: "mkdir", Description: i18n.T("completion.descriptions.mkdir")},
		{Text: "rmdir", Description: i18n.T("completion.descriptions.rmdir")},
		{Text: "touch", Description: i18n.T("completion.descriptions.touch")},
		{Text: "echo", Description: i18n.T("completion.descriptions.echo")},
		{Text: "clear", Description: i18n.T("completion.descriptions.clear")},
		{Text: "exit", Description: i18n.T("completion.descriptions.exit")},
		{Text: "ps", Description: i18n.T("completion.descriptions.ps")},
		{Text: "kill", Description: i18n.T("completion.descriptions.kill")},
		{Text: "rm", Description: i18n.T("completion.descriptions.rm")},
		{Text: "rmi", Description: i18n.T("completion.descriptions.rmi")},
		{Text: "start", Description: i18n.T("completion.descriptions.start")},
		{Text: "stop", Description: i18n.T("completion.descriptions.stop")},
		{Text: "exec", Description: i18n.T("completion.descriptions.exec")},
		{Text: "pull", Description: i18n.T("completion.descriptions.pull")},
		{Text: "tail", Description: i18n.T("completion.descriptions.tail")},
		{Text: "head", Description: i18n.T("completion.descriptions.head")},
		{Text: "grep", Description: i18n.T("completion.descriptions.grep")},
		{Text: "vi", Description: i18n.T("completion.descriptions.vi")},
		{Text: "nano", Description: i18n.T("completion.descriptions.nano")},
		{Text: "find", Description: i18n.T("completion.descriptions.find")},
		{Text: "locate", Description: i18n.T("completion.descriptions.locate")},
		{Text: "netstat", Description: i18n.T("completion.descriptions.netstat")},
		{Text: "free", Description: i18n.T("completion.descriptions.free")},
		{Text: "top", Description: i18n.T("completion.descriptions.top")},
		{Text: "htop", Description: i18n.T("completion.descriptions.htop")},
		{Text: "df", Description: i18n.T("completion.descriptions.df")},
		{Text: "du", Description: i18n.T("completion.descriptions.du")},
		{Text: "uname", Description: i18n.T("completion.descriptions.uname")},
		{Text: "docker", Description: i18n.T("completion.descriptions.docker")},
		{Text: "theme", Description: i18n.T("completion.descriptions.theme")},
		{Text: "lang", Description: i18n.T("completion.descriptions.lang")},
		{Text: "alias", Description: i18n.T("completion.descriptions.alias")},
		{Text: "config", Description: i18n.T("completion.descriptions.config")},
		{Text: "project", Description: "Docker Compose プロジェクト操作"},
	}

	if s.config != nil {
		for alias := range s.config.Aliases {
			suggests = append(suggests, Suggest{
				Text:        alias,
				Description: i18n.T("completion.alias_value", s.config.Aliases[alias]),
			})
		}
	}

	return filterHasPrefix(suggests, prefix, true)
}

// completeFiles はファイル名の補完を提供します
func (s *Shell) completeFiles(prefix string) []Suggest {
	return s.completeFileSystem(prefix, false, true)
}

// completeDirectories はディレクトリ名の補完を提供します
func (s *Shell) completeDirectories(prefix string) []Suggest {
	return s.completeFileSystem(prefix, true, false)
}

// completeFilesAndDirectories はファイルとディレクトリの補完を提供します
func (s *Shell) completeFilesAndDirectories(prefix string) []Suggest {
	return s.completeFileSystem(prefix, true, true)
}

// completeFileSystem はファイルシステムの補完を提供します
func (s *Shell) completeFileSystem(prefix string, includeDirs, includeFiles bool) []Suggest {
	var suggests []Suggest

	dir := filepath.Dir(prefix)
	base := filepath.Base(prefix)
	if !filepath.IsAbs(dir) {
		if dir == "." || prefix == "" || !strings.Contains(prefix, string(filepath.Separator)) {
			dir = s.getCurrentDir()
			if prefix != "" && !strings.Contains(prefix, string(filepath.Separator)) {
				base = prefix
			} else {
				base = ""
			}
		} else {
			dir = filepath.Join(s.getCurrentDir(), dir)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return suggests
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(base, ".") {
			continue
		}
		if !strings.HasPrefix(name, base) {
			continue
		}
		var fullPath string
		if strings.Contains(prefix, string(filepath.Separator)) {
			fullPath = filepath.Join(filepath.Dir(prefix), name)
		} else {
			fullPath = name
		}

		if entry.IsDir() {
			if includeDirs {
				suggests = append(suggests, Suggest{Text: fullPath + string(filepath.Separator), Description: i18n.T("completion.entry_directory")})
			}
		} else if includeFiles {
			suggests = append(suggests, Suggest{Text: fullPath, Description: i18n.T("completion.entry_file")})
		}
	}

	return suggests
}

// completeThemes はテーマの補完を提供します
func (s *Shell) completeThemes(prefix string) []Suggest {
	suggests := []Suggest{
		{Text: "default", Description: "デフォルトテーマ"},
		{Text: "minimal", Description: "ミニマルテーマ"},
		{Text: "robbyrussell", Description: "Robby Russell テーマ"},
		{Text: "agnoster", Description: "Agnoster テーマ"},
		{Text: "pure", Description: "Pure テーマ"},
	}
	return filterHasPrefix(suggests, prefix, true)
}

// completeLanguages は言語の補完を提供します
func (s *Shell) completeLanguages(prefix string) []Suggest {
	suggests := []Suggest{{Text: "en", Description: "English"}, {Text: "ja", Description: "日本語"}}
	return filterHasPrefix(suggests, prefix, true)
}

// Docker補完関数群
func (s *Shell) completeDockerContainers(prefix string, running bool) []Suggest {
	containers := s.getDockerContainers(running)
	if len(containers) == 0 {
		return []Suggest{}
	}
	var suggests []Suggest
	for _, container := range containers {
		var description string
		if running {
			description = "実行中のコンテナ"
		} else {
			description = "コンテナ"
		}
		suggests = append(suggests, Suggest{Text: container, Description: description})
	}
	return filterHasPrefix(suggests, prefix, true)
}

func (s *Shell) completeDockerImages(prefix string) []Suggest {
	images := s.getDockerImages()
	if len(images) == 0 {
		return []Suggest{}
	}
	var suggests []Suggest
	for _, image := range images {
		suggests = append(suggests, Suggest{Text: image, Description: "Dockerイメージ"})
	}
	return filterHasPrefix(suggests, prefix, true)
}

func (s *Shell) completeDockerNetworks(prefix string) []Suggest {
	networks := s.getDockerNetworks()
	if len(networks) == 0 {
		return []Suggest{}
	}
	var suggests []Suggest
	for _, network := range networks {
		suggests = append(suggests, Suggest{Text: network, Description: "Dockerネットワーク"})
	}
	return filterHasPrefix(suggests, prefix, true)
}

// completeProjectCommand は `project` の補完を提供します
func (s *Shell) completeProjectCommand(words []string, currentArg, beforeCursor string) []Suggest {
	// project | project <name> | project <name> <sub>
	if len(words) == 1 { // after typing 'project' and a space?
		// 提案: ps と、現在検出できるプロジェクト名
		suggests := []Suggest{{Text: "ps", Description: "全プロジェクト一覧"}}
		// 検出してプロジェクト名候補
		projects := s.detectComposeProjects()
		for _, p := range projects {
			suggests = append(suggests, Suggest{Text: p, Description: "プロジェクト"})
		}
		return filterHasPrefix(suggests, currentArg, true)
	}

	if len(words) == 2 { // project <name>
		// サブコマンド候補
		suggests := []Suggest{
			{Text: "ps", Description: "プロジェクトのサービス一覧"},
			{Text: "logs", Description: "サービスのログ"},
			{Text: "start", Description: "サービス/プロジェクト開始"},
			{Text: "restart", Description: "再起動"},
			{Text: "stop", Description: "停止"},
		}
		return filterHasPrefix(suggests, currentArg, true)
	}

	if len(words) >= 3 { // project <name> <sub>
		sub := words[2]
		// サービス名補完（ps 以外の時）
		if sub == "logs" || sub == "start" || sub == "restart" || sub == "stop" {
			services := s.detectComposeServices(words[1])
			var suggests []Suggest
			for _, sv := range services {
				suggests = append(suggests, Suggest{Text: sv, Description: "サービス"})
			}
			return filterHasPrefix(suggests, currentArg, true)
		}
	}
	return []Suggest{}
}

// detectComposeProjects は稼働中/停止中コンテナの compose プロジェクト名を列挙
func (s *Shell) detectComposeProjects() []string {
	// シンプルに `docker ps -a --format` でラベル抽出
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Label \"com.docker.compose.project\"}}")
	out, err := cmd.Output()
	if err != nil {
		return []string{}
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	m := map[string]bool{}
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || l == "<no value>" {
			continue
		}
		m[l] = true
	}
	var projects []string
	for p := range m {
		projects = append(projects, p)
	}
	return projects
}

// detectComposeServices は指定プロジェクトのサービス名を列挙
func (s *Shell) detectComposeServices(project string) []string {
	// docker ps -a から該当プロジェクトの service ラベルを集める
	format := "{{.Label \"com.docker.compose.project\"}}\t{{.Label \"com.docker.compose.service\"}}"
	cmd := exec.Command("docker", "ps", "-a", "--format", format)
	out, err := cmd.Output()
	if err != nil {
		return []string{}
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	m := map[string]bool{}
	for _, l := range lines {
		parts := strings.Split(l, "\t")
		if len(parts) != 2 {
			continue
		}
		if strings.TrimSpace(parts[0]) == project {
			sv := strings.TrimSpace(parts[1])
			if sv != "" && sv != "<no value>" {
				m[sv] = true
			}
		}
	}
	var services []string
	for sv := range m {
		services = append(services, sv)
	}
	return services
}

func (s *Shell) completeDockerVolumes(prefix string) []Suggest {
	volumes := s.getDockerVolumes()
	if len(volumes) == 0 {
		return []Suggest{}
	}
	var suggests []Suggest
	for _, volume := range volumes {
		suggests = append(suggests, Suggest{Text: volume, Description: "Dockerボリューム"})
	}
	return filterHasPrefix(suggests, prefix, true)
}

func (s *Shell) completeDockerSubcommands(prefix string) []Suggest {
	suggests := []Suggest{
		{Text: "ps", Description: i18n.T("completion.docker_subcommands.ps")},
		{Text: "images", Description: i18n.T("completion.docker_subcommands.images")},
		{Text: "run", Description: i18n.T("completion.docker_subcommands.run")},
		{Text: "exec", Description: i18n.T("completion.docker_subcommands.exec")},
		{Text: "start", Description: i18n.T("completion.docker_subcommands.start")},
		{Text: "stop", Description: i18n.T("completion.docker_subcommands.stop")},
		{Text: "restart", Description: i18n.T("completion.docker_subcommands.restart")},
		{Text: "rm", Description: i18n.T("completion.docker_subcommands.rm")},
		{Text: "rmi", Description: i18n.T("completion.docker_subcommands.rmi")},
		{Text: "pull", Description: i18n.T("completion.docker_subcommands.pull")},
		{Text: "push", Description: i18n.T("completion.docker_subcommands.push")},
		{Text: "build", Description: i18n.T("completion.docker_subcommands.build")},
		{Text: "logs", Description: i18n.T("completion.docker_subcommands.logs")},
		{Text: "inspect", Description: i18n.T("completion.docker_subcommands.inspect")},
		{Text: "network", Description: i18n.T("completion.docker_subcommands.network")},
		{Text: "volume", Description: i18n.T("completion.docker_subcommands.volume")},
		{Text: "system", Description: i18n.T("completion.docker_subcommands.system")},
		{Text: "version", Description: i18n.T("completion.docker_subcommands.version")},
	}
	return filterHasPrefix(suggests, prefix, true)
}

func (s *Shell) completeDockerCommand(words []string, currentArg, beforeCursor string) []Suggest {
	if len(words) == 2 && !strings.HasSuffix(beforeCursor, " ") {
		return s.completeDockerSubcommands(currentArg)
	}
	if len(words) < 2 {
		return []Suggest{}
	}
	subcommand := words[1]
	switch subcommand {
	case "rm":
		if len(words) >= 3 && (len(words) > 3 || strings.HasSuffix(beforeCursor, " ")) {
			return s.completeDockerContainers(currentArg, false)
		}
		return s.completeDockerContainers(currentArg, false)
	case "rmi":
		return s.completeDockerImages(currentArg)
	case "stop", "restart", "exec", "logs", "inspect":
		return s.completeDockerContainers(currentArg, true)
	case "start":
		return s.completeDockerContainers(currentArg, false)
	case "run", "push":
		return s.completeDockerImages(currentArg)
	case "pull":
		return []Suggest{}
	case "network":
		if len(words) == 3 && !strings.HasSuffix(beforeCursor, " ") {
			suggests := []Suggest{
				{Text: "ls", Description: i18n.T("completion.docker_network_subcommands.ls")},
				{Text: "create", Description: i18n.T("completion.docker_network_subcommands.create")},
				{Text: "rm", Description: i18n.T("completion.docker_network_subcommands.rm")},
				{Text: "inspect", Description: i18n.T("completion.docker_network_subcommands.inspect")},
			}
			return filterHasPrefix(suggests, currentArg, true)
		}
		if len(words) >= 4 && (words[2] == "rm" || words[2] == "inspect") {
			return s.completeDockerNetworks(currentArg)
		}
		return []Suggest{}
	case "volume":
		if len(words) == 3 && !strings.HasSuffix(beforeCursor, " ") {
			suggests := []Suggest{
				{Text: "ls", Description: i18n.T("completion.docker_volume_subcommands.ls")},
				{Text: "create", Description: i18n.T("completion.docker_volume_subcommands.create")},
				{Text: "rm", Description: i18n.T("completion.docker_volume_subcommands.rm")},
				{Text: "inspect", Description: i18n.T("completion.docker_volume_subcommands.inspect")},
			}
			return filterHasPrefix(suggests, currentArg, true)
		}
		if len(words) >= 4 && (words[2] == "rm" || words[2] == "inspect") {
			return s.completeDockerVolumes(currentArg)
		}
		return []Suggest{}
	case "build":
		return s.completeDirectories(currentArg)
	default:
		return s.completeFilesAndDirectories(currentArg)
	}
}

// completeProjectTopLevel は "projec" 等の入力中にも project を優先的に提案する
func (s *Shell) completeProjectTopLevel(prefix string) []Suggest {
	suggests := []Suggest{{Text: "project", Description: "Docker Compose プロジェクト操作"}}
	return filterHasPrefix(suggests, prefix, true)
}
