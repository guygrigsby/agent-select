package routing

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

var markdownLink = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

func missingLocalLinks(docPath string, data []byte) []string {
	var missing []string
	for _, match := range markdownLink.FindAllSubmatch(data, -1) {
		u, err := url.Parse(string(match[1]))
		if err != nil {
			missing = append(missing, fmt.Sprintf("invalid link in %s", docPath))
			continue
		}
		if u.IsAbs() || u.Host != "" || u.Path == "" {
			continue
		}
		target := filepath.Join(filepath.Dir(docPath), filepath.FromSlash(u.Path))
		if _, err := os.Stat(target); err != nil {
			missing = append(missing, fmt.Sprintf("%s links to missing %s; restore the target or fix the link", docPath, target))
		}
	}
	return missing
}

func TestDocumentationLocalLinks(t *testing.T) {
	paths, err := filepath.Glob("../docs/specs/*.md")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, "../README.md")
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, problem := range missingLocalLinks(path, data) {
			t.Error(problem)
		}
	}
}

func TestDocumentationLinkGuardDetectsMissingTarget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "README.md")
	missing := missingLocalLinks(path, []byte("[missing](missing.md) [web](https://example.com/doc) [anchor](#here)"))
	if len(missing) != 1 {
		t.Fatalf("missing target mutation must produce one actionable problem, got %v", missing)
	}
}
