package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
)

const Folder = ".guard"

func Load(root string) (result Config, err error) {
	dir, err := os.OpenRoot(root)
	if err != nil {
		return Config{}, err
	}
	defer func() { err = errors.Join(err, dir.Close()) }()
	info, err := dir.Lstat(Folder)
	if err != nil {
		return Config{}, fmt.Errorf("load rule sets from .guard: %w", err)
	}
	if !info.IsDir() {
		return Config{}, fmt.Errorf(".guard must be a non-symlink directory")
	}
	combined := Config{Version: 1, TimeoutMS: 100}
	ids := make(map[string]bool)
	err = fs.WalkDir(dir.FS(), Folder, func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		extension := strings.ToLower(path.Ext(file))
		if extension != ".yaml" && extension != ".yml" {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("rule-set YAML must be a regular non-symlink file: %s", file)
		}
		cfg, err := loadFile(root, file)
		if err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		set := path.Base(path.Dir(file))
		if path.Dir(file) == Folder {
			set = "General"
		}
		for i := range cfg.Rules {
			rule := &cfg.Rules[i]
			key := set + "/" + rule.ID
			if ids[key] {
				return fmt.Errorf("duplicate rule %s in rule set %s", rule.ID, set)
			}
			ids[key] = true
			rule.Set, rule.Source = set, file
			rule.ScopeInclude, rule.ScopeExclude, rule.TimeoutMS = cfg.Include, cfg.Exclude, cfg.TimeoutMS
		}
		combined.Sets = append(combined.Sets, cfg)
		combined.Rules = append(combined.Rules, cfg.Rules...)
		return nil
	})
	if err != nil {
		return Config{}, err
	}
	if len(combined.Rules) == 0 {
		return Config{}, fmt.Errorf("no YAML rule sets found in .guard")
	}
	return combined, nil
}
