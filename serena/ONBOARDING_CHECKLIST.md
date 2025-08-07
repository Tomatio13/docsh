# オンボーディングチェックリスト

- [ ] Go 1.22+ を使用できる環境を確認
- [ ] `go mod tidy` で依存解決
- [ ] `go build -o docsh main.go` でビルド
- [ ] `./docsh --lang ja` で対話モード起動
- [ ] `./docsh ps` / `./docsh images` など基本動作確認
- [ ] `mapping search ps` でマッピング検索の動作確認
- [ ] `tail -f <container>` → `docker logs -f` のストリーミング確認（Docker 必須）
- [ ] 初期タスクのうち「高」優先度を 1 つ以上完了
