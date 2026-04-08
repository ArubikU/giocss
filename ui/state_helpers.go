package ui

import "strings"

func applyPseudoStateStyles(css map[string]string, props map[string]any, ss *StyleSheet, stateEnabled bool, pseudo string) {
	if !stateEnabled || ss == nil {
		return
	}
	for _, cls := range UIClassTokens(anyToString(props["class"], "")) {
		if m := ss.GetClassProps(cls + pseudo); m != nil {
			for k, v := range m {
				css[CanonicalName(k)] = v
			}
		}
	}
	if strings.TrimSpace(anyToString(props["tag"], "")) != "" {
		tag := strings.ToLower(strings.TrimSpace(anyToString(props["tag"], "")))
		if m := ss.GetClassProps(tag + pseudo); m != nil {
			for k, v := range m {
				css[CanonicalName(k)] = v
			}
		}
	}
	CSSExpandBoxShorthand(css, "padding")
	CSSExpandBoxShorthand(css, "margin")
	CSSExpandBorderShorthand(css)
	CSSResolveVariables(css)
}

func hasPseudoStateSelector(props map[string]any, ss *StyleSheet, pseudo string) bool {
	if ss == nil {
		return false
	}
	for _, cls := range UIClassTokens(anyToString(props["class"], "")) {
		if m := ss.GetClassProps(cls + pseudo); len(m) > 0 {
			return true
		}
	}
	tag := strings.ToLower(strings.TrimSpace(anyToString(props["tag"], "")))
	if tag != "" {
		if m := ss.GetClassProps(tag + pseudo); len(m) > 0 {
			return true
		}
	}
	return false
}

func UIClassTokens(raw string) []string {
	parts := strings.Fields(strings.TrimSpace(raw))
	if len(parts) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		cls := strings.TrimSpace(strings.TrimPrefix(part, "."))
		if cls == "" {
			continue
		}
		if _, ok := seen[cls]; ok {
			continue
		}
		seen[cls] = struct{}{}
		out = append(out, cls)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func ApplyHoverStyles(css map[string]string, props map[string]any, ss *StyleSheet, hovered bool) {
	applyPseudoStateStyles(css, props, ss, hovered, ":hover")
}

func ApplyActiveStyles(css map[string]string, props map[string]any, ss *StyleSheet, active bool) {
	applyPseudoStateStyles(css, props, ss, active, ":active")
}

func HasHoverSelector(props map[string]any, ss *StyleSheet) bool {
	return hasPseudoStateSelector(props, ss, ":hover")
}

func HasActiveSelector(props map[string]any, ss *StyleSheet) bool {
	return hasPseudoStateSelector(props, ss, ":active")
}

func ApplyFocusStyles(css map[string]string, props map[string]any, ss *StyleSheet, focused bool) {
	applyPseudoStateStyles(css, props, ss, focused, ":focus")
}

func ApplyDisabledStyles(css map[string]string, props map[string]any, ss *StyleSheet, disabled bool) {
	applyPseudoStateStyles(css, props, ss, disabled, ":disabled")
}

func ApplyCheckedStyles(css map[string]string, props map[string]any, ss *StyleSheet, checked bool) {
	applyPseudoStateStyles(css, props, ss, checked, ":checked")
}

func ApplyInvalidStyles(css map[string]string, props map[string]any, ss *StyleSheet, invalid bool) {
	applyPseudoStateStyles(css, props, ss, invalid, ":invalid")
}

func HasFocusSelector(props map[string]any, ss *StyleSheet) bool {
	return hasPseudoStateSelector(props, ss, ":focus")
}

func HasDisabledSelector(props map[string]any, ss *StyleSheet) bool {
	return hasPseudoStateSelector(props, ss, ":disabled")
}

func HasCheckedSelector(props map[string]any, ss *StyleSheet) bool {
	return hasPseudoStateSelector(props, ss, ":checked")
}

func HasInvalidSelector(props map[string]any, ss *StyleSheet) bool {
	return hasPseudoStateSelector(props, ss, ":invalid")
}
