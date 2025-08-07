# Serena MCP オンボーディング（docsh）

このディレクトリは Serena MCP がこのリポジトリを素早く理解・運用できるためのオンボーディング資料をまとめたものです。

- 目的: Docker コマンドマッピングシェル「docsh」のビルド/実行と、主要コード領域の把握、よく使う検索レシピ、初期タスクの提示
- 対象: Serena MCP（および開発者）

## クイックスタート
- ビルド/実行手順: `serena/BUILD_RUN.md`
- リポジトリ構成と主要ポイント: `serena/CODEMAP.md`
- よく使う検索レシピ: `serena/SEARCH_GUIDE.md`
- 初期タスク集: `serena/INITIAL_TASKS.md`
- チェックリスト: `serena/ONBOARDING_CHECKLIST.md`

## 実行例
```bash
# ビルド
go mod tidy
go build -o docsh main.go

# 対話モード
./docsh

# 直接実行
./docsh ps
./docsh images
./docsh run nginx
```

## 注意
- このプロジェクトは Go 1.22+ が必要です
- Docker コマンドを実行する機能があるため、Docker がインストールされ起動していると利便性が高いです
