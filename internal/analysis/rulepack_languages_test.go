package analysis

import "testing"

func TestCommunityLanguageRuleSets(t *testing.T) {
	engine := communityPackEngine(t, "architecture", "source-policy", "complexity",
		"api-policy", "dependency-policy", "maintainability", "repository-hygiene")
	if len(engine.rules) != 25 {
		t.Fatalf("expected all 25 language-scoped rules, got %d", len(engine.rules))
	}
	for _, test := range []struct {
		path, set string
		count     int
	}{
		{"a.go", "go", 4},
		{"a.js", "javascript", 4},
		{"a.jsx", "javascript", 4},
		{"a.mjs", "javascript", 4},
		{"a.cjs", "javascript", 4},
		{"a.ts", "typescript", 4},
		{"a.tsx", "typescript", 4},
		{"a.py", "python", 4},
		{"a.rs", "rust", 4},
		{"package.json", "javascript", 1},
		{"nested/package.json", "javascript", 1},
		{"a.txt", "general", 4},
		{"a.json", "", 0},
		{"a.java", "", 0},
	} {
		t.Run(test.path, func(t *testing.T) {
			count := 0
			for _, rule := range engine.rules {
				if !rule.applies(test.path) {
					continue
				}
				count++
				if rule.spec.Set != test.set {
					t.Errorf("rule %s from %s incorrectly applies to %s", rule.spec.ID, rule.spec.Set, test.path)
				}
			}
			if count != test.count {
				t.Fatalf("expected %d applicable rules, got %d", test.count, count)
			}
		})
	}
}
