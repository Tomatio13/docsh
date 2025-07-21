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
		// 引数がない場合はホームディレクトリに移動
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not get home directory: %v", err)
		}
		target = homeDir
	} else {
		target = args[0]

		// Dockerコンテナ名かどうかをチェック
		if s.shellExecutor.IsDockerAvailable() {
			isContainer, err := s.isDockerContainer(target)
			if err == nil && isContainer {
				// Dockerコンテナに入る
				fmt.Printf("Entering Docker container: %s\n", target)
				return s.enterContainer(target)
			}
		}

		// Windows特有のパス処理
		if runtime.GOOS == "windows" {
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

	// パスをクリーンアップ
	target = filepath.Clean(target)

	// ディレクトリの存在確認
	if info, err := os.Stat(target); err != nil {
		return fmt.Errorf("cd: %s: %v", target, err)
	} else if !info.IsDir() {
		return fmt.Errorf("cd: %s: not a directory", target)
	}

	// ディレクトリを変更
	if err := os.Chdir(target); err != nil {
		return fmt.Errorf("cd: %s: %v", target, err)
	}

	// 現在のディレクトリを更新
	s.cwd = target

	return nil
}

func (s *Shell) resolveWindowsSpecialPath(path string) string {
	if runtime.GOOS != "windows" {
		return ""
	}

	// Windows特殊ディレクトリのマッピング
	specialDirs := map[string]string{
		"~":         os.Getenv("USERPROFILE"),
		"desktop":   filepath.Join(os.Getenv("USERPROFILE"), "Desktop"),
		"docs":      filepath.Join(os.Getenv("USERPROFILE"), "Documents"),
		"downloads": filepath.Join(os.Getenv("USERPROFILE"), "Downloads"),
		"music":     filepath.Join(os.Getenv("USERPROFILE"), "Music"),
		"pictures":  filepath.Join(os.Getenv("USERPROFILE"), "Pictures"),
		"videos":    filepath.Join(os.Getenv("USERPROFILE"), "Videos"),
	}

	lowerPath := strings.ToLower(path)
	if specialPath, exists := specialDirs[lowerPath]; exists {
		return specialPath
	}

	return ""
}
