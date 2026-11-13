package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestLargeContextTerminalInput(t *testing.T) {
	root := t.TempDir()
	for i := range 1400 {
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("file%04d.txt", i)), []byte("hello\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range []struct {
		name, colorFGBG string
	}{
		{"dark", "15;0"},
		{"light", "0;15"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "/usr/bin/script", "-q", "/dev/null", os.Args[0],
				"-test.run=^TestLargeContextTerminalHelper$", "-test.v")
			cmd.Env = append(os.Environ(), "TERM=screen-256color", "COLORTERM=truecolor", "CI=",
				"COLORFGBG="+test.colorFGBG, "GUARDRAIL_TEST_CONTEXT="+root)
			input, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			result := make(chan error, 1)
			go func() {
				time.Sleep(2 * time.Second)
				_, err := input.Write([]byte("q"))
				result <- err
			}()
			output, err := cmd.CombinedOutput()
			writeErr := <-result
			if err != nil || writeErr != nil || !strings.Contains(string(output), "PASS") {
				t.Fatalf("large context must accept keyboard quit without panic: run=%v input=%v\n%s", err, writeErr, output)
			}
			if !strings.Contains(string(output), "GUARDRAIL") ||
				strings.Contains(string(output), "48;2;8;20;33") ||
				strings.Contains(string(output), "48;2;250;251;253") {
				t.Fatalf("startup must render without forcing the old %s background:\n%s", test.name, output)
			}
		})
	}
}

func TestLargeContextTerminalHelper(t *testing.T) {
	root := os.Getenv("GUARDRAIL_TEST_CONTEXT")
	if root == "" {
		t.Skip("subprocess helper")
	}
	if err := unix.IoctlSetWinsize(int(os.Stdout.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Row: 32, Col: 120}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := Run(ctx, root); err != nil {
		t.Fatalf("terminal run: %v", err)
	}
}
