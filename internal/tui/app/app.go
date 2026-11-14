package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/MoeidHeidari/guardrail/internal/tui/screens"
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func Run(ctx context.Context, contextPath string) (err error) {
	root, err := watcher.ResolveRoot(contextPath)
	if err != nil {
		return err
	}
	terminal, err := tui.NewApp()
	if err != nil {
		return fmt.Errorf("create terminal workspace: %w", err)
	}
	defer func() {
		if closeErr := terminal.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close terminal workspace: %w", closeErr))
		}
	}()
	theme := styles.Detect()
	files, err := watcher.New(root)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := files.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close filesystem watcher: %w", closeErr))
		}
	}()
	screen := screens.NewMain(theme, files.Root, files.Initial)
	live, reports, stopAnalysis := analysis.StartLive(ctx, files.Root)
	defer stopAnalysis()
	screen.BeginAnalysis(live.Submit(files.Initial))
	previews, stopPreviews := screen.StartPreviewWorker(ctx)
	defer stopPreviews()
	terminal.SetRootComponent(screen)

	if err := terminal.Open(); err != nil {
		return fmt.Errorf("open terminal workspace: %w", err)
	}
	frames := newFrameQueue()
	defer frames.timer.Stop()
	animation := time.NewTimer(time.Hour)
	animation.Stop()
	defer animation.Stop()
	var animationTick <-chan time.Time
	scheduleAnimation := func() {
		if screen.HasActivity() {
			if animationTick == nil {
				animation.Reset(200 * time.Millisecond)
				animationTick = animation.C
			}
		} else {
			animation.Stop()
			animationTick = nil
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	updates := make(chan watcher.Update, 8)
	done := make(chan struct{})
	var watchErr error
	go func() {
		watchErr = files.Run(ctx, updates)
		close(done)
	}()
	defer func() {
		cancel()
		<-done
		err = errors.Join(err, watchErr)
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-terminal.StopCh():
			return nil
		case <-done:
			return nil
		case <-frames.tick:
			if frames.take() {
				screen.InvalidateRender()
			}
			terminal.MarkDirty()
			terminal.Render()
		case <-animationTick:
			animationTick = nil
			if screen.AdvanceAnimation() {
				frames.request(false)
			}
			scheduleAnimation()
		case result := <-previews:
			if screen.ApplyPreview(result) {
				frames.request(true)
			}
			scheduleAnimation()
		case report := <-reports:
			if screen.ApplyAnalysis(report) {
				frames.request(true)
			}
		case update := <-updates:
			screen.Apply(update)
			screen.BeginAnalysis(live.Submit(update))
			frames.request(true)
			scheduleAnimation()
		case event := <-terminal.Events():
			if mouse, ok := event.(tui.MouseEvent); ok && mouse.Button == tui.MouseNone && mouse.Action != tui.MouseDrag {
				continue
			}
			terminal.Dispatch(event)
			frames.request(true)
			scheduleAnimation()
		}
	}
}
