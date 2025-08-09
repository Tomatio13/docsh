package shell

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"docsh/i18n"
)

// isBuiltinCommand は内蔵コマンドかどうかを判定します
func (s *Shell) isBuiltinCommand(command string) bool {
	builtinCommands := []string{
		"dir", "ls", "cat", "type", "copy", "cp", "move", "mv",
		"del", "rm", "mkdir", "md", "rmdir", "rd", "cls", "clear",
		"echo", "set", "where",
	}

	for _, builtin := range builtinCommands {
		if command == builtin {
			return true
		}
	}
	return false
}

// executeBuiltinCommand は内蔵コマンドを実行します
func (s *Shell) executeBuiltinCommand(command string, args []string) error {
	switch command {
	case "dir", "ls":
		return s.listDirectory(args)
	case "cat", "type":
		return s.catFile(args)
	case "copy", "cp":
		return s.copyFile(args)
	case "move", "mv":
		return s.moveFile(args)
	case "del", "rm":
		return s.deleteFile(args)
	case "mkdir", "md":
		return s.makeDirectory(args)
	case "rmdir", "rd":
		return s.removeDirectory(args)
	case "cls", "clear":
		return s.clearScreen()
	case "echo":
		return s.echoCommand(args)
	case "set":
		return s.setCommand(args)
	case "where":
		return s.whereCommand(args)
	default:
		return fmt.Errorf("unknown builtin command: %s", command)
	}
}

// listDirectory はディレクトリの内容を表示します
func (s *Shell) listDirectory(args []string) error {
	dir := "."
	showAll := false
	longFormat := false
	sortByTime := false
	reverseOrder := false

	// 引数を解析
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			// フラグを処理
			for _, flag := range arg[1:] {
				switch flag {
				case 'a':
					showAll = true
				case 'l':
					longFormat = true
				case 't':
					sortByTime = true
				case 'r':
					reverseOrder = true
				}
			}
		} else {
			// ディレクトリ名
			dir = arg
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// 隠しファイルをフィルタリング
	if !showAll {
		var filteredEntries []os.DirEntry
		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name(), ".") {
				filteredEntries = append(filteredEntries, entry)
			}
		}
		entries = filteredEntries
	}

	// ソート処理
	if sortByTime {
		// 時間順でソート
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				info1, _ := entries[i].Info()
				info2, _ := entries[j].Info()
				if info1.ModTime().Before(info2.ModTime()) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}
	} else {
		// 名前順でソート（デフォルト）
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].Name() > entries[j].Name() {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}
	}

	// 逆順にする
	if reverseOrder {
		for i := 0; i < len(entries)/2; i++ {
			j := len(entries) - 1 - i
			entries[i], entries[j] = entries[j], entries[i]
		}
	}

	// 表示
	if longFormat {
		// 詳細表示
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			var typeStr string
			var sizeStr string
			if entry.IsDir() {
				typeStr = "d"
				sizeStr = fmt.Sprintf("%8s", "<DIR>")
			} else {
				typeStr = "-"
				sizeStr = fmt.Sprintf("%8d", info.Size())
			}

			// 権限表示（簡易版）
			mode := info.Mode()
			perms := typeStr
			if mode&0400 != 0 {
				perms += "r"
			} else {
				perms += "-"
			}
			if mode&0200 != 0 {
				perms += "w"
			} else {
				perms += "-"
			}
			if mode&0100 != 0 {
				perms += "x"
			} else {
				perms += "-"
			}
			perms += "------" // 簡易版なので他のユーザー権限は省略

			fmt.Printf("%s %s %s %s\n",
				perms,
				info.ModTime().Format("Jan 02 15:04"),
				sizeStr,
				entry.Name())
		}
	} else {
		// 簡易表示
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			var sizeStr string
			if entry.IsDir() {
				sizeStr = "<DIR>"
			} else {
				sizeStr = fmt.Sprintf("%d", info.Size())
			}

			fmt.Printf("%s %8s %s %s\n",
				info.ModTime().Format("2006-01-02 15:04"),
				sizeStr,
				"",
				entry.Name())
		}
	}

	return nil
}

