# コードマップ（主要ポイント）

- `main.go`: エントリポイント。`config` の読込、`i18n` 初期化、`shell.NewShell` 実行、引数ありは直実行/無しは対話モード
- `shell/shell.go`: シェル本体。エイリアス展開、パース、内蔵コマンド、マッピング/実行ハンドリング、ストリーミング系（`tail -f` → `docker logs -f`）
- `internal/parser/command.go`: コマンドパーサ。`ParsedCommand`、Linux/Docker/Builtin の判定、短/長オプション解析
- `internal/executor/docker.go`: 実行層。Docker-only モード前提の振る舞い、マッピング解決、`docker` 実行、`ps` 出力整形、DryRun など
- `internal/engine/mapping.go`: マッピングエンジン。Linux コマンド → Docker コマンドの対応付け検索
- `config/*.go`: 設定の読み込み/エイリアス展開、`data/config.yaml` 連携
- `data/`: 設定・マッピング・ロケールデータ
  - `data/config.yaml`: シェル挙動/表示/I18n 設定など
  - `data/mappings.yaml`: Linux → Docker コマンドマッピング定義
  - `data/locales/*.yaml`: 言語別テキスト
- `i18n/`: 多言語化初期化とメッセージ
- `themes/theme.go`: テーマ定義
- `shell/*.go`: 補完、プロンプト、出力整形、Git 連携などの補助

## 実行フロー概略
1. `main.go` で設定/I18n 初期化 → `shell.NewShell`
2. 入力を `internal/parser` で `ParsedCommand` に分解
3. `internal/executor` が `ParsedCommand` をもとに
   - Builtin／Linux（マッピング検索）／Docker を分岐
   - Docker 実行と出力整形
