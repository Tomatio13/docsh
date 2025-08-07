# ビルド/実行ガイド（docsh）

## 前提
- Go 1.22 以降
- （任意）Docker（`ps` や `logs` などの確認に利用）

## ビルド
```bash
# 依存関係
go mod tidy

# ビルド
go build -o docsh main.go
```

### 代替: スクリプト/Makefile
- `./build.sh`（クロスビルドスクリプト。現状表示名に "CherryShell" が残っているため改善候補）
- `make local`（Makefile のローカルビルドターゲットはバイナリ名 `go-zsh` を出力。整合性の改善候補）

## 実行
```bash
# 対話モード
./docsh

# 直接コマンド
./docsh ps
./docsh images
./docsh run nginx
```

## ヒント
- `data/config.yaml` でプロンプトや挙動を調整可能
- 言語は `--lang` または環境変数で指定可能（例: `./docsh --lang ja`）
