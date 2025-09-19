# lazygotest 配布 TODO

- [x] cmd/gotui ディレクトリを cmd/lazygotest へリネームし、`main.go` のバイナリ名表示を lazygotest に合わせる
- [x] internal/adapter/primary/tui/view.go のタイトル表記を lazygotest に統一する
- [x] README.md のコマンド例と説明をすべて lazygotest 名称に更新する
- [x] HOMEBREW_SETUP.md のインストール手順を lazygotest 名称に更新する
- [x] 新しいエントリポイントに合わせて go install や go build の例を検証し README に反映する
- [x] gofmt と golangci-lint を実行してフォーマットと静的解析を確認する
- [x] go test ./... を実行して挙動確認する
