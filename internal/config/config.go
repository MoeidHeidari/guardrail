package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"
)

const FileName = ".guard/rules.yml"

type Rule struct {
	Description  string   `yaml:"description" json:"description"`
	Set          string   `yaml:"-" json:"set"`
	Source       string   `yaml:"-" json:"source"`
	ScopeInclude []string `yaml:"-" json:"-"`
	ScopeExclude []string `yaml:"-" json:"-"`
	TimeoutMS    int      `yaml:"-" json:"-"`
	ID           string   `yaml:"id" json:"id"`
	Message      string   `yaml:"message" json:"message"`
	Type         string   `yaml:"type" json:"type"`
	Include      []string `yaml:"include" json:"include"`
	Pattern      string   `yaml:"pattern" json:"-"`
	Code         string   `yaml:"code" json:"-"`
	File         string   `yaml:"file" json:"-"`
}

type Config struct {
	Sets      []Config `yaml:"-"`
	Version   int      `yaml:"version"`
	Include   []string `yaml:"include"`
	Exclude   []string `yaml:"exclude"`
	TimeoutMS int      `yaml:"timeout_ms"`
	Rules     []Rule   `yaml:"rules"`
}

var DefaultInclude = []string{
	"**/*.go", "**/*.js", "**/*.jsx", "**/*.mjs", "**/*.cjs",
	"**/*.ts", "**/*.tsx", "**/*.py", "**/*.rs",
	"**/*.json", "**/*.yaml", "**/*.yml", "**/*.toml", "**/*.sh",
	"**/*.css", "**/*.html", "**/Dockerfile", "**/Makefile",
}

var DefaultExclude = []string{
	".git/**", ".guardrail/**", ".guard/**", FileName, "**/node_modules/**",
	"**/vendor/**", "**/.venv/**", "**/venv/**", "**/__pycache__/**",
	"**/dist/**", "**/build/**", "**/target/**", "**/bin/**",
}

func loadFile(root, path string) (cfg Config, err error) {
	dir, err := os.OpenRoot(root)
	if err != nil {
		return cfg, err
	}
	defer func() { err = errors.Join(err, dir.Close()) }()
	file, err := dir.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("load %s: %w", path, err)
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	data, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err != nil {
		return cfg, err
	}
	if len(data) > 1<<20 {
		return cfg, errors.New("configuration exceeds 1 MiB")
	}
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return cfg, fmt.Errorf("%s must contain exactly one YAML document", path)
	}
	if cfg.Version != 1 {
		return cfg, errors.New("configuration version must be 1")
	}
	if cfg.TimeoutMS == 0 {
		cfg.TimeoutMS = 100
	}
	if cfg.TimeoutMS < 1 || cfg.TimeoutMS > 10000 {
		return cfg, errors.New("timeout_ms must be between 1 and 10000")
	}
	if cfg.Include == nil {
		cfg.Include = append([]string(nil), DefaultInclude...)
	}
	if cfg.Exclude == nil {
		cfg.Exclude = append([]string(nil), DefaultExclude...)
	}
	if len(cfg.Rules) == 0 {
		return cfg, errors.New("configuration must define at least one rule")
	}
	ids := make(map[string]bool)
	for _, rule := range cfg.Rules {
		if strings.TrimSpace(rule.ID) == "" || ids[rule.ID] {
			return cfg, fmt.Errorf("rule id must be nonempty and unique: %q", rule.ID)
		}
		ids[rule.ID] = true
		switch rule.Type {
		case "regex":
			if rule.Pattern == "" || rule.Code != "" || rule.File != "" {
				return cfg, fmt.Errorf("regex rule %s requires only pattern", rule.ID)
			}
		case "javascript":
			if (rule.Code == "") == (rule.File == "") || rule.Pattern != "" {
				return cfg, fmt.Errorf("JavaScript rule %s requires either code or file", rule.ID)
			}
		default:
			return cfg, fmt.Errorf("rule %s has unsupported type %q", rule.ID, rule.Type)
		}
	}
	patterns := append(append([]string(nil), cfg.Include...), cfg.Exclude...)
	for _, rule := range cfg.Rules {
		patterns = append(patterns, rule.Include...)
	}
	for _, pattern := range patterns {
		if !doublestar.ValidatePattern(pattern) {
			return cfg, fmt.Errorf("invalid glob pattern %q", pattern)
		}
	}
	return cfg, nil
}

func Matches(patterns []string, path string) bool {
	for _, pattern := range patterns {
		if doublestar.MatchUnvalidated(pattern, path) {
			return true
		}
	}
	return false
}

func (cfg Config) Applies(path string) bool {
	if len(cfg.Sets) > 0 {
		for _, set := range cfg.Sets {
			if set.Applies(path) {
				return true
			}
		}
		return false
	}
	return !strings.HasPrefix(path, ".git/") && !strings.HasPrefix(path, ".guardrail/") &&
		!strings.HasPrefix(path, ".guard/") &&
		path != FileName && Matches(cfg.Include, path) && !Matches(cfg.Exclude, path)
}
