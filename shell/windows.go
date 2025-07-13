package shell

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func (s *Shell) isBuiltinCommand(command string) bool {
	builtinCommands := map[string]bool{
		"ls":     true,
		"cat":    true,
		"clear":  true,
		"cp":     true,
		"mv":     true,
		"rm":     true,
		"mkdir":  true,
		"rmdir":  true,
		"echo":   true,
		"env":    true,
		"which":  true,
	}
	
	return builtinCommands[strings.ToLower(command)]
}

func (s *Shell) executeBuiltinCommand(command string, args []string) error {
	switch strings.ToLower(command) {
	case "ls":
		return s.unixLs(args)
	case "cat":
		return s.unixCat(args)
	case "clear":
		return s.unixClear()
	case "cp":
		return s.unixCp(args)
	case "mv":
		return s.unixMv(args)
	case "rm":
		return s.unixRm(args)
	case "mkdir":
		return s.unixMkdir(args)
	case "rmdir":
		return s.unixRmdir(args)
	case "echo":
		return s.unixEcho(args)
	case "env":
		return s.unixEnv(args)
	case "which":
		return s.unixWhich(args)
	default:
		return fmt.Errorf("unknown builtin command: %s", command)
	}
}

func (s *Shell) unixLs(args []string) error {
	path := "."
	showLong := false
	showAll := false
	
	// オプション解析
	var paths []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "l") {
				showLong = true
			}
			if strings.Contains(arg, "a") {
				showAll = true
			}
		} else {
			paths = append(paths, arg)
		}
	}
	
	if len(paths) > 0 {
		path = paths[0]
	}
	
	// Windowsパス形式を正規化
	if runtime.GOOS == "windows" {
		path = s.normalizeWindowsPath(path)
	}
	
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("ls: cannot access '%s': %v", path, err)
	}
	
	// 隠しファイルのフィルタリング
	if !showAll {
		var visibleFiles []os.FileInfo
		for _, file := range files {
			if !strings.HasPrefix(file.Name(), ".") {
				visibleFiles = append(visibleFiles, file)
			}
		}
		files = visibleFiles
	}
	
	// ソート（名前順）
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name())
	})
	
	if showLong {
		// ls -l形式
		for _, file := range files {
			var perm string
			if file.IsDir() {
				perm = "d"
			} else {
				perm = "-"
			}
			
			// 簡易的な権限表示
			perm += "rwxrwxrwx" // Unix風の権限表示（簡略化）
			
			fmt.Printf("%s %8d %s %s\n",
				perm,
				file.Size(),
				file.ModTime().Format("Jan 02 15:04"),
				file.Name())
		}
	} else {
		// 通常のls形式
		var names []string
		for _, file := range files {
			name := file.Name()
			if file.IsDir() {
				name += "/"
			}
			names = append(names, name)
		}
		
		// 複数列で表示
		cols := 4
		for i, name := range names {
			fmt.Printf("%-20s", name)
			if (i+1)%cols == 0 {
				fmt.Println()
			}
		}
		if len(names)%cols != 0 {
			fmt.Println()
		}
	}
	
	// 出力完了
	
	return nil
}

func (s *Shell) unixCat(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("cat: missing file argument")
	}
	
	for _, filename := range args {
		filename = s.normalizeWindowsPath(filename)
		
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cat: cannot read '%s': %v\n", filename, err)
			continue
		}
		
		// Windows形式の改行に対応
		contentStr := string(content)
		contentStr = strings.ReplaceAll(contentStr, "\r\n", "\n")
		fmt.Print(contentStr)
		
		if len(args) > 1 {
			fmt.Println() // 複数ファイルの場合は改行を追加
		}
	}
	
	return nil
}

func (s *Shell) unixClear() error {
	// ANSI エスケープシーケンスでクリア
	fmt.Print("\033[2J\033[H")
	return nil
}

