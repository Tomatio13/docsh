# 検索ガイド（実用レシピ）

Serena での探索時にそのまま使える質問例。まずは広く、次に絞っていくのがコツ。

## 全体/エントリ
- 「認証や設定の初期化はどこで行われますか？」
- 「対話モードと直接実行はどのように切り替えていますか？」

## パース/実行
- 「コマンドのオプションはどのように解釈・保持されますか？」
- 「Linux コマンドが未対応のとき、どうエラーを返しますか？」
- 「`ps` コマンド出力整形はどこで実装されていますか？」

## マッピング
- 「Linux → Docker のマッピングはどこに定義され、どう検索されますか？」
- 「`tail -f` を `docker logs -f` に結びつける処理はどこですか？」

## I18n/設定
- 「言語はどのように決定されますか？優先順位は？」
- 「エイリアスはどの層で展開されますか？」

## Grep 向け正確検索（例）
- `func NewCommandParser\(`
- `type ParsedCommand struct`
- `IsDockerAvailable\(`
- `formatSimplePsOutput\(`
- `ExecuteWithMappingAndOptions\(`
