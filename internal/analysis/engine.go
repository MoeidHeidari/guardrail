package analysis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/MoeidHeidari/guardrail/internal/config"
	"github.com/dop251/goja"
)

const ContentLimit = 4 << 20
const FindingLimit = 1000

type compiledRule struct {
	spec    config.Rule
	regex   *regexp.Regexp
	program *goja.Program
}

type Engine struct {
	Config config.Config
	rules  []compiledRule
}

func New(root string, cfg config.Config) (*Engine, error) {
	engine := &Engine{Config: cfg}
	for _, spec := range cfg.Rules {
		rule := compiledRule{spec: spec}
		var err error
		if spec.Type == "regex" {
			rule.regex, err = regexp.Compile(spec.Pattern)
		} else {
			code := spec.Code
			if spec.File != "" {
				path := filepath.ToSlash(filepath.Clean(spec.File))
				folder := strings.SplitN(path, "/", 2)[0]
				if (folder != ".guardrail" && folder != ".guard") || !strings.Contains(path, "/") || filepath.IsAbs(spec.File) {
					return nil, fmt.Errorf("rule %s: file must be inside .guard or .guardrail", spec.ID)
				}
				var content []byte
				content, err = readScript(root, folder, strings.TrimPrefix(path, folder+"/"))
				if err == nil {
					code = string(content)
				}
			}
			if err == nil {
				if len(code) > 65536 {
					return nil, fmt.Errorf("rule %s: JavaScript exceeds 64 KiB", spec.ID)
				}
				rule.program, err = goja.Compile(spec.ID, "("+strings.TrimSpace(code)+"\n)", true)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("compile rule %s: %w", spec.ID, err)
		}
		engine.rules = append(engine.rules, rule)
	}
	return engine, nil
}

func Read(root, path string) (content []byte, err error) {
	dir, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, dir.Close()) }()
	return readRoot(dir, path)
}

func readScript(root, folder, path string) (content []byte, err error) {
	dir, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, dir.Close()) }()
	info, err := dir.Lstat(folder)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s must be a non-symlink directory", folder)
	}
	scripts, err := dir.OpenRoot(folder)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, scripts.Close()) }()
	return readRoot(scripts, path)
}

func readRoot(dir *os.Root, path string) (content []byte, err error) {
	info, err := dir.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("analysis requires a regular, non-symlink file")
	}
	file, err := dir.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	info, err = file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("analysis file is no longer a regular file")
	}
	content, err = io.ReadAll(io.LimitReader(file, ContentLimit+1))
	if err != nil {
		return nil, err
	}
	if len(content) > ContentLimit {
		return nil, errors.New("file exceeds 4 MiB analysis limit")
	}
	if !utf8.Valid(content) || strings.IndexByte(string(content), 0) >= 0 {
		return nil, errors.New("analysis requires UTF-8 text without NUL bytes")
	}
	return content, nil
}

func (engine *Engine) Evaluate(ctx context.Context, path string, content []byte) Result {
	result := Result{Path: path, Digest: Digest(content)}
	if !engine.Config.Applies(path) {
		return result
	}
	var err error
	parseCtx, cancel := context.WithTimeout(ctx, time.Duration(engine.Config.TimeoutMS)*time.Millisecond)
	result.Syntax, err = Parse(parseCtx, path, content)
	cancel()
	if err != nil {
		result.Findings = append(result.Findings, ErrorFinding(path, "parser", err))
	}
	for _, rule := range engine.rules {
		if rule.spec.Source != "" && (!config.Matches(rule.spec.ScopeInclude, path) || config.Matches(rule.spec.ScopeExclude, path)) {
			continue
		}
		if ctx.Err() != nil {
			result.Findings = append(result.Findings, ErrorFinding(path, rule.spec.ID, ctx.Err()))
			break
		}
		if rule.spec.Include != nil && !config.Matches(rule.spec.Include, path) {
			continue
		}
		var findings []Finding
		if rule.regex != nil {
			matches := rule.regex.FindAllIndex(content, FindingLimit+1)
			if len(matches) > FindingLimit {
				err = errors.New("regex exceeded 1000 findings per file")
			} else {
				err = nil
				for _, match := range matches {
					line, column := position(content, match[0])
					findings = append(findings, Finding{Path: path, Rule: rule.spec.ID,
						Kind: "violation", Line: line, Column: column, Message: ruleMessage(rule.spec)})
				}
			}
		} else {
			findings, err = engine.javascript(ctx, rule, path, string(content))
		}
		if err != nil {
			findings = []Finding{ErrorFinding(path, rule.spec.ID, err)}
		}
		for i := range findings {
			findings[i].Set, findings[i].Source = rule.spec.Set, rule.spec.Source
		}
		result.Findings = append(result.Findings, findings...)
	}
	return result
}

func position(content []byte, offset int) (int, int) {
	prefix := string(content[:offset])
	return strings.Count(prefix, "\n") + 1, utf8.RuneCountInString(prefix[strings.LastIndex(prefix, "\n")+1:]) + 1
}

func ruleMessage(rule config.Rule) string {
	if rule.Message != "" {
		return rule.Message
	}
	return "Rule " + rule.ID + " was violated"
}
