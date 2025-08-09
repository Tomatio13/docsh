package main

import (
	"fmt"
	"os"
	"path/filepath"

	"docknaut/config"
	"docknaut/i18n"
	"docknaut/shell"
)

func main() {
	// データパスを設定（実行ファイルからの相対パス）
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Warning: Could not determine executable path: %v\n", err)
		execPath = "."
	}
	dataPath := filepath.Join(filepath.Dir(execPath), "data")

	// データディレクトリが存在しない場合はカレントディレクトリの data を使用
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		dataPath = "data"
	}

	// 設定を読み込む
	cfg := config.NewConfig()
	cfg.DataPath = dataPath
	if err := cfg.LoadConfigFile(); err != nil {
		fmt.Printf("Warning: Could not load config file: %v\n", err)
	}

	// 言語を取得
	language := cfg.GetLanguage(os.Args)

	// 国際化を初期化
	if err := i18n.Init(language); err != nil {
		fmt.Printf("Warning: Could not initialize i18n: %v\n", err)
		// フォールバックとして英語で初期化
		i18n.Init("en")
	}

	// アプリケーション情報を表示
	// fmt.Println(i18n.T("app.title"))
	// fmt.Println(i18n.T("app.description"))
	// fmt.Println(i18n.T("app.exit_instruction"))
	// fmt.Println(i18n.T("shell.runtime_separator"))
	// fmt.Println(i18n.T("shell.runtime_info"))
	// fmt.Printf(i18n.T("shell.runtime_os")+"\n", runtime.GOOS)
	// fmt.Printf(i18n.T("shell.runtime_arch")+"\n", runtime.GOARCH)
	// fmt.Println(i18n.T("shell.runtime_separator"))

	// シェルを初期化
	s := shell.NewShell(cfg, dataPath)

	// コマンドライン引数が渡された場合は直接実行
	if len(os.Args) > 1 {
		// 引数を結合してコマンドとして実行
		command := ""
		for i, arg := range os.Args[1:] {
			if i > 0 {
				command += " "
			}
			command += arg
		}

		// 直接コマンドを実行
		if err := s.ExecuteCommand(command); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// インタラクティブモードでシェルを開始（Bubble Tea REPL）
	s.Start()
}
