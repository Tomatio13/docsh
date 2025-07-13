package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func (s *Shell) changeDirectory(args []string) error {
	var target string
	
	if len(args) == 0 {
		// cdコマンドのみの場合はホームディレクトリへ
		if home, err := os.UserHomeDir(); err == nil {
			target = home
		} else {
			return fmt.Errorf("could not get home directory: %v", err)
		}
	} else {
		target = args[0]
		
		// Windows特有のパス処理
		if runtime.GOOS == "windows" {
			target = s.normalizeWindowsPath(target)
			target = s.expandWindowsEnvironmentVariable(target)
			
			// Windows特殊ディレクトリの処理
			if windowsPath := s.resolveWindowsSpecialPath(target); windowsPath != "" {
				target = windowsPath
			}
		}
	}
	
	// 相対パスを絶対パスに変換
	if !filepath.IsAbs(target) {
		current := s.getCurrentDir()
		target = filepath.Join(current, target)
	}
	
	// Windowsの場合はパスを正規化
	if runtime.GOOS == "windows" {
		target = filepath.Clean(target)
	}
	
	// ディレクトリの存在確認
	if info, err := os.Stat(target); err != nil {
		return fmt.Errorf("directory not found: %s", target)
	} else if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", target)
	}
	
	// ディレクトリ変更実行
	if err := os.Chdir(target); err != nil {
		return fmt.Errorf("failed to change directory: %v", err)
	}
	
	s.cwd = target
	return nil
}

func (s *Shell) resolveWindowsSpecialPath(path string) string {
	if runtime.GOOS != "windows" {
		return ""
	}
	
	lowerPath := strings.ToLower(path)
	
	// Windows特殊ディレクトリのマッピング
	specialDirs := s.getWindowsSpecialFolders()
	
	for name, fullPath := range specialDirs {
		if lowerPath == strings.ToLower(name) {
			return fullPath
		}
	}
	
	// ドライブレター処理
	if len(path) == 2 && path[1] == ':' {
		// "C:" -> "C:\"
		return path + "\\"
	}
	
	// ネットワークパス対応
	if strings.HasPrefix(path, "\\\\") {
		return path
	}
	
	return ""
}