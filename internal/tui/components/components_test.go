package components_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

func TestContainerOptions(t *testing.T) {
	theme := styles.Default()
	container := components.NewContainer(theme,
		tui.WithDirection(tui.Row),
		tui.WithGap(2),
		tui.WithWidth(40),
	)
	if container.Style().Direction != tui.Row || container.Style().Gap != 2 {
		t.Fatal("container did not apply layout options")
	}
	if !container.Background().Bg.Equal(theme.Surface) {
		t.Fatal("container did not apply theme background")
	}
	if !container.TextStyle().Fg.Equal(theme.Foreground) {
		t.Fatal("container did not apply theme foreground")
	}
	container.Calculate(80, 24)
	if container.Rect().Width != 40 {
		t.Fatalf("width = %d, want 40", container.Rect().Width)
	}
}

func TestNestedSectionsInheritTheme(t *testing.T) {
	theme := styles.Default()
	theme.Foreground = tui.RGBColor(255, 128, 64)
	theme.Surface = tui.RGBColor(20, 30, 40)
	label := tui.New(tui.WithText("Ready"))
	section := components.NewSection(theme, "Diagnostics", label)
	container := components.NewContainer(theme, tui.WithPadding(1))
	container.AddChild(section)
	buffer := tui.NewBuffer(40, 12)
	container.RenderTo(buffer, 40, 12)

	if !strings.Contains(buffer.StringTrimmed(), "Diagnostics") {
		t.Fatalf("missing section title:\n%s", buffer.StringTrimmed())
	}
	if label.Parent() != section || section.Parent() != container {
		t.Fatal("section did not preserve child composition")
	}
	if !section.Border().All(tui.BorderNone) {
		t.Fatal("section should not have borders")
	}
	rect := label.Rect()
	cell := buffer.Cell(rect.X, rect.Y)
	if cell.Rune != 'R' || !cell.Style.Fg.Equal(theme.Foreground) || !cell.Style.Bg.Equal(theme.Surface) {
		t.Fatalf("label did not inherit theme: cell=%+v label=%+v section=%+v\n%s", cell, rect, section.Rect(), buffer.StringTrimmed())
	}
}

func TestLayoutKeepsSidebarWidthWithLongContent(t *testing.T) {
	theme := styles.Default()
	sidebar := components.NewSidebar(theme)
	content := components.NewSection(theme, "Overview")
	root := components.NewLayout(theme,
		components.NewHeader(theme, "GUARDRAIL", "q / ctrl+c quit"),
		sidebar,
		content,
	)
	root.Calculate(80, 24)
	sidebarWidth, contentWidth := sidebar.Rect().Width, content.Rect().Width
	content.AddChild(tui.New(tui.WithText(strings.Repeat("content ", 100))))
	root.Calculate(80, 24)
	if sidebar.Rect().Width != sidebarWidth {
		t.Fatalf("sidebar width = %d, want %d", sidebar.Rect().Width, sidebarWidth)
	}
	if content.Rect().Width != contentWidth {
		t.Fatalf("content width = %d, want %d", content.Rect().Width, contentWidth)
	}
}

func TestLayoutResizes(t *testing.T) {
	theme := styles.Default()
	header := components.NewHeader(theme, "GUARDRAIL", "q / ctrl+c quit")
	sidebar := components.NewSidebar(theme)
	content := components.NewSection(theme, "Overview")
	root := components.NewLayout(theme, header, sidebar, content)

	for _, size := range []struct{ width, height int }{
		{120, 40}, {80, 24}, {40, 12}, {20, 8}, {80, 24},
	} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			root.MarkDirty()
			root.Calculate(size.width, size.height)
			h, s, c := header.Rect(), sidebar.Rect(), content.Rect()
			if h.Height != 3 {
				t.Errorf("header height = %d, want 3", h.Height)
			}
			if s.Y != h.Y+h.Height+1 || s.Y != c.Y || s.Height != c.Height {
				t.Errorf("body is not aligned below header: header=%+v sidebar=%+v content=%+v", h, s, c)
			}
			if c.X != s.X+s.Width+1 || c.Width <= 0 || s.Width <= 0 {
				t.Errorf("body columns overlap or collapse: sidebar=%+v content=%+v", s, c)
			}
			if s.Width != (h.Width-1)/4 {
				t.Errorf("sidebar width = %d, want %d", s.Width, (h.Width-1)/4)
			}
			if c.X+c.Width != h.X+h.Width {
				t.Errorf("content right edge = %d, want header right edge %d", c.X+c.Width, h.X+h.Width)
			}
			bounds := tui.Rect{Width: size.width, Height: size.height}
			for _, element := range []*tui.Element{header, sidebar, content} {
				if !bounds.ContainsRect(element.Rect()) {
					t.Errorf("element exceeds terminal: %+v in %+v", element.Rect(), bounds)
				}
			}
			buffer := tui.NewBuffer(size.width, size.height)
			root.RenderTo(buffer, size.width, size.height)
			for x := h.X; x < h.X+h.Width; x++ {
				if !buffer.Cell(x, h.Y+h.Height).Style.Bg.Equal(theme.Background) {
					t.Fatal("sections painted over the gap below the header")
				}
			}
			for y := s.Y; y < s.Y+s.Height; y++ {
				if !buffer.Cell(s.X+s.Width, y).Style.Bg.Equal(theme.Background) {
					t.Fatal("sections painted over the sidebar-content gap")
				}
			}
		})
	}
}
