package shell

import (
	"fmt"
	"os"
	"runtime"
)

// enableWindowsAnsiSupport enables ANSI escape sequences on Windows
func enableWindowsAnsiSupport() {
	if runtime.GOOS != "windows" {
		return
	}
	
	// Windows 10以降ではANSIエスケープシーケンスが標準でサポートされているため
	// 特別な設定は不要。ここでは単純にフラグとして機能
}

// ensureCleanOutput はコマンド実行後の出力を正規化します
func (s *Shell) ensureCleanOutput() {
	// 標準出力のフラッシュ
	os.Stdout.Sync()
	
	// Windows環境での追加対策
	if runtime.GOOS == "windows" {
		// バッファを確実にフラッシュ
		fmt.Print("") // 空文字で出力バッファをフラッシュ
	}
}

// printWithFlush は出力後に確実にフラッシュを行います
func (s *Shell) printWithFlush(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	s.ensureCleanOutput()
}

// printlnWithFlush は改行付き出力後に確実にフラッシュを行います
func (s *Shell) printlnWithFlush(args ...interface{}) {
	fmt.Println(args...)
	s.ensureCleanOutput()
}