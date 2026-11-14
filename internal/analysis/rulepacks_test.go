package analysis

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/config"
)

func communityPackEngine(t *testing.T, packs ...string) *Engine {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join("..", "..", "rulepacks")
	for _, pack := range packs {
		installCommunityPack(t, root, source, pack)
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := New(root, functionalTestConfig(cfg))
	if err != nil {
		t.Fatal(err)
	}
	return engine
}

func installCommunityPack(t *testing.T, root, source, pack string) {
	t.Helper()
	functionCategory := pack
	if pack == "no-comments" {
		functionCategory = "source-policy"
	}
	destination := filepath.Join(root, ".guard", "functions", functionCategory)
	if err := os.MkdirAll(destination, 0700); err != nil {
		t.Fatal(err)
	}
	sets, err := filepath.Glob(filepath.Join(source, "*", pack+".yml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sets) == 0 {
		t.Fatalf("no language rule sets for %s", pack)
	}
	for _, set := range sets {
		language := filepath.Base(filepath.Dir(set))
		target := filepath.Join(root, ".guard", language)
		if err := os.MkdirAll(target, 0700); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(set)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(target, pack+".yml"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(filepath.Join(source, "functions", functionCategory))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(source, "functions", functionCategory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(destination, entry.Name()), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCommunityRulePacks(t *testing.T) {
	t.Run("repository-hygiene", func(t *testing.T) {
		engine := communityPackEngine(t, "repository-hygiene")
		cases := []struct {
			name, content string
			want          []Finding
		}{
			{"empty", "", nil},
			{"clean", "hello\n", nil},
			{"crlf", "hello\r\nworld\r\n", nil},
			{"unicode-spaces", "é😀 \t\r\n", []Finding{{Rule: "no-trailing-whitespace", Line: 1, Column: 3}}},
			{"missing-newline", "ok\né😀", []Finding{{Rule: "require-final-newline", Line: 2, Column: 3}}},
			{"blank-spaces", " \t\n", []Finding{{Rule: "no-trailing-whitespace", Line: 1, Column: 1}}},
			{"not-marker", "prefix <<<<<<< ours\n========\n", nil},
			{"conflicts", "<<<<<<< ours\n||||||| base\n=======\n>>>>>>> theirs\n", []Finding{
				{Rule: "no-conflict-markers", Line: 1, Column: 1},
				{Rule: "no-conflict-markers", Line: 2, Column: 1},
				{Rule: "no-conflict-markers", Line: 3, Column: 1},
				{Rule: "no-conflict-markers", Line: 4, Column: 1},
			}},
		}
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				result := engine.Evaluate(context.Background(), "sample.txt", []byte(test.content))
				if len(result.Findings) != len(test.want) {
					t.Fatalf("unexpected findings: %+v", result.Findings)
				}
				for i, want := range test.want {
					got := result.Findings[i]
					if got.Kind != "violation" || got.Set != "general" ||
						got.Rule != want.Rule || got.Line != want.Line || got.Column != want.Column || got.Message == "" {
						t.Fatalf("finding %d: got %+v, expected %+v", i, got, want)
					}
				}
			})
		}
	})
	t.Run("maintainability", func(t *testing.T) {
		engine := communityPackEngine(t, "maintainability")
		for _, ending := range []string{"\n", "\r\n"} {
			for _, count := range []int{0, 299, 300, 301} {
				for _, final := range []bool{false, true} {
					content := strings.Repeat("line"+ending, count)
					if !final {
						content = strings.TrimSuffix(content, ending)
					}
					result := engine.Evaluate(context.Background(), "sample.txt", []byte(content))
					if count <= 300 {
						if len(result.Findings) != 0 {
							t.Fatalf("%d lines must pass: %+v", count, result.Findings)
						}
					} else if len(result.Findings) != 1 {
						t.Fatalf("expected one length finding: %+v", result.Findings)
					} else {
						got := result.Findings[0]
						if got.Kind != "violation" || got.Set != "general" || got.Rule != "max-file-lines" ||
							got.Line != 301 || got.Column != 1 || got.Message != "This file has 301 lines; the limit is 300." {
							t.Fatalf("incorrect length finding: %+v", got)
						}
					}
				}
			}
		}
	})
}
