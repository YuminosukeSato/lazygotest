package config

import "time"

// Config はアプリ全体の設定値。
// How: CLI/環境変数/YAML を将来マージする前提で単純な構造体に集約。
// Why not: MVP では YAML 読み込みを省略し、CLI/既定のみ。
type Config struct {
	Watch      bool
	Cover      bool
	Race       bool
	Short      bool
	Timeout    time.Duration
	Tags       []string
	Editor     string
	EditorArgs string
	Parallel   int
	Debug      bool
	Packages   []string
}
