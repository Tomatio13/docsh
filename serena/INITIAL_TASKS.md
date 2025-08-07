# 初期タスク集（優先度付き）

## 高
- ドキュメント整合性の改善
  - `README_ja.md` に "Cherry Shell" 表記が残存 → `docsh` に統一
  - `build.sh` の表示/出力名が `cherrysh` → `docsh` へ
  - `Makefile` の `APP_NAME = go-zsh` → `docsh` へ（バイナリ名含め一貫性確保）
- リリースバッジ等のリンク整備（`README.md` のプレースホルダ `your-username` を実リポジトリに）

## 中
- マッピング追加/見直し
  - `docker compose` 系の簡易マッピング（`compose up`, `compose ps`, `compose logs -f` など）
  - よく使う `logs -f <container>` のショート
- I18n の未翻訳キー/文言の見直し
- `data/config.yaml` のデフォルト値レビュー（色/詳細度/履歴件数）

## 低
- `ps` 出力整形の拡張（列幅/フィルタ/色）
- テーマ/プロンプトのバリエーション追加

## 品質/セキュリティチェック（継続）
- 依存の更新: `go mod tidy && go mod download`
- Lint/Test: 将来的に `golangci-lint` 導入、`go test ./...` 整備
