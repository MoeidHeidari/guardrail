package analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MoeidHeidari/guardrail/internal/watcher"
)

func awaitLive(t *testing.T, results <-chan LiveResult, generation uint64) LiveResult {
	t.Helper()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for {
		select {
		case result := <-results:
			if result.Generation == generation {
				return result
			}
		case <-timer.C:
			t.Fatal("live analysis did not finish")
			return LiveResult{}
		}
	}
}

func TestLiveAnalysisUpdatesDeletionAndReload(t *testing.T) {
	root := t.TempDir()
	write := func(path, text string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, path), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, ".guardrail"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".guard"), 0700); err != nil {
		t.Fatal(err)
	}
	write(".guardrail/rule.js", `(path, content) => content.includes("bad")`)
	write(".guard/rules.yml", "version: 1\nrules:\n  - id: js\n    type: javascript\n    file: .guardrail/rule.js\n")
	write("a.js", "const bad = 1;")
	files, err := watcher.New(root)
	if err != nil {
		t.Fatal(err)
	}
	initial := files.Initial
	if err := files.Close(); err != nil {
		t.Fatal(err)
	}
	live, results, stop := StartLive(context.Background(), root)
	defer stop()
	result := awaitLive(t, results, live.Submit(initial))
	if result.Error != "" || result.Report.Checked != 1 || result.Report.ExitCode() != 1 {
		t.Fatalf("startup results incorrect: %+v", result)
	}
	write("a.js", "const good = 1;")
	result = awaitLive(t, results, live.Submit(watcher.Update{Upsert: []watcher.Entry{{Path: "a.js"}},
		Changes: []watcher.Change{{Path: "a.js", Kind: "modified"}}}))
	if result.Report.ExitCode() != 0 {
		t.Fatalf("resolved violation was retained: %+v", result)
	}
	write(".guardrail/rule.js", `() => true`)
	result = awaitLive(t, results, live.Submit(watcher.Update{Changes: []watcher.Change{{Path: ".guardrail/rule.js", Kind: "modified"}}}))
	if result.Report.ExitCode() != 1 {
		t.Fatal("script changes did not invalidate analysis")
	}
	write(".guard/rules.yml", "unknown: true")
	result = awaitLive(t, results, live.Submit(watcher.Update{Changes: []watcher.Change{{Path: ".guard/rules.yml", Kind: "modified"}}}))
	if result.Error == "" || len(result.Rules) != 0 {
		t.Fatal("invalid configuration was not visible")
	}
	write(".guard/rules.yml", "version: 1\nrules:\n  - id: js\n    type: javascript\n    file: .guardrail/rule.js\n")
	result = awaitLive(t, results, live.Submit(watcher.Update{Removed: []string{"a.js"},
		Changes: []watcher.Change{{Path: ".guard/rules.yml", Kind: "modified"}}}))
	if result.Error != "" || result.Report.Checked != 0 || len(result.Report.Findings()) != 0 {
		t.Fatalf("deleted file retained findings: %+v", result)
	}
}

func TestLiveCoalescesRapidSaves(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".guard"), 0700); err != nil {
		t.Fatal(err)
	}

	for path, text := range map[string]string{
		".guard/rules.yml": "version: 1\ninclude: ['**/*.txt']\nrules:\n  - id: bad\n    type: regex\n    pattern: bad\n",
		"a.txt":            "bad",
	} {
		if err := os.WriteFile(filepath.Join(root, path), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	live, results, stop := StartLive(context.Background(), root)
	defer stop()
	live.Submit(watcher.Update{Entries: []watcher.Entry{{Path: "a.txt"}}})
	var generation uint64
	for range 100 {
		generation = live.Submit(watcher.Update{Upsert: []watcher.Entry{{Path: "a.txt"}}})
	}
	result := awaitLive(t, results, generation)
	if result.Report.Checked != 1 || result.Report.ExitCode() != 1 {
		t.Fatalf("latest snapshot missing: %+v", result)
	}
}

func TestLiveDiscoversNewRuleSet(t *testing.T) {
	root := t.TempDir()
	write := func(path, text string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, path)), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(".guard/rules.yml", "version: 1\nrules:\n  - id: match\n    type: regex\n    pattern: NEVER\n")
	write("a.go", "package main\n// TODO")
	live, reports, stop := StartLive(context.Background(), root)
	defer stop()
	result := awaitLive(t, reports, live.Submit(watcher.Update{Entries: []watcher.Entry{{Path: "a.go"}}}))
	if len(result.Report.Findings()) != 0 {
		t.Fatal("initial set should pass")
	}
	write(".guard/quality/todo.yaml", "version: 1\nrules:\n  - id: match\n    type: regex\n    pattern: TODO\n")
	result = awaitLive(t, reports, live.Submit(watcher.Update{Changes: []watcher.Change{{Path: ".guard/quality/todo.yaml", Kind: "created"}}}))
	findings := result.Report.Findings()
	if len(findings) != 1 || findings[0].Set != "quality" || findings[0].Source != ".guard/quality/todo.yaml" {
		t.Fatalf("new rule set not applied or attributed: %+v", result)
	}
	if err := os.Remove(filepath.Join(root, ".guard/quality/todo.yaml")); err != nil {
		t.Fatal(err)
	}
	result = awaitLive(t, reports, live.Submit(watcher.Update{Changes: []watcher.Change{{Path: ".guard/quality/todo.yaml", Kind: "deleted"}}}))
	if len(result.Report.Findings()) != 0 || len(result.Rules) != 1 {
		t.Fatal("deleted set retained findings")
	}
}
