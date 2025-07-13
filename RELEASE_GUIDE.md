# 🚀 CherryShell リリースガイド

このガイドでは、CherryShellの新しいリリースを作成する方法を説明します。

## 📋 リリース前の準備

### 1. バージョン管理の確認
- 全ての変更がmainブランチにマージされていることを確認
- ビルドテストが全て通過していることを確認
- 必要に応じてCHANGELOG.mdを更新

### 2. ローカルテスト
```bash
# ローカルでビルドテストを実行（現在のプラットフォーム用）
./build.sh

# 実行ファイルが正常に動作することを確認
./cherrysh  # Linux/macOS
# または
./cherrysh.exe  # Windows

# 複数プラットフォームでのテストが必要な場合は手動でビルド
GOOS=linux GOARCH=amd64 go build -o test-linux .
GOOS=darwin GOARCH=amd64 go build -o test-macos .
GOOS=windows GOARCH=amd64 go build -o test-windows.exe .
```

## 🏷️ リリースタグの作成

### 1. セマンティックバージョニング
CherryShellは[セマンティックバージョニング](https://semver.org/)を使用します：

- **MAJOR**: 互換性のない API の変更
- **MINOR**: 後方互換性のある機能の追加
- **PATCH**: 後方互換性のあるバグ修正

例：`v1.2.3`

### 2. タグの作成とプッシュ
```bash
# 新しいバージョンタグを作成
git tag -a v1.2.3 -m "Release v1.2.3"

# タグをリモートにプッシュ
git push origin v1.2.3
```

## 🤖 自動リリースプロセス

タグがプッシュされると、GitHub Actionsが自動的に以下を実行します：

1. **マルチプラットフォームビルド**
   - Windows (64bit/32bit)
   - Linux (64bit)
   - macOS (Intel/Apple Silicon)

2. **圧縮ファイルの作成**
   - Windows: `.zip` 形式
   - Linux/macOS: `.tar.gz` 形式

3. **GitHub Releasesの作成**
   - 自動的にリリースノートを生成
   - 圧縮ファイルをアップロード

## 📝 リリースノートの編集

自動作成されたリリースを手動で編集することができます：

1. [GitHub Releases](https://github.com/your-username/cherryshell/releases)にアクセス
2. 対象のリリースの「Edit release」をクリック
3. `.github/release-template.md`を参考にして内容を編集
4. 「Update release」をクリック

### リリースノートの構成例

```markdown
## 🎉 CherryShell v1.2.3

### 新機能 / New Features
- 新しいテーマ「Sakura Night」を追加
- 自動補完機能の改善

### 改善 / Improvements
- 起動時間を30%短縮
- メモリ使用量を削減

### バグ修正 / Bug Fixes
- Windows環境でのパス処理の問題を修正
- 日本語入力時の文字化けを修正

### 技術的変更 / Technical Changes
- Go 1.22.2に更新
- 依存関係の更新
```

## 📦 配布ファイル

各リリースには以下のファイルが含まれます：

- `cherrysh-v1.2.3-windows-x64.zip`
- `cherrysh-v1.2.3-windows-x86.zip`
- `cherrysh-v1.2.3-linux-x64.tar.gz`
- `cherrysh-v1.2.3-macos-x64.tar.gz`
- `cherrysh-v1.2.3-macos-arm64.tar.gz`

## 🔧 トラブルシューティング

### ビルドが失敗する場合
1. GitHub Actionsのログを確認
2. ローカルで `./build.sh` を実行してエラーを確認
3. 依存関係の問題がないか確認

### リリースが作成されない場合
1. タグ名が `v*` の形式になっているか確認
2. GitHub Actionsの権限設定を確認
3. `GITHUB_TOKEN` の権限が適切か確認

## 🚨 緊急時の対応

### リリースを取り消す場合
1. GitHub Releasesページで該当リリースを削除
2. 必要に応じてタグも削除：
   ```bash
   git tag -d v1.2.3
   git push origin :refs/tags/v1.2.3
   ```

### ホットフィックスリリース
1. 修正をmainブランチに適用
2. パッチバージョンを上げてタグを作成
3. 通常のリリースプロセスに従う

## 📊 リリース後の確認

- [ ] 全プラットフォームのダウンロードリンクが正常に動作する
- [ ] 実行ファイルが正常に動作する
- [ ] READMEのバッジが最新バージョンを表示している
- [ ] 必要に応じてソーシャルメディアでリリースを告知

---

**注意**: このプロセスは自動化されていますが、リリース前には必ず手動でテストを実行してください。 