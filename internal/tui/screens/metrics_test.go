package screens

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui"
)

func TestMetricCardsShowCountsAndMeaningfulRatios(t *testing.T) {
	m := overviewFixture(t)
	m.renderWidth = 160
	m.analysisResult.Report.Checked = 4
	for i, expected := range []struct{ value, detail string }{
		{"1", "1/2 of all findings"},
		{"1", "1/2 of all findings"},
		{"1", "1/4 of checked files"},
		{"1", "1/1 with findings"},
	} {
		card := m.metricPanels().Children()[i]
		if card.Children()[1].Text() != expected.value || card.Children()[3].Text() != expected.detail {
			t.Fatalf("card %d has incorrect count or denominator", i)
		}
	}
	for _, page := range []workspacePage{overviewPage, violationsPage} {
		m.openWorkspace(page)
		frame := render(m, 200, 40)
		for _, text := range []string{"1/2 of all findings", "1/4 of checked files", "1/1 with findings"} {
			if !strings.Contains(frame, text) {
				t.Fatalf("missing %q on page %v:\n%s", text, page, frame)
			}
		}
	}
}

func TestMetricBarFillAndTrack(t *testing.T) {
	m := testScreen()
	for _, test := range []struct{ count, total, filled int }{
		{0, 0, 0}, {0, 5, 0}, {1, 4, 5}, {4, 4, 20}, {8, 4, 20},
	} {
		spans := m.metricBar(test.count, test.total, 20, m.theme.Accent).RichText()
		if tui.StringWidth(spans[0].Text) != test.filled ||
			tui.StringWidth(spans[1].Text) != 20-test.filled {
			t.Fatalf("incorrect bar fill for %d/%d: %v", test.count, test.total, spans)
		}
		if !spans[1].Style.Fg.Equal(m.theme.Separator) {
			t.Fatal("empty track must use the muted separator color")
		}
	}
}

func TestMetricsDistinguishUnavailableFromZero(t *testing.T) {
	m := testScreen()
	for _, card := range m.metricPanels().Children() {
		if card.Children()[1].Text() != "--" {
			t.Fatal("uninitialized analysis must not appear successful")
		}
	}
	m.analysisGeneration = 1
	if value := m.metricPanels().Children()[0].Children()[1].Text(); value != "0" {
		t.Fatalf("completed empty analysis must show zero, got %q", value)
	}
	m.analysisPending = true
	if detail := m.metricPanels().Children()[0].Children()[3].Text(); detail != "Updating..." {
		t.Fatal("pending analysis must mark results as updating")
	}
	m.analysisResult.Error = "invalid configuration"
	if value := m.metricPanels().Children()[0].Children()[1].Text(); value != "--" {
		t.Fatal("failed analysis must not display success-shaped counts")
	}
}
