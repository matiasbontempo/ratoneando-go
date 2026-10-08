package scrapers

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Every scraper file declares the source name it reports. A copy-pasted file once kept
// "disco" as its source (Farmacity), so its products were shown, and recorded in the price
// history, as Disco's. Each file must report exactly one source, named like the file, and no
// two files may report the same one.
func TestEveryScraperReportsItsOwnSource(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	sourceLiteral := regexp.MustCompile(`Source:\s+"([a-z]+)"`)
	seen := map[string]string{}

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		matches := sourceLiteral.FindAllStringSubmatch(string(content), -1)
		if len(matches) == 0 {
			continue
		}

		expected := strings.TrimSuffix(file, ".go")
		for _, match := range matches {
			source := match[1]
			if source != expected {
				t.Errorf("%s reports source %q, expected %q", file, source, expected)
			}
			if other, taken := seen[source]; taken && other != file {
				t.Errorf("source %q is reported by both %s and %s", source, other, file)
			}
			seen[source] = file
		}
	}

	if len(seen) < 6 {
		t.Fatalf("expected to find the scraper sources, found only %v", seen)
	}
}
