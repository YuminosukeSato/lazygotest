package discovery

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"regexp"
	"strings"
)

// Package は go list -json の必要最小限の情報。
type Package struct {
	ImportPath string `json:"ImportPath"`
	Dir        string `json:"Dir"`
}

// ListPackages はパターン（例: ./...）に一致するパッケージを列挙する。
func ListPackages(ctx context.Context, patterns []string) ([]Package, error) {
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}
	args := append([]string{"list", "-json"}, patterns...)
	cmd := exec.CommandContext(ctx, "go", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(out))
	var pkgs []Package
	for dec.More() {
		var p Package
		if err := dec.Decode(&p); err != nil {
			return nil, err
		}
		pkgs = append(pkgs, p)
	}
	return pkgs, nil
}

var testNameRe = regexp.MustCompile(`^(Test|Example|Fuzz)\w+$`)

// ListTests は指定パッケージのテスト関数名を列挙する。
func ListTests(ctx context.Context, importPath string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "go", "test", "-list", ".", importPath)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	names := make([]string, 0, 16)
	s := bufio.NewScanner(out)
	for s.Scan() {
		ln := strings.TrimSpace(s.Text())
		if testNameRe.MatchString(ln) {
			names = append(names, ln)
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return names, nil
}
