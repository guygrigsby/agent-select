package buildgate

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Exercise the real Makefile against an isolated module, including source
// outside the historical internal/ and cmd/ directories.
func TestLintCoversRouting(t *testing.T) {
	makefile, err := os.ReadFile(filepath.Join("..", "..", "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		source string
		valid  bool
	}{
		{"formatted", "package routing\n\nvar value = 1\n", true},
		{"unformatted", "package routing\n\nvar value=1\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for _, pkg := range []string{"internal", "cmd", "routing"} {
				if err := os.Mkdir(filepath.Join(dir, pkg), 0700); err != nil {
					t.Fatal(err)
				}
			}
			for name, data := range map[string][]byte{
				"Makefile":         makefile,
				"go.mod":           []byte("module lintfixture\n\ngo 1.22\n"),
				"routing/route.go": []byte(tc.source),
			} {
				if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command("make", "lint")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local", "GOPROXY=off")
			out, err := cmd.CombinedOutput()
			if (err == nil) != tc.valid {
				t.Fatalf("make lint: %v\n%s\nremedy: Makefile lint must scan every Go source directory, including routing/", err, out)
			}
		})
	}
}
