package analysis

import (
	"context"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/MoeidHeidari/guardrail/internal/config"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
)

type LiveResult struct {
	Generation uint64
	Report     Report
	Rules      []config.Rule
	Error      string
}

type liveEntry struct {
	entry   watcher.Entry
	version uint64
}

type cachedResult struct {
	version uint64
	result  Result
}

type Live struct {
	mu         sync.Mutex
	root       string
	entries    map[string]liveEntry
	generation uint64
	rulesEpoch uint64
	wake       chan struct{}
	cancelJob  context.CancelFunc
}

func StartLive(ctx context.Context, root string) (*Live, <-chan LiveResult, func()) {
	ctx, cancel := context.WithCancel(ctx)
	live := &Live{root: root, entries: make(map[string]liveEntry), wake: make(chan struct{}, 1)}
	results := make(chan LiveResult, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		live.run(ctx, results)
	}()
	return live, results, func() { cancel(); <-done }
}

func (live *Live) Submit(update watcher.Update) uint64 {
	live.mu.Lock()
	defer live.mu.Unlock()
	live.generation++
	if live.cancelJob != nil {
		live.cancelJob()
	}
	if update.Entries != nil {
		clear(live.entries)
		live.rulesEpoch++
		for _, entry := range update.Entries {
			live.entries[entry.Path] = liveEntry{entry, live.generation}
		}
	}
	for _, path := range update.Removed {
		delete(live.entries, path)
	}
	for _, entry := range update.Upsert {
		live.entries[entry.Path] = liveEntry{entry, live.generation}
	}
	for _, change := range update.Changes {
		path := filepath.ToSlash(change.Path)
		if path == config.FileName || path == ".guardrail" || path == ".guard" ||
			strings.HasPrefix(path, ".guardrail/") || strings.HasPrefix(path, ".guard/") {
			live.rulesEpoch++
		}
		if entry, exists := live.entries[change.Path]; exists {
			entry.version = live.generation
			live.entries[change.Path] = entry
		}
	}
	select {
	case live.wake <- struct{}{}:
	default:
	}
	return live.generation
}

func (live *Live) snapshot(ctx context.Context) (context.Context, context.CancelFunc, uint64, uint64, map[string]liveEntry) {
	live.mu.Lock()
	defer live.mu.Unlock()
	job, cancel := context.WithCancel(ctx)
	live.cancelJob = cancel
	entries := make(map[string]liveEntry, len(live.entries))
	for path, entry := range live.entries {
		entries[path] = entry
	}
	return job, cancel, live.generation, live.rulesEpoch, entries
}

func (live *Live) run(ctx context.Context, results chan<- LiveResult) {
	cache := make(map[string]cachedResult)
	var engine *Engine
	var epoch uint64
	var loadError string
	for {
		select {
		case <-ctx.Done():
			return
		case <-live.wake:
		}
		job, cancel, generation, rulesEpoch, entries := live.snapshot(ctx)
		if rulesEpoch != epoch {
			clear(cache)
			epoch = rulesEpoch
			cfg, err := config.Load(live.root)
			engine = nil
			if err == nil {
				engine, err = New(live.root, cfg)
			}
			loadError = ""
			if err != nil {
				loadError = err.Error()
			}
		}
		result := LiveResult{Generation: generation, Error: loadError}
		if engine != nil {
			result.Rules = engine.Config.Rules
			for path := range cache {
				if _, exists := entries[path]; !exists {
					delete(cache, path)
				}
			}
			for path, entry := range entries {
				if job.Err() != nil {
					break
				}
				path = filepath.ToSlash(path)
				if entry.entry.IsDir() || !engine.Config.Applies(path) || !entry.entry.Mode.IsRegular() {
					result.Report.Skipped++
					continue
				}
				result.Report.Checked++
				cached, exists := cache[path]
				if !exists || cached.version != entry.version {
					content, err := Read(live.root, path)
					file := Result{Path: path}
					if err != nil {
						file.Findings = []Finding{ErrorFinding(path, "read", err)}
					} else {
						file = engine.Evaluate(job, path, content)
					}
					if job.Err() != nil {
						break
					}
					cached = cachedResult{entry.version, file}
					cache[path] = cached
				}
				result.Report.Files = append(result.Report.Files, cached.result)
			}
			sort.Slice(result.Report.Files, func(i, j int) bool {
				return result.Report.Files[i].Path < result.Report.Files[j].Path
			})
		}
		valid := job.Err() == nil
		cancel()
		if valid {
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}