func (s *Shell) unixCp(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("cp: missing source or destination")
	}
	
	src := s.normalizeWindowsPath(args[0])
	dst := s.normalizeWindowsPath(args[1])
	
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Errorf("cp: cannot read '%s': %v", src, err)
	}
	
	if err := ioutil.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("cp: cannot write '%s': %v", dst, err)
	}
	
	fmt.Printf("        1 file(s) copied.\n")
	return nil
}

func (s *Shell) unixMv(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("mv: missing source or destination")
	}
	
	src := s.normalizeWindowsPath(args[0])
	dst := s.normalizeWindowsPath(args[1])
	
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("mv: cannot move '%s' to '%s': %v", src, dst, err)
	}
	
	fmt.Printf("        1 file(s) moved.\n")
	return nil
}

func (s *Shell) unixRm(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("rm: missing file argument")
	}
	
	deletedCount := 0
	for _, filename := range args {
		filename = s.normalizeWindowsPath(filename)
		
		if err := os.Remove(filename); err != nil {
			fmt.Fprintf(os.Stderr, "rm: cannot delete '%s': %v\n", filename, err)
			continue
		}
		deletedCount++
	}
	
	if deletedCount > 0 {
		fmt.Printf("        %d file(s) deleted.\n", deletedCount)
	}
	
	return nil
}

func (s *Shell) unixMkdir(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("mkdir: missing directory argument")
	}
	
	for _, dirname := range args {
		dirname = s.normalizeWindowsPath(dirname)
		
		if err := os.MkdirAll(dirname, 0755); err != nil {
			return fmt.Errorf("mkdir: cannot create directory '%s': %v", dirname, err)
		}
	}
	
	return nil
}

func (s *Shell) unixRmdir(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("rmdir: missing directory argument")
	}
	
	for _, dirname := range args {
		dirname = s.normalizeWindowsPath(dirname)
		
		if err := os.Remove(dirname); err != nil {
			return fmt.Errorf("rmdir: cannot remove directory '%s': %v", dirname, err)
		}
	}
	
	return nil
}

func (s *Shell) unixEcho(args []string) error {
	// Unix風のecho（改行あり）
	fmt.Println(strings.Join(args, " "))
	return nil
}

func (s *Shell) unixEnv(args []string) error {
	if len(args) == 0 {
		// 全環境変数を表示（ソート済み）
		envVars := os.Environ()
		sort.Strings(envVars)
		for _, env := range envVars {
			fmt.Println(env)
		}
		return nil
	}
	
	// Unix風のenv: 単一変数の表示は行わない
	// 引数がある場合は設定として扱う
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			os.Setenv(parts[0], parts[1])
		} else {
			return fmt.Errorf("env: invalid argument '%s'", arg)
		}
	}
	
	return nil
}

func (s *Shell) unixWhich(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("which: missing command argument")
	}
	
	command := args[0]
	
	// PATH環境変数から検索
	pathEnv := os.Getenv("PATH")
	var pathSeparator string
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	} else {
		pathSeparator = ":"
	}
	paths := strings.Split(pathEnv, pathSeparator)
	
	found := false
	for _, path := range paths {
		if path == "" {
			continue
		}
		
		var extensions []string
		if runtime.GOOS == "windows" {
			extensions = []string{"", ".exe", ".com", ".bat", ".cmd"}
		} else {
			extensions = []string{""}
		}
		
		for _, ext := range extensions {
			fullPath := filepath.Join(path, command+ext)
			if _, err := os.Stat(fullPath); err == nil {
				fmt.Println(fullPath)
				found = true
				break // Unixのwhichは最初の一つだけ表示
			}
		}
		if found {
			break
		}
	}
	
	if !found {
		return fmt.Errorf("%s not found", command)
	}
	
	return nil
}

func (s *Shell) normalizeWindowsPath(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}
	
	// バックスラッシュをスラッシュに統一（Goが自動変換）
	path = strings.ReplaceAll(path, "\\", "/")
	
	return path
}