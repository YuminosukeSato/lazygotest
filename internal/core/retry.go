package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// FailSet は直近の失敗テスト集合。
type FailSet struct {
	// map[pkg]set[testName]
	Items map[string]map[string]struct{} `json:"items"`
}

func NewFailSet() *FailSet { return &FailSet{Items: map[string]map[string]struct{}{}} }

func (f *FailSet) Add(pkg, name string) {
	m, ok := f.Items[pkg]
	if !ok {
		m = map[string]struct{}{}
		f.Items[pkg] = m
	}
	m[name] = struct{}{}
}

func (f *FailSet) NamesByPkg(pkg string) []string {
	m := f.Items[pkg]
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func (f *FailSet) Save() error {
	dir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "lazygotest", "last_failed.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func LoadFailSet() (*FailSet, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "lazygotest", "last_failed.json")
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return NewFailSet(), nil
	}
	if err != nil {
		return nil, err
	}
	var f FailSet
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	if f.Items == nil {
		f.Items = map[string]map[string]struct{}{}
	}
	return &f, nil
}
