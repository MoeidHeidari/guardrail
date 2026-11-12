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

func (engine *Engine) javascript(ctx context.Context, rule compiledRule, path, content string) ([]Finding, error) {
	runtime := goja.New()
	runtime.SetMaxCallStackSize(256)
	validator, err := runtime.RunString(`(() => {
		const stringify = JSON.stringify, isArray = Array.isArray;
		return value => {
			if (typeof value === "boolean") return value;
			if (!isArray(value)) throw Error("rule must return a boolean or an array of findings");
			if (value.length > 1000) throw Error("rule exceeded 1000 findings per file");
			const result = stringify(value);
			if (result === "null" || typeof result !== "string") throw Error("invalid findings");
			return result;
		};
	})()`)
	if err != nil {
		return nil, err
	}
	validate, ok := goja.AssertFunction(validator)
	if !ok {
		return nil, errors.New("cannot initialize result validation")
	}
	timeout := engine.Config.TimeoutMS
	if rule.spec.TimeoutMS > 0 {
		timeout = rule.spec.TimeoutMS
	}
	deadline, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()
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
	value, err := runtime.RunProgram(rule.program)
	if err != nil {
		return nil, err
	}
	function, ok := goja.AssertFunction(value)
	if !ok {
		return nil, errors.New("JavaScript rule must evaluate to a function")
	}
	value, err = function(goja.Undefined(), runtime.ToValue(path), runtime.ToValue(content))
	if err != nil {
		return nil, err
	}
	value, err = validate(goja.Undefined(), value)
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
