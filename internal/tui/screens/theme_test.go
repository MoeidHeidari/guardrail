package screens

import (
	"context"
	"reflect"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
)

func TestPreviewUsesSelectedTheme(t *testing.T) {
	code := "package main\nfunc main() { println(\"hello\") }\n"
	dark, _, err := highlightWithStyle("main.go", code, styles.Default().Syntax)
	if err != nil {
		t.Fatal(err)
	}
	light, _, err := highlightWithStyle("main.go", code, styles.Light().Syntax)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(dark, light) {
		t.Fatal("light and dark syntax colors must differ")
	}
	for _, async := range []bool{false, true} {
		m := fileScreen(t, code)
		m.theme = styles.Light()
		m.contentTab = true
		if async {
			results, stop := m.StartPreviewWorker(context.Background())
			render(m, 120, 32)
			result := waitPreview(t, results)
			stop()
			if !m.ApplyPreview(result) {
				t.Fatal("light theme preview result rejected")
			}
		} else {
			m.loadPreview()
		}
		if !reflect.DeepEqual(m.preview.lines, light) {
			t.Fatalf("async=%v: preview did not use light syntax palette", async)
		}
		if previewText(m.preview.lines) != code {
			t.Fatal("theme must not change file content")
		}
	}
}
