docsh Docker Compose拡張 要件定義書
1. 概要
1.1 目的
既存のdocshにDocker Compose管理機能を追加し、プロジェクト単位でのDocker操作を可能にする。
1.2 基本方針

既存機能は一切変更しない
新機能は追加のみ
シンプルで分かりやすい拡張

2. 追加機能
2.1 プロジェクト検出

#### プロジェクト表示モード
```
🐳 docsh $ ps --by-project

📦 Active Projects:
┌─ myapp (/home/user/projects/myapp) ──────────────┐
│ 🌐 frontend  running  0.0.0.0:8080->8080/tcp    │
│ 🔧 api      running  0.0.0.0:3000->3000/tcp     │
│ 🗃️ postgres running  5432/tcp                   │
└──────────────────────────────────────────────────┘

┌─ backend (/home/user/work/backend) ──────────────┐
│ 🚀 app      running  0.0.0.0:4000->4000/tcp     │
│ 📦 redis    running  6379/tcp                   │
└──────────────────────────────────────────────────┘

🐳 Individual containers:
confident_benz    running  0.0.0.0:4010->8080/tcp
```

#### プロジェクト詳細表示
```
🐳 docsh $ project myapp ps

Project: myapp (/home/user/projects/myapp)
Config: docker-compose.yml

🌐 frontend   running   0.0.0.0:8080->8080/tcp   (depends: api)
🔧 api       running   0.0.0.0:3000->3000/tcp   (depends: postgres)
🗃️ postgres  running   5432/tcp                 (healthy)
```

#### プロジェクト操作
プロジェクト名を指定してサービスを操作
bash# 新規コマンド群
🐳 docsh $ project myapp ps           # プロジェクトのコンテナ一覧
🐳 docsh $ project myapp logs api     # 特定サービスのログ
🐳 docsh $ project myapp restart api  # 特定サービスの再起動
🐳 docsh $ project myapp stop         # プロジェクト全停止

```
3. 検出ロジック
go// Docker APIからComposeラベルを読み取り
Labels["com.docker.compose.project"]            // プロジェクト名
Labels["com.docker.compose.project.working_dir"] // 作業ディレクトリ
Labels["com.docker.compose.service"]            // サービス名


↓これは要らない。
🐳 Dockerコマンド:
  docker ps, docker run, docker exec, docker logs, など

↓以下の並びと表示にしてほしい
🐳 Docker ライフサイクルコマンド:
  pull <image>            レジストリからイメージを取得
  start <container>       停止中のコンテナを開始
  stop <container>        実行中のコンテナを停止
  exec <container> <cmd>  コンテナでコマンド実行
  login <container>        コンテナにログイン (/bin/bash)
  rm [--force] <container> コンテナを削除
  rmi [--force] <image>   イメージを削除

🐳 Docker Compose ライフサイクルコマンド:
  project ps                          サービス毎にコンテナ一覧
  project <service> start             特定サービスの開始
  project <service> logs              特定サービスのログ
  project <service> restart           特定サービスの再起動
  project <service> stop              サービス全停止
  ps --by-project                     サービス毎にコンテナ一覧

🔧 内蔵コマンド:
  help                    使用方法
  mapping [list|search|show] <args>  コマンドマッピングを管理
  alias <name>=<command>              エイリアス設定
  theme [name]                       テーマ設定
  config [show|set]                 設定管理
  exit                              シェル終了

