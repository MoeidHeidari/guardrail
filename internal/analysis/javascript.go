package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dop251/goja"
)

type jsFinding struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
}

func (engine *Engine) javascript(ctx context.Context, rule compiledRule, path, content, syntaxJSON string, snapshot syntaxSnapshot) ([]Finding, error) {
	timeout := engine.Config.TimeoutMS
	if rule.spec.TimeoutMS > 0 {
		timeout = rule.spec.TimeoutMS
	}
	deadline, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()
	if err := deadline.Err(); err != nil {
		return nil, err
	}
	runtime := goja.New()
	runtime.SetMaxCallStackSize(256)
	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		select {
		case <-deadline.Done():
			runtime.Interrupt(fmt.Errorf("JavaScript execution interrupted: %w", deadline.Err()))
		case <-done:
		}
	}()
	defer func() { close(done); <-stopped }()
	session, err := newJSSession(runtime, content, syntaxJSON, snapshot)
	if err != nil {
		return nil, err
	}
	change, err := session.buildChange(goja.Undefined(), runtime.ToValue(content), session.api)
	if err != nil {
		return nil, err
	}
	value, err := runtime.RunProgram(rule.program)
	if err != nil {
		return nil, err
	}
	function, ok := goja.AssertFunction(value)
	if !ok {
		return nil, errors.New("JavaScript rule must evaluate to a function")
	}
	value, err = function(goja.Undefined(), runtime.ToValue(path), change)
	if err != nil {
		return nil, err
	}
	value, err = session.validate(goja.Undefined(), value)
	if err != nil {
		return nil, err
	}
	if deadline.Err() != nil {
		return nil, deadline.Err()
	}
	if boolean, ok := value.Export().(bool); ok {
		if !boolean {
			return nil, nil
		}
		return []Finding{{Path: path, Rule: rule.spec.ID, Kind: "violation", Message: ruleMessage(rule.spec)}}, nil
	}
	var locations []jsFinding
	decoder := json.NewDecoder(strings.NewReader(value.String()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&locations); err != nil {
		return nil, fmt.Errorf("invalid findings: %w", err)
	}
	if len(locations) > FindingLimit {
		return nil, errors.New("JavaScript exceeded 1000 findings per file")
	}
	lines := strings.Split(content, "\n")
	var findings []Finding
	for _, location := range locations {
		if location.Line < 1 || location.Line > len(lines) || location.Column < 1 ||
			location.Column > utf8.RuneCountInString(lines[location.Line-1])+1 {
			return nil, errors.New("finding line/column is outside the file (locations are 1-based)")
		}
		if len(location.Message) > 4096 {
			return nil, errors.New("finding message exceeds 4096 bytes")
		}
		if location.Message == "" {
			location.Message = ruleMessage(rule.spec)
		}
		findings = append(findings, Finding{Path: path, Rule: rule.spec.ID, Kind: "violation",
			Line: location.Line, Column: location.Column, Message: location.Message})
	}
	return findings, nil
}
