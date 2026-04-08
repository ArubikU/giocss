package ui

import (
	"strconv"
	"testing"
)

func TestApplyAdvancedSelectorStylesDescendantCombinator(t *testing.T) {
	ss := NewStyleSheet()
	if got := ss.ParseCSSText("form .field { color: red; }"); got != 1 {
		t.Fatalf("ParseCSSText() = %d, want 1", got)
	}
	css := map[string]string{}
	propsByPath := map[string]map[string]any{
		"root/0":   {"tag": "form"},
		"root/0/0": {"tag": "input", "class": "field"},
	}
	ApplyAdvancedSelectorStyles(css, ss, testAdvancedSelectorContext("root/0/0", propsByPath))
	if css["color"] != "red" {
		t.Fatalf("descendant selector color = %q, want %q", css["color"], "red")
	}
}

func TestApplyAdvancedSelectorStylesChildCombinator(t *testing.T) {
	ss := NewStyleSheet()
	ss.ParseCSSText("form > .field { color: red; }")
	css := map[string]string{}
	propsByPath := map[string]map[string]any{
		"root/0":     {"tag": "form"},
		"root/0/0":   {"tag": "div"},
		"root/0/0/0": {"tag": "input", "class": "field"},
	}
	ApplyAdvancedSelectorStyles(css, ss, testAdvancedSelectorContext("root/0/0/0", propsByPath))
	if css["color"] != "" {
		t.Fatalf("child selector color = %q, want empty", css["color"])
	}
}

func TestApplyAdvancedSelectorStylesAdjacentSibling(t *testing.T) {
	ss := NewStyleSheet()
	ss.ParseCSSText("label + input { border-color: blue; }")
	css := map[string]string{}
	propsByPath := map[string]map[string]any{
		"root/0/0": {"tag": "label"},
		"root/0/1": {"tag": "input"},
	}
	ApplyAdvancedSelectorStyles(css, ss, testAdvancedSelectorContext("root/0/1", propsByPath))
	if css["border-color"] != "blue" {
		t.Fatalf("adjacent sibling border-color = %q, want %q", css["border-color"], "blue")
	}
}

func TestApplyAdvancedSelectorStylesGeneralSiblingAndNthChild(t *testing.T) {
	ss := NewStyleSheet()
	ss.ParseCSSText(".group ~ .item:nth-child(3) { background-color: gold; }")
	css := map[string]string{}
	propsByPath := map[string]map[string]any{
		"root/0/0": {"tag": "div", "class": "group"},
		"root/0/1": {"tag": "div", "class": "item"},
		"root/0/2": {"tag": "div", "class": "item"},
	}
	ApplyAdvancedSelectorStyles(css, ss, testAdvancedSelectorContext("root/0/2", propsByPath))
	if css["background-color"] != "gold" {
		t.Fatalf("general sibling nth-child background-color = %q, want %q", css["background-color"], "gold")
	}
}

func TestParseNthChildExpressionVariants(t *testing.T) {
	tests := []struct {
		expr  string
		wantA int
		wantB int
	}{
		{expr: "odd", wantA: 2, wantB: 1},
		{expr: "even", wantA: 2, wantB: 0},
		{expr: "2n+1", wantA: 2, wantB: 1},
		{expr: "-n+3", wantA: -1, wantB: 3},
		{expr: "5", wantA: 0, wantB: 5},
	}
	for _, tt := range tests {
		a, b, ok := parseNthChildExpression(tt.expr)
		if !ok || a != tt.wantA || b != tt.wantB {
			t.Fatalf("parseNthChildExpression(%q) = (%d, %d, %v), want (%d, %d, true)", tt.expr, a, b, ok, tt.wantA, tt.wantB)
		}
	}
}

func testAdvancedSelectorContext(path string, propsByPath map[string]map[string]any) AdvancedSelectorContext {
	return AdvancedSelectorContext{
		Path:  path,
		Props: propsByPath[path],
		LookupProps: func(target string) (map[string]any, bool) {
			props, ok := propsByPath[target]
			return props, ok
		},
		ParentPath: func(target string) string {
			idx := -1
			for i := len(target) - 1; i >= 0; i-- {
				if target[i] == '/' {
					idx = i
					break
				}
			}
			if idx <= 0 {
				return ""
			}
			return target[:idx]
		},
		PreviousSiblingPath: func(target string) string {
			siblings := testPreviousSiblingPaths(target, propsByPath)
			if len(siblings) == 0 {
				return ""
			}
			return siblings[0]
		},
		PreviousSiblingPaths: func(target string) []string {
			return testPreviousSiblingPaths(target, propsByPath)
		},
	}
}

func testPreviousSiblingPaths(path string, propsByPath map[string]map[string]any) []string {
	parent := testAdvancedSelectorContext(path, propsByPath).ParentPath(path)
	if parent == "" {
		return nil
	}
	lastSlash := len(path) - 1
	for ; lastSlash >= 0; lastSlash-- {
		if path[lastSlash] == '/' {
			break
		}
	}
	if lastSlash < 0 || lastSlash+1 >= len(path) {
		return nil
	}
	childIndex, ok := parseTestInt(path[lastSlash+1:])
	if !ok {
		return nil
	}
	out := make([]string, 0, childIndex)
	for i := childIndex - 1; i >= 0; i-- {
		sibling := parent + "/" + testItoa(i)
		if _, exists := propsByPath[sibling]; exists {
			out = append(out, sibling)
		}
	}
	return out
}

func parseTestInt(raw string) (int, bool) {
	value, err := strconv.Atoi(raw)
	return value, err == nil
}

func testItoa(value int) string {
	return strconv.Itoa(value)
}
