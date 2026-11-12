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
		result := Result{Path: relative}
		content, err := Read(root, relative)
		if err != nil {
			result.Findings = []Finding{ErrorFinding(relative, "read", err)}
		} else {
			result = engine.Evaluate(ctx, relative, content)
		}
		report.Files = append(report.Files, result)
		return nil
	})
	return report, err
}
