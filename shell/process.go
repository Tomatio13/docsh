package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func (s *Shell) executeExternalCommand(command string, args []string) error {
	var cmd *exec.Cmd
	
	// Windows環境での実行方法を調整
	if runtime.GOOS == "windows" {
		resolvedCommand := s.resolveWindowsCommand(command)
		if resolvedCommand != "" {
			cmd = exec.Command(resolvedCommand, args...)
		} else {
			// コマンドが見つからない場合はcmd.exe経由で実行
			fullArgs := append([]string{"/C", command}, args...)
			cmd = exec.Command("cmd.exe", fullArgs...)
		}
	} else {
		// Linux/Unix環境
		cmd = exec.Command(command, args...)
	}
	
	// 標準入出力を現在のプロセスと共有
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	// 作業ディレクトリを設定
	cmd.Dir = s.getCurrentDir()
	
	// 環境変数を継承
	cmd.Env = os.Environ()
	
	// コマンド実行
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// 終了コードが0以外の場合
			return fmt.Errorf("command failed with exit code %d", exitError.ExitCode())
		}
		return fmt.Errorf("failed to execute command '%s': %v", command, err)
	}
	
	return nil
}

func (s *Shell) isExecutableFile(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	
	// ファイルであることを確認
	if info.IsDir() {
		return false
	}
	
	// Windows環境では拡張子をチェック
	if runtime.GOOS == "windows" {
		executableExts := []string{".exe", ".com", ".bat", ".cmd"}
		for _, ext := range executableExts {
			if strings.HasSuffix(strings.ToLower(filename), ext) {
				return true
			}
		}
		return false
	}
	
	// Unix系では実行権限をチェック
	mode := info.Mode()
	return mode&0111 != 0
}

func (s *Shell) resolveWindowsCommand(command string) string {
	if runtime.GOOS != "windows" {
		return ""
	}
	
	// 既に拡張子がある場合はそのまま検索
	if s.hasWindowsExecutableExtension(command) {
		if fullPath, err := exec.LookPath(command); err == nil {
			return fullPath
		}
		return ""
	}
	
	// 実行可能な拡張子を順番に試行
	executableExts := []string{".exe", ".com", ".bat", ".cmd"}
	
	for _, ext := range executableExts {
		commandWithExt := command + ext
		if fullPath, err := exec.LookPath(commandWithExt); err == nil {
			return fullPath
		}
	}
	
	// 現在のディレクトリも検索
	currentDir := s.getCurrentDir()
	for _, ext := range executableExts {
		fullPath := filepath.Join(currentDir, command+ext)
		if s.isExecutableFile(fullPath) {
			return fullPath
		}
	}
	
	return ""
}

func (s *Shell) hasWindowsExecutableExtension(filename string) bool {
	if runtime.GOOS != "windows" {
		return false
	}
	
	executableExts := []string{".exe", ".com", ".bat", ".cmd", ".msi", ".ps1"}
	lowerFilename := strings.ToLower(filename)
	
	for _, ext := range executableExts {
		if strings.HasSuffix(lowerFilename, ext) {
			return true
		}
	}
	
	return false
}

func (s *Shell) getWindowsSystemDirectories() []string {
	systemDirs := []string{
		os.Getenv("SYSTEMROOT") + "\\System32",
		os.Getenv("SYSTEMROOT") + "\\SysWOW64", // 32bit互換
		os.Getenv("SYSTEMROOT"),
		"C:\\Windows\\System32",
		"C:\\Windows\\SysWOW64",
		"C:\\Windows",
	}
	
	// 空の要素を除去
	var validDirs []string
	for _, dir := range systemDirs {
		if dir != "" && dir != "\\System32" && dir != "\\SysWOW64" {
			validDirs = append(validDirs, dir)
		}
	}
	
	return validDirs
}