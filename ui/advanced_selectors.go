package ui

import (
	"strconv"
	"strings"
	"unicode"
)

type AdvancedSelectorRule struct {
	Selector string
	Steps    []AdvancedSelectorStep
	Props    map[string]string
}

type AdvancedSelectorStep struct {
	Tag        string
	ID         string
	Classes    []string
	Pseudos    []SelectorPseudo
	Combinator string
}

type SelectorPseudo struct {
	Name string
	Arg  string
}

type AdvancedSelectorContext struct {
	Path                 string
	Props                map[string]any
	LookupProps          func(path string) (map[string]any, bool)
	ParentPath           func(path string) string
	PreviousSiblingPath  func(path string) string
	PreviousSiblingPaths func(path string) []string
	PseudoState          func(path string, props map[string]any, pseudo string) bool
}

func isAdvancedSelector(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return false
	}
	if strings.Contains(strings.ToLower(trimmed), ":nth-child(") {
		return true
	}
	depth := 0
	lastNonSpace := byte(0)
	for i := 0; i < len(trimmed); i++ {
		ch := trimmed[i]
		switch ch {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case '>', '+', '~':
			if depth == 0 {
				return true
			}
		case ' ', '\t', '\n', '\r':
			if depth == 0 && lastNonSpace != 0 {
				for j := i + 1; j < len(trimmed); j++ {
					next := trimmed[j]
					if next == ' ' || next == '\t' || next == '\n' || next == '\r' {
						continue
					}
					if next != '>' && next != '+' && next != '~' && lastNonSpace != '>' && lastNonSpace != '+' && lastNonSpace != '~' {
						return true
					}
					break
				}
			}
		default:
			lastNonSpace = ch
		}
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			lastNonSpace = ch
		}
	}
	return false
}

func parseAdvancedSelectorRule(selector string, decls map[string]string) (AdvancedSelectorRule, bool) {
	steps, ok := parseAdvancedSelectorSteps(selector)
	if !ok {
		return AdvancedSelectorRule{}, false
	}
	return AdvancedSelectorRule{
		Selector: strings.TrimSpace(selector),
		Steps:    steps,
		Props:    cloneStringMap(decls),
	}, true
}

func parseAdvancedSelectorSteps(selector string) ([]AdvancedSelectorStep, bool) {
	trimmed := strings.TrimSpace(selector)
	if trimmed == "" {
		return nil, false
	}
	var (
		steps       []AdvancedSelectorStep
		pendingComb string
		i           int
	)
	for i < len(trimmed) {
		sawSpace := false
		for i < len(trimmed) {
			ch := trimmed[i]
			if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
				break
			}
			sawSpace = true
			i++
		}
		if i >= len(trimmed) {
			break
		}
		if sawSpace && len(steps) > 0 && pendingComb == "" {
			pendingComb = " "
		}
		if trimmed[i] == '>' || trimmed[i] == '+' || trimmed[i] == '~' {
			pendingComb = string(trimmed[i])
			i++
			continue
		}
		start := i
		depth := 0
		for i < len(trimmed) {
			ch := trimmed[i]
			if ch == '(' {
				depth++
			} else if ch == ')' {
				if depth > 0 {
					depth--
				}
			} else if depth == 0 {
				if ch == '>' || ch == '+' || ch == '~' || ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
					break
				}
			}
			i++
		}
		step, ok := parseAdvancedSelectorStep(trimmed[start:i])
		if !ok {
			return nil, false
		}
		step.Combinator = pendingComb
		pendingComb = ""
		steps = append(steps, step)
	}
	if len(steps) == 0 {
		return nil, false
	}
	steps[0].Combinator = ""
	return steps, true
}

func parseAdvancedSelectorStep(raw string) (AdvancedSelectorStep, bool) {
	step := AdvancedSelectorStep{}
	token := strings.TrimSpace(raw)
	if token == "" {
		return step, false
	}
	for i := 0; i < len(token); {
		switch token[i] {
		case '*':
			if step.Tag == "" {
				step.Tag = "*"
			}
			i++
		case '.':
			name, next := readSelectorIdentifier(token, i+1)
			if name == "" {
				return AdvancedSelectorStep{}, false
			}
			step.Classes = append(step.Classes, name)
			i = next
		case '#':
			name, next := readSelectorIdentifier(token, i+1)
			if name == "" {
				return AdvancedSelectorStep{}, false
			}
			step.ID = name
			i = next
		case ':':
			name, next := readSelectorIdentifier(token, i+1)
			if name == "" {
				return AdvancedSelectorStep{}, false
			}
			pseudo := SelectorPseudo{Name: strings.ToLower(name)}
			i = next
			if i < len(token) && token[i] == '(' {
				arg, end, ok := readPseudoArgument(token, i)
				if !ok {
					return AdvancedSelectorStep{}, false
				}
				pseudo.Arg = strings.TrimSpace(arg)
				i = end
			}
			step.Pseudos = append(step.Pseudos, pseudo)
		default:
			name, next := readSelectorIdentifier(token, i)
			if name == "" {
				return AdvancedSelectorStep{}, false
			}
			if step.Tag == "" {
				step.Tag = strings.ToLower(name)
			}
			i = next
		}
	}
	return step, true
}

func readSelectorIdentifier(token string, start int) (string, int) {
	i := start
	for i < len(token) {
		r := rune(token[i])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || token[i] == '-' || token[i] == '_' {
			i++
			continue
		}
		break
	}
	return token[start:i], i
}

func readPseudoArgument(token string, open int) (string, int, bool) {
	depth := 0
	for i := open; i < len(token); i++ {
		switch token[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return token[open+1 : i], i + 1, true
			}
		}
	}
	return "", 0, false
}

