# 🐳 docsh - Docker コマンドマッピングシェル（日本語）

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey?style=for-the-badge" alt="Platform">
  <img src="https://img.shields.io/badge/i18n-English%20%7C%20Japanese-blue?style=for-the-badge" alt="i18n">
</p>

<p align="center">
  <a href="README.md"><img src="https://img.shields.io/badge/english-document-white.svg" alt="EN doc"></a>
  <a href="README_ja.md"><img src="https://img.shields.io/badge/ドキュメント-日本語-white.svg" alt="JA doc"/></a>
</p>

**docsh** は、Linux 風の直感的な操作で Docker を扱えるインタラクティブシェルです。よく使う Linux コマンドを Docker コマンドにマッピングして実行できます。

## ✨ 特長

- **🐳 Docker コマンドマッピング**: `ls`→`docker ps`、`kill`→`docker stop` など直感的
- **🌍 クロスプラットフォーム**: Windows / macOS / Linux 対応
- **⚡ 対話型シェル**: 快適な対話 UI（Bubble Tea ベース）
- **🔧 設定可能**: YAML と `~/.docshrc` で柔軟にカスタマイズ
- **🔗 エイリアス**: コマンドショートカットを簡単作成
- **🌐 i18n**: 英語・日本語に対応
- **📜 履歴**: コマンド履歴の保持
- **🎨 プロンプト**: カスタマイズ可能な見た目

## 📦 インストール

### 🚀 バイナリ

最新リリースをダウンロードしてください。

### 🛠️ ソースからビルド

```bash
git clone https://github.com/your-username/docsh.git
cd docsh
go build -o docsh main.go
```

全プラットフォーム向けビルド:

```bash
./build.sh
```

## 🚀 使い方

### 対話モード

```bash
# 対話シェル起動
./docsh
```

### 直接実行

```bash
# Linux風 → Docker へマッピングして実行
./docsh ls                   # -> docker ps
./docsh kill myapp           # -> docker stop myapp
./docsh rm myapp             # -> docker rm myapp

# そのまま Docker コマンド文字列を渡すことも可能
./docsh "docker ps"
./docsh "docker images"
```

### よく使う操作（対話シェル内）

```bash
# コンテナ管理（マッピング/内蔵）
ps                           # docker ps
logs <container>             # docker logs <container>
exec <container> <command>   # docker exec <container> <command>
stop <container>             # docker stop <container>
rm <container>               # docker rm <container>
rmi <image>                  # docker rmi <image>
```

## 🐳 Docker ライフサイクルコマンド（help より）

```
🐳 Docker ライフサイクルコマンド:
  pull <image>            レジストリからイメージを取得
  start <container>       停止中のコンテナを開始
  stop <container>        実行中のコンテナを停止
  exec <container> <cmd>  コンテナでコマンド実行
  login <container>        コンテナにログイン (/bin/bash)
  rm [--force] <container> コンテナを削除
  rmi [--force] <image>   イメージを削除
  log     <container>          コンテナのログを表示
  tail -f <container>          コンテナのログをリアルタイム表示        
  top                                       リソース使用状況を表示
  htop                                      リソース使用状況をグラフ表示
⚠️  注意:  tail -fと、topを終了するには、表示中にexitと入力してください。
```

## 📦 プロジェクト/Compose 運用（project 系コマンド）

Compose ラベルが付いたコンテナ群を「プロジェクト」として扱い、サービス単位の操作を簡単にします。

- **一覧表示（全プロジェクト）**
  ```bash
  ps --by-project
  # または
  project ps
  ```

- **プロジェクトのサービス一覧**
  ```bash
  project <project> ps
  ```

- **サービスのログ表示（推奨）**
  ```bash
  project <project> logs <service> -f --tail 100
  # docker logs の引数順に合わせ、[OPTIONS] を先に、最後にコンテナ名を解決して実行します
  ```

- **サービス省略形（サービス名が全体で一意な場合のみ）**
  ```bash
  project <service> logs -f --tail 100
  # 同名サービスが複数プロジェクトに存在するときは曖昧エラーになります
  ```

- **プロジェクト/サービス開始（Compose 対応）**
  ```bash
  # プロジェクト全体（docker-compose.yml があれば compose を優先）
  project <project> start

  # 特定サービス
  project <project> start <service>
  ```

- **再起動/停止（Compose 対応）**
  ```bash
  project <project> restart [<service>]
  project <project> stop    [<service>]
  ```

ヘルプに表示される対応表（抜粋）:

```
🐳 Docker Compose ライフサイクルコマンド:
  project ps                          サービス毎にコンテナ一覧
  project <service> start             特定サービスの開始
  project <service> logs              特定サービスのログ
  project <service> restart           特定サービスの再起動
  project <service> stop              サービス全停止
  ps --by-project                     サービス毎にコンテナ一覧
```

## 🌐 言語設定

`~/.docshrc` で設定できます。

```bash
# ~/.docshrc
LANG="ja"   # または "en"
```

変更後は `docsh` を再起動してください。

## ⚙️ 設定

設定は `data/config.yaml`（出荷デフォルト）と `~/.docshrc`（ユーザー上書き）で管理します。

```yaml
shell:
  prompt: "🐳 docsh> "
  history_size: 1000
  auto_complete: true
  dry_run_mode: false
  show_mappings: true

docker:
  default_options: []
  timeout: 30
  auto_detect: true

i18n:
  default_language: "ja"
  supported_languages: ["ja", "en"]
  locale_dir: "data/locales"
  fallback_language: "en"
```

## 🔗 エイリアス

YAML と `~/.docshrc` の両方で設定できます。

```yaml
# data/config.yaml
aliases:
  dps: "docker ps"
  dpa: "docker ps -a"
  di:  "docker images"
```

```bash
# ~/.docshrc
alias dps="docker ps"
alias dpa="docker ps -a"
alias di="docker images"
```

## 🛠️ 開発

```bash
# 現在のプラットフォーム向けビルド
go build -o docsh main.go

# 全プラットフォーム向けビルド
./build.sh

# テスト
go test ./...
```

## 📁 プロジェクト構成

```
docsh/
├── main.go                 # エントリポイント
├── config/                 # 設定
│   ├── config.go
│   ├── alias.go
│   └── yaml.go
├── i18n/                   # 多言語化
│   └── i18n.go
├── shell/                  # シェル本体
│   ├── shell.go
│   ├── command.go
│   └── prompt.go
├── tui/                    # TUI コンポーネント
├── data/                   # 設定・データ
│   ├── config.yaml
│   ├── mappings.yaml
│   └── locales/
└── themes/                 # テーマ
    ├── theme.go
    └── banner.go
```

## 🌍 対応言語

- **🇯🇵 日本語 (ja)**: フルサポート
- **🇺🇸 英語 (en)**: フルサポート

## 🤝 コントリビュート

1. リポジトリをフォーク
2. フィーチャーブランチを作成
3. 変更を加える（必要に応じてテスト追加）
4. プルリクエストを作成

## 📄 ライセンス

本プロジェクトは MIT ライセンスで提供されます。

---

<p align="center">
🐳 <strong>docsh</strong> - Docker 操作を、よりシンプルに。
</p>