// catFile はファイルの内容を表示します
func (s *Shell) catFile(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cat <file>")
	}

	for i, filename := range args {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("windows.cat_error")+"\n", filename, err)
			continue
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("windows.cat_error")+"\n", filename, err)
			continue
		}

		contentStr := string(content)
		fmt.Print(contentStr)

		// 複数ファイルの場合は改行を追加
		if i < len(args)-1 {
			fmt.Println() // 複数ファイルの場合は改行を追加
		}
	}

	return nil
}

// clearScreen は画面をクリアします
func (s *Shell) clearScreen() error {
	fmt.Print("\033[2J\033[H")
	return nil
}

// copyFile はファイルをコピーします
func (s *Shell) copyFile(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: copy <source> <destination>")
	}

	source := args[0]
	destination := args[1]

	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	fmt.Printf(i18n.T("windows.files_copied") + "\n")
	return nil
}

// moveFile はファイルを移動します
func (s *Shell) moveFile(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: move <source> <destination>")
	}

	source := args[0]
	destination := args[1]

	err := os.Rename(source, destination)
	if err != nil {
		return err
	}

	fmt.Printf(i18n.T("windows.files_moved") + "\n")
	return nil
}

// deleteFile はファイルを削除します
func (s *Shell) deleteFile(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: del <file>")
	}

	// オプション解析
	recursive := false
	force := false
	var files []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			// オプション処理
			for _, flag := range arg[1:] {
				switch flag {
				case 'r', 'R':
					recursive = true
				case 'f':
					force = true
				}
			}
		} else {
			// ファイル名
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("usage: rm [-rf] <file>")
	}

	deletedCount := 0
	for _, filename := range files {
		var err error
		if recursive {
			err = os.RemoveAll(filename)
		} else {
			err = os.Remove(filename)
		}

		if err != nil {
			if !force {
				fmt.Fprintf(os.Stderr, i18n.T("windows.rm_error")+"\n", filename, err)
			}
			continue
		}
		deletedCount++
	}

	if deletedCount > 0 {
		fmt.Printf(i18n.T("windows.files_deleted")+"\n", deletedCount)
	}

	return nil
}

// makeDirectory はディレクトリを作成します
func (s *Shell) makeDirectory(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mkdir <directory>")
	}

	for _, dirname := range args {
		err := os.MkdirAll(dirname, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

// removeDirectory はディレクトリを削除します
func (s *Shell) removeDirectory(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: rmdir <directory>")
	}

	for _, dirname := range args {
		err := os.Remove(dirname)
		if err != nil {
			return err
		}
	}

	return nil
}

// echoCommand はテキストを表示します
func (s *Shell) echoCommand(args []string) error {
	fmt.Println(strings.Join(args, " "))
	return nil
}

// setCommand は環境変数を表示または設定します
func (s *Shell) setCommand(args []string) error {
	if len(args) == 0 {
		// 全環境変数を表示
		for _, env := range os.Environ() {
			fmt.Println(env)
		}
		return nil
	}

	// 環境変数の設定は実装が複雑なため、表示のみサポート
	varName := args[0]
	if value, exists := os.LookupEnv(varName); exists {
		fmt.Printf("%s=%s\n", varName, value)
	}

	return nil
}

// whereCommand はコマンドのパスを表示します
func (s *Shell) whereCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: where <command>")
	}

	command := args[0]

	// PATH環境変数からコマンドを検索
	path := os.Getenv("PATH")
	pathDirs := strings.Split(path, string(os.PathListSeparator))

	var extensions []string
	if runtime.GOOS == "windows" {
		pathext := os.Getenv("PATHEXT")
		if pathext != "" {
			extensions = strings.Split(strings.ToLower(pathext), ";")
		} else {
			extensions = []string{".exe", ".cmd", ".bat"}
		}
	} else {
		extensions = []string{""}
	}

	for _, dir := range pathDirs {
		for _, ext := range extensions {
			fullPath := filepath.Join(dir, command+ext)
			if _, err := os.Stat(fullPath); err == nil {
				fmt.Println(fullPath)
				return nil
			}
		}
	}

	return fmt.Errorf("command not found: %s", command)
}
