package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitIgnoreSemantics(t *testing.T) {
	root := t.TempDir()
	data := "# comment\n*.log\n!important.log\n/root.txt\ncache/\nignored/*\n!ignored/keep.txt\nexcluded/\n!excluded/keep.txt\n**/generated/*.go\n\\#literal\n"
	if err := os.WriteFile(filepath.Join(root, FileName), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	matcher, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		path               string
		directory, ignored bool
	}{
		{"a.log", false, true}, {"sub/a.log", false, true}, {"important.log", false, false},
		{"root.txt", false, true}, {"sub/root.txt", false, false}, {"cache", true, true},
		{"cache", false, false}, {"cache/a.txt", false, true},
		{"ignored/drop.txt", false, true}, {"ignored/keep.txt", false, false},
		{"excluded/keep.txt", false, true}, {"a/generated/a.go", false, true},
		{"#literal", false, true}, {FileName, false, false}, {".", true, false},
	} {
		if got := matcher.Matches(test.path, test.directory); got != test.ignored {
			t.Errorf("%s: got %v want %v", test.path, got, test.ignored)
		}
	}
	if _, err := Load(t.TempDir()); err != nil {
		t.Fatal(err)
	}
}