func ApplyAdvancedSelectorStyles(css map[string]string, ss *StyleSheet, ctx AdvancedSelectorContext) {
	if css == nil || ss == nil || len(ctx.Path) == 0 || ctx.LookupProps == nil || ctx.ParentPath == nil {
		return
	}
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	if len(ss.advancedRules) == 0 {
		return
	}
	for _, rule := range ss.advancedRules {
		if !matchAdvancedRule(rule, ctx) {
			continue
		}
		for k, v := range rule.Props {
			css[CanonicalName(k)] = v
		}
	}
	CSSExpandBoxShorthand(css, "padding")
	CSSExpandBoxShorthand(css, "margin")
	CSSExpandBorderShorthand(css)
	CSSResolveVariables(css)
}

func matchAdvancedRule(rule AdvancedSelectorRule, ctx AdvancedSelectorContext) bool {
	if len(rule.Steps) == 0 || ctx.Props == nil {
		return false
	}
	return matchAdvancedSelectorStep(rule.Steps, len(rule.Steps)-1, ctx.Path, ctx.Props, ctx)
}

func matchAdvancedSelectorStep(steps []AdvancedSelectorStep, idx int, path string, props map[string]any, ctx AdvancedSelectorContext) bool {
	if idx < 0 || !matchAdvancedCompoundSelector(steps[idx], path, props, ctx) {
		return false
	}
	if idx == 0 {
		return true
	}
	combinator := steps[idx].Combinator
	switch combinator {
	case ">":
		parent := ctx.ParentPath(path)
		if parent == "" {
			return false
		}
		parentProps, ok := ctx.LookupProps(parent)
		return ok && matchAdvancedSelectorStep(steps, idx-1, parent, parentProps, ctx)
	case "+":
		if ctx.PreviousSiblingPath == nil {
			return false
		}
		sibling := ctx.PreviousSiblingPath(path)
		if sibling == "" {
			return false
		}
		siblingProps, ok := ctx.LookupProps(sibling)
		return ok && matchAdvancedSelectorStep(steps, idx-1, sibling, siblingProps, ctx)
	case "~":
		if ctx.PreviousSiblingPaths == nil {
			return false
		}
		for _, sibling := range ctx.PreviousSiblingPaths(path) {
			siblingProps, ok := ctx.LookupProps(sibling)
			if ok && matchAdvancedSelectorStep(steps, idx-1, sibling, siblingProps, ctx) {
				return true
			}
		}
		return false
	default:
		for parent := ctx.ParentPath(path); parent != ""; parent = ctx.ParentPath(parent) {
			parentProps, ok := ctx.LookupProps(parent)
			if ok && matchAdvancedSelectorStep(steps, idx-1, parent, parentProps, ctx) {
				return true
			}
		}
		return false
	}
}

func matchAdvancedCompoundSelector(step AdvancedSelectorStep, path string, props map[string]any, ctx AdvancedSelectorContext) bool {
	tag := strings.ToLower(strings.TrimSpace(anyToString(props["tag"], "")))
	if step.Tag != "" && step.Tag != "*" && step.Tag != tag {
		return false
	}
	if step.ID != "" && strings.TrimSpace(anyToString(props["id"], "")) != step.ID {
		return false
	}
	if len(step.Classes) > 0 {
		classSet := make(map[string]struct{}, len(step.Classes))
		for _, cls := range UIClassTokens(anyToString(props["class"], "")) {
			classSet[cls] = struct{}{}
		}
		for _, cls := range step.Classes {
			if _, ok := classSet[cls]; !ok {
				return false
			}
		}
	}
	for _, pseudo := range step.Pseudos {
		if !matchAdvancedPseudo(path, props, pseudo, ctx) {
			return false
		}
	}
	return true
}

func matchAdvancedPseudo(path string, props map[string]any, pseudo SelectorPseudo, ctx AdvancedSelectorContext) bool {
	switch pseudo.Name {
	case "nth-child":
		return matchNthChild(path, pseudo.Arg)
	default:
		if ctx.PseudoState == nil {
			return false
		}
		return ctx.PseudoState(path, props, pseudo.Name)
	}
}

func matchNthChild(path, expr string) bool {
	index, ok := selectorChildOrdinal(path)
	if !ok {
		return false
	}
	a, b, ok := parseNthChildExpression(expr)
	if !ok {
		return false
	}
	if a == 0 {
		return index == b
	}
	if a > 0 {
		delta := index - b
		return delta >= 0 && delta%a == 0
	}
	delta := b - index
	return delta >= 0 && delta%(-a) == 0
}

func selectorChildOrdinal(path string) (int, bool) {
	idx := strings.LastIndex(path, "/")
	if idx < 0 || idx+1 >= len(path) {
		return 0, false
	}
	childIndex, err := strconv.Atoi(path[idx+1:])
	if err != nil || childIndex < 0 {
		return 0, false
	}
	return childIndex + 1, true
}

func parseNthChildExpression(expr string) (int, int, bool) {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(expr), " ", ""))
	switch normalized {
	case "odd":
		return 2, 1, true
	case "even":
		return 2, 0, true
	case "":
		return 0, 0, false
	}
	if !strings.Contains(normalized, "n") {
		value, err := strconv.Atoi(normalized)
		return 0, value, err == nil
	}
	parts := strings.SplitN(normalized, "n", 2)
	aRaw := parts[0]
	bRaw := parts[1]
	var a int
	switch aRaw {
	case "", "+":
		a = 1
	case "-":
		a = -1
	default:
		parsed, err := strconv.Atoi(aRaw)
		if err != nil {
			return 0, 0, false
		}
		a = parsed
	}
	b := 0
	if bRaw != "" {
		parsed, err := strconv.Atoi(bRaw)
		if err != nil {
			return 0, 0, false
		}
		b = parsed
	}
	return a, b, true
}
