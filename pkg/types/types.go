package types

// TestID は (パッケージ, テスト名) を一意に識別する。
type TestID struct {
	Pkg  string
	Name string
}

type TestStatus int

const (
	StatusUnknown TestStatus = iota
	StatusRunning
	StatusPassed
	StatusFailed
	StatusSkipped
)
