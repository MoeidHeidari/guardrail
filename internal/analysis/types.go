package analysis

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
)

type Finding struct {
	Set     string `json:"set"`
	Source  string `json:"source"`
	Path    string `json:"path"`
	Rule    string `json:"rule"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Kind    string `json:"kind"`
}

type Result struct {
	Path     string    `json:"path"`
	Digest   string    `json:"digest"`
	Syntax   Syntax    `json:"syntax"`
	Findings []Finding `json:"findings"`
}

type Report struct {
	Files   []Result `json:"files"`
	Checked int      `json:"checked"`
	Skipped int      `json:"skipped"`
}

func Digest(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func ErrorFinding(path, rule string, err error) Finding {
	finding := Finding{Path: path, Rule: rule, Kind: "error", Message: err.Error()}
	var syntax *SyntaxError
	if errors.As(err, &syntax) {
		finding.Line, finding.Column = syntax.Line, syntax.Column
	}
	return finding
}

func (report Report) Findings() []Finding {
	var findings []Finding
	for _, file := range report.Files {
		findings = append(findings, file.Findings...)
	}
	sort.Slice(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		if a.Set != b.Set {
			return a.Set < b.Set
		}
		return a.Rule < b.Rule
	})
	return findings
}

func (report Report) ExitCode() int {
	code := 0
	for _, finding := range report.Findings() {
		if finding.Kind == "error" {
			return 2
		}
		code = 1
	}
	return code
}

func (finding Finding) String() string {
	rule := finding.Rule
	if finding.Set != "" {
		rule = finding.Set + "/" + rule
	}
	return fmt.Sprintf("%s:%d:%d [%s] %s: %s", finding.Path, finding.Line,
		finding.Column, finding.Kind, rule, finding.Message)
}
