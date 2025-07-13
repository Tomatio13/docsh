# 🌸 Cherry Shell 🌸

**Windows専用の美しくシンプルなシェル - Cherry Shell**

Cherry Shell（チェリーシェル）は、桜貝（Sakura-gai）という美しい桃色の二枚貝にちなんで名付けられた、Windows専用の小さいながらも美しいシェルです。

## ✨ 特徴

- 🎨 **美しいテーマシステム** - 4つのビルトインテーマ（default、robbyrussell、agnoster、simple）
- 🔧 **Unix風コマンド** - WindowsでもUnix風のコマンド（ls、cat、clear等）が使用可能
- ⚙️ **設定ファイル対応** - .cherryshrcファイルでエイリアスや環境変数を設定
- 🚀 **外部プログラム実行** - 既存のexeファイルやWindowsコマンドを実行可能
- 🖥️ **Windows最適化** - Windows環境変数とパス処理に完全対応
- 💝 **軽量・高速** - Goで書かれた軽量なシェル

## 🔧 インストール・ビルド

### 必要環境
- Go 1.22.2以上
- Windows 10/11 (64-bit推奨)

### ビルド方法

```cmd
# リポジトリをクローン
git clone <repository-url>
cd go-zsh

# Windows ビルド
GOOS=windows GOARCH=amd64 go build -o cherrysh-windows-x64.exe .

# Linux向けビルド
 go build -o cherrysh .

# 実行
cherrysh-windows-x64.exe
```

### バッチファイルでのビルド

```cmd
# build.batを実行
build.bat
```

## 🎯 使用方法

### 基本コマンド

```cmd
# ディレクトリ移動
cd <directory>
cd Desktop        # 特殊フォルダ対応
cd Documents      # 特殊フォルダ対応
cd C:             # ドライブ変更

# ファイル一覧表示
ls                 # 基本表示
ls -l              # 詳細表示
ls -a              # 隠しファイル含む
ls -la             # 詳細＋隠しファイル

# ファイル内容表示
cat <filename>

# 現在のディレクトリ表示
pwd

# 画面クリア
clear

# ファイル・ディレクトリ操作
cp <source> <dest>     # ファイルコピー
mv <source> <dest>     # ファイル移動
rm <file>              # ファイル削除
mkdir <directory>      # ディレクトリ作成
rmdir <directory>      # ディレクトリ削除

# 環境変数・システム情報
env                    # 環境変数一覧
which <command>        # コマンドのパス検索
echo <text>            # テキスト出力

# シェル終了
exit
```

### エイリアス管理

```cmd
# エイリアス一覧表示
alias

# 新しいエイリアス作成
alias ll='ls -la'
alias la='ls -la'
alias ..='cd ..'
alias ...='cd ../..'
```

### テーマ管理

```cmd
# 利用可能なテーマ一覧
theme

# テーマ変更
theme default      # デフォルトテーマ
theme robbyrussell # oh-my-zsh風テーマ
theme agnoster     # agnoster風テーマ
theme simple       # シンプルテーマ
```

## ⚙️ 設定ファイル（.cherryshrc）

Cherry Shellは起動時に以下の場所から設定ファイルを読み込みます：

1. カレントディレクトリの `.cherryshrc`
2. ホームディレクトリの `.cherryshrc`
3. `%USERPROFILE%\.cherryshrc`

### 設定例

```cmd
# Cherry Shell Configuration File
# 🌸 Cherry Shell - Beautiful & Simple Shell 🌸

# プロンプト設定
PROMPT="cherry:%s$ "

# テーマ設定
THEME="robbyrussell"

# エイリアス設定
alias ll='ls -la'
alias la='ls -la'
alias l='ls -l'
alias grep='grep --color=auto'
alias ..='cd ..'
alias ...='cd ../..'

# カスタム環境変数
# EDITOR="notepad"
# BROWSER="chrome"
```

## 🎨 テーマシステム

Cherry Shellは4つの美しいビルトインテーマを提供します：

### default
```
cherry:C:\Users\user$ 
```
- シンプルでクリーンなデフォルトテーマ
- 基本的なカラーサポート

### robbyrussell
```
C:\Users\user ➜ 
```
- oh-my-zsh風のカラフルなテーマ
- シアンとレッドのカラーリング
- 矢印記号（➜）を使用

### agnoster
```
user@hostname C:\Users\user $ 
```
- ユーザー名とホスト名を表示
- グリーンとブルーのカラーリング
- より詳細な情報表示

### simple
```
C:\Users\user $ 
```
- 最小限のシンプルなデザイン
- カラーなしのクリーンな表示

## 🖥️ Windows専用機能

### 特殊フォルダ対応
```cmd
cd Desktop        # デスクトップフォルダ
cd Documents      # ドキュメントフォルダ
cd Downloads      # ダウンロードフォルダ
cd Pictures       # ピクチャフォルダ
cd Music          # ミュージックフォルダ
cd Videos         # ビデオフォルダ
cd ProgramFiles   # Program Filesフォルダ
cd Windows        # Windowsフォルダ
cd System32       # System32フォルダ
```

### 環境変数展開
```cmd
cd %USERPROFILE%\Documents
cd %PROGRAMFILES%\MyApp
```

### Windows実行ファイル対応
- .exe、.com、.bat、.cmd ファイルの自動実行
- PATHからの実行ファイル検索
- cmd.exe内部コマンドの実行

### Windows環境変数の自動設定
- PATHEXT、TEMP、TMP、SYSTEMROOT等の設定
- USERPROFILE、APPDATA、LOCALAPPDATA等の設定
- Windows標準パスの自動追加

## 🚀 外部プログラム実行

Cherry Shellは以下の方法で外部プログラムを実行できます：

```cmd
# 実行ファイルの直接実行
notepad.exe myfile.txt
calc.exe

# PATHからの実行
git status
npm install

# cmd.exe内部コマンド
dir
type myfile.txt
```

## 🛠️ 開発・デバッグ

### 実行時情報
起動時に以下の情報が表示されます：
- Runtime OS: windows
- Runtime ARCH: amd64
- ANSIカラーサポート状況

### 設定ファイルの自動生成
初回起動時に `%USERPROFILE%\.cherryshrc` が自動生成されます。

## 📝 ライセンス

MIT License

---

**🌸 Cherry Shell - 小さいながらも美しい、あなたのためのシェル 🌸**

