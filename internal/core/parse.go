package core

import "time"

// TestEvent は `go test -json` の 1 行をデコードした構造体。
// How: 標準の JSON デコーダで逐次読み出すだけで良い。
// Why not: 互換性を保つため、フィールドは必要最小限に留める。
type TestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"` // run|pass|fail|output|skip|bench|pause|cont
	Package string    `json:"Package"`
	Test    string    `json:"Test,omitempty"`
	Elapsed float64   `json:"Elapsed,omitempty"`
	Output  string    `json:"Output,omitempty"`
}
