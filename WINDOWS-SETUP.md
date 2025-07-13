# 🌸 Cherry Shell - Windows セットアップガイド

## Windows 10/11 64-bit での使用方法

### 必要環境
- Windows 10 (64-bit) またはWindows 11
- コマンドプロンプトまたはPowerShell

### インストール手順

1. **ファイルのダウンロード**
   - `cherrysh-windows-x64.exe` をダウンロード
   - 任意のフォルダに配置（例：`C:\Tools\CherryShell\`）

2. **実行権限の設定**
   - Windows Defenderで警告が出る場合は、除外設定を行う
   - ファイルを右クリック → プロパティ → 「ブロックの解除」をチェック

3. **基本的な使用**
   ```cmd
   # コマンドプロンプトで実行
   C:\Tools\CherryShell> cherrysh-windows-x64.exe
   ```

### 動作テスト

**自動テストスクリプト**を使用して動作確認できます：
```cmd
test-windows.bat
```

**手動テスト**：
```cmd
cherrysh-windows-x64.exe

# Cherry Shell内で以下のコマンドをテスト
cherry:C:\> ls
cherry:C:\> pwd
cherry:C:\> cd Users
cherry:C:\Users> ls -la
cherry:C:\Users> exit
```

### 設定ファイル

Windows環境では以下の場所に `.cherryshrc` を配置できます：

1. **カレントディレクトリ**: `.cherryshrc`
2. **ユーザーホーム**: `%USERPROFILE%\.cherryshrc`

**設定例（.cherryshrc）**：
```bash
# Cherry Shell Configuration File
# 🌸 Cherry Shell - Beautiful & Simple Shell 🌸

# プロンプト設定
PROMPT="cherry:%s> "

# テーマ設定  
THEME="robbyrussell"

# エイリアス設定
alias ll='ls -la'
alias ..='cd ..'
alias cls='clear'

# Windows固有のエイリアス
alias dir='ls'
alias type='cat'
```

### Windows固有の機能

Cherry ShellはWindows環境で以下の機能をサポートします：

- **Windowsパス**: `C:\path\to\file` 形式の認識
- **ドライブ変更**: `cd C:`, `cd D:` などの操作
- **実行ファイル検索**: `.exe`, `.com`, `.bat`, `.cmd` の自動認識
- **Windows環境変数**: `%USERPROFILE%`, `%PATH%` などの使用

### トラブルシューティング

**Windows Defenderの警告**
- 実行ファイルを除外リストに追加
- または信頼できるフォルダに配置

**文字化けする場合**
- コマンドプロンプトで `chcp 65001` を実行してUTF-8に設定

**実行できない場合**
- ファイルのプロパティでブロック解除を確認
- 管理者権限で実行を試行

### パフォーマンス

- **ファイルサイズ**: 約3.0MB
- **起動時間**: 1秒未満
- **メモリ使用量**: 約10-20MB

---

**🌸 Cherry Shell - Windows環境でも美しく動作します 🌸**