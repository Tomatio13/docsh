package shell

import (
	"os"
	"runtime"
	"strings"
)

func (s *Shell) initializeWindowsEnvironment() {
	if runtime.GOOS != "windows" {
		return
	}

	// Windows特有の環境変数を設定
	s.setDefaultWindowsEnvironmentVariables()

	// PATHにWindows固有のディレクトリを追加
	s.enhanceWindowsPath()
}

func (s *Shell) setDefaultWindowsEnvironmentVariables() {
	// Windows版として識別
	os.Setenv("GO_ZSH_OS", "windows")
	os.Setenv("GO_ZSH_ARCH", runtime.GOARCH)

	// Windowsの重要な環境変数が設定されていない場合のデフォルト値
	defaults := map[string]string{
		"PATHEXT":    ".COM;.EXE;.BAT;.CMD;.VBS;.VBE;.JS;.JSE;.WSF;.WSH;.MSC",
		"TEMP":       "C:\\Windows\\Temp",
		"TMP":        "C:\\Windows\\Temp",
		"SYSTEMROOT": "C:\\Windows",
		"WINDIR":     "C:\\Windows",
		"COMSPEC":    "C:\\Windows\\System32\\cmd.exe",
		"PROMPT":     "$P$G",
	}

	for key, defaultValue := range defaults {
		if os.Getenv(key) == "" {
			os.Setenv(key, defaultValue)
		}
	}

	// ユーザーディレクトリ関連
	if os.Getenv("USERPROFILE") == "" {
		if homedrive := os.Getenv("HOMEDRIVE"); homedrive != "" {
			if homepath := os.Getenv("HOMEPATH"); homepath != "" {
				os.Setenv("USERPROFILE", homedrive+homepath)
			}
		}
	}

	// アプリケーションデータディレクトリ
	if os.Getenv("APPDATA") == "" && os.Getenv("USERPROFILE") != "" {
		os.Setenv("APPDATA", os.Getenv("USERPROFILE")+"\\AppData\\Roaming")
	}

	if os.Getenv("LOCALAPPDATA") == "" && os.Getenv("USERPROFILE") != "" {
		os.Setenv("LOCALAPPDATA", os.Getenv("USERPROFILE")+"\\AppData\\Local")
	}
}

func (s *Shell) enhanceWindowsPath() {
	currentPath := os.Getenv("PATH")
	if currentPath == "" {
		currentPath = ""
	}

	// Windows標準パスを追加
	standardPaths := []string{
		os.Getenv("SYSTEMROOT") + "\\System32",
		os.Getenv("SYSTEMROOT"),
		os.Getenv("SYSTEMROOT") + "\\System32\\WindowsPowerShell\\v1.0",
		os.Getenv("SYSTEMROOT") + "\\System32\\Wbem",
		"C:\\Windows\\System32",
		"C:\\Windows",
		"C:\\Windows\\System32\\WindowsPowerShell\\v1.0",
		"C:\\Windows\\System32\\Wbem",
	}

	// 現在のPATHを分割
	pathEntries := strings.Split(currentPath, ";")
	existingPaths := make(map[string]bool)
	for _, path := range pathEntries {
		if path != "" {
			existingPaths[strings.ToLower(path)] = true
		}
	}

	// 新しいパスを追加（重複チェック）
	var newPaths []string
	newPaths = append(newPaths, pathEntries...)

	for _, stdPath := range standardPaths {
		if stdPath != "" && stdPath != "\\System32" && !existingPaths[strings.ToLower(stdPath)] {
			// パスが実際に存在するかチェック
			if _, err := os.Stat(stdPath); err == nil {
				newPaths = append(newPaths, stdPath)
				existingPaths[strings.ToLower(stdPath)] = true
			}
		}
	}

	// 新しいPATHを設定
	newPath := strings.Join(newPaths, ";")
	os.Setenv("PATH", newPath)
}

func (s *Shell) expandWindowsEnvironmentVariable(input string) string {
	if runtime.GOOS != "windows" {
		return input
	}

	result := input

	// %VARIABLE% 形式の環境変数を展開
	for {
		start := strings.Index(result, "%")
		if start == -1 {
			break
		}

		end := strings.Index(result[start+1:], "%")
		if end == -1 {
			break
		}

		end = start + 1 + end
		varName := result[start+1 : end]
		varValue := os.Getenv(varName)

		if varValue != "" {
			result = result[:start] + varValue + result[end+1:]
		} else {
			// 変数が見つからない場合はそのまま残す
			break
		}
	}

	return result
}

func (s *Shell) getWindowsSpecialFolders() map[string]string {
	folders := make(map[string]string)

	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		folders["Desktop"] = userProfile + "\\Desktop"
		folders["Documents"] = userProfile + "\\Documents"
		folders["Downloads"] = userProfile + "\\Downloads"
		folders["Pictures"] = userProfile + "\\Pictures"
		folders["Music"] = userProfile + "\\Music"
		folders["Videos"] = userProfile + "\\Videos"
	}

	if programFiles := os.Getenv("PROGRAMFILES"); programFiles != "" {
		folders["ProgramFiles"] = programFiles
	}

	if programFilesX86 := os.Getenv("PROGRAMFILES(X86)"); programFilesX86 != "" {
		folders["ProgramFiles(x86)"] = programFilesX86
	}

	if systemRoot := os.Getenv("SYSTEMROOT"); systemRoot != "" {
		folders["Windows"] = systemRoot
		folders["System32"] = systemRoot + "\\System32"
		folders["SysWOW64"] = systemRoot + "\\SysWOW64"
	}

	return folders
}

func (s *Shell) isWindowsSystemCommand(command string) bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// cmd.exe の内部コマンド
	internalCommands := []string{
		"assoc", "attrib", "break", "bcdedit", "cacls", "call", "cd", "chcp",
		"chdir", "chkdsk", "chkntfs", "choice", "cls", "cmd", "color", "comp",
		"compact", "convert", "copy", "date", "del", "dir", "diskpart", "doskey",
		"driverquery", "echo", "endlocal", "erase", "exit", "expand", "fc",
		"find", "findstr", "for", "format", "fsutil", "ftype", "goto", "gpresult",
		"graftabl", "help", "icacls", "if", "label", "md", "mkdir", "mklink",
		"mode", "more", "move", "openfiles", "path", "pause", "popd", "print",
		"prompt", "pushd", "rd", "recover", "rem", "ren", "rename", "replace",
		"rmdir", "robocopy", "set", "setlocal", "setx", "sfc", "shift", "shutdown",
		"sort", "start", "subst", "systeminfo", "tasklist", "taskkill", "time",
		"title", "tree", "type", "ver", "verify", "vol", "xcopy", "wmic",
	}

	lowerCommand := strings.ToLower(command)
	for _, cmd := range internalCommands {
		if cmd == lowerCommand {
			return true
		}
	}

	return false
}
