package analysis

import (
	"context"
	"github.com/MoeidHeidari/guardrail/internal/ignore"
	"io/fs"
	"path/filepath"
)

func (engine *Engine) Check(ctx context.Context, root string) (Report, error) {
	report := Report{}
	matcher, err := ignore.Load(root)
	if err != nil {
		return report, err
	}
	var paths []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if matcher.Matches(relative, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			report.Skipped++
			return nil
		}
		if entry.IsDir() {
			if relative == ".git" || relative == ".guardrail" || relative == ".guard" {
				return filepath.SkipDir
			}
			return nil
		}
		if !engine.Config.Applies(relative) || entry.Type()&fs.ModeSymlink != 0 {
			report.Skipped++
			return nil
		}
		report.Checked++
		paths = append(paths, relative)
		return nil
	})
	if err != nil {
		return report, err
	}
	results := make([]Result, len(paths))
	parallelFor(ctx, len(paths), func(ctx context.Context, index int) {
		path := paths[index]
		content, err := Read(root, path)
		if err != nil {
			results[index] = Result{Path: path, Findings: []Finding{ErrorFinding(path, "read", err)}}
			return
		}
		results[index] = engine.Evaluate(ctx, path, content)
	})
	if err := ctx.Err(); err != nil {
		return report, err
	}
	report.Files = results
	return report, nil
}
