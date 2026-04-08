package layout

import (
	"strconv"
	"strings"

	pkgcore "github.com/ArubikU/giocss/pkg"
	uicore "github.com/ArubikU/giocss/ui"
)

func IntrinsicNodeSize(node *Node, ss *uicore.StyleSheet, viewportW int, viewportH int, parentCSS map[string]string) (int, int) {
	if node == nil {
		return 0, 0
	}
	css := ResolveNodeStyle(node, ss, viewportW)
	css = MergeInheritedTextCSS(css, parentCSS)
	if strings.EqualFold(strings.TrimSpace(css["display"]), "none") {
		return 0, 0
	}
	hasExplicitWidth := strings.TrimSpace(css["width"]) != ""
	hasExplicitHeight := strings.TrimSpace(css["height"]) != ""
	width := NodeLayoutLength(node, css, "width", "width", viewportW, viewportW, viewportH, -1)
	height := NodeLayoutLength(node, css, "height", "height", viewportH, viewportW, viewportH, -1)
	paddingTop := NodeLayoutLength(node, css, "paddingTop", "padding-top", viewportH, viewportW, viewportH, nodePropInt(node, "pady", nodePropInt(node, "padding", 0)))
	paddingRight := NodeLayoutLength(node, css, "paddingRight", "padding-right", viewportW, viewportW, viewportH, nodePropInt(node, "padx", nodePropInt(node, "padding", 0)))
	paddingBottom := NodeLayoutLength(node, css, "paddingBottom", "padding-bottom", viewportH, viewportW, viewportH, nodePropInt(node, "pady", nodePropInt(node, "padding", 0)))
	paddingLeft := NodeLayoutLength(node, css, "paddingLeft", "padding-left", viewportW, viewportW, viewportH, nodePropInt(node, "padx", nodePropInt(node, "padding", 0)))
	gap := NodeLayoutLength(node, css, "gap", "gap", max(viewportW, viewportH), viewportW, viewportH, 0)
	kind := strings.ToLower(strings.TrimSpace(node.Tag))

	if width >= 0 && height >= 0 {
		width, height = applyIntrinsicMinMax(node, css, viewportW, viewportH, width, height)
		return max(1, width), max(1, height)
	}

	switch kind {
	case "text", "label", "span", "p", "h1", "h2", "h3", "h4", "h5", "h6":
		fontSize := uicore.CSSFontSize(css, 14)
		text := uicore.CSSApplyLetterSpacing(uicore.CSSTextTransform(nodePropString(node, "text", ""), css), css)
		if uicore.CSSAllowsWrap(css) && !pkgcore.ShouldClipText(css) {
			wrapW := width
			if wrapW <= 0 {
				wrapW = viewportW
			}
			inner := max(1, wrapW-paddingLeft-paddingRight)
			text = uicore.WrapTextToWidth(text, fontSize, inner)
		}
		w, h := uicore.EstimateTextLayoutBox(text, fontSize, css)
		if width < 0 {
			width = w + paddingLeft + paddingRight
		}
		if height < 0 {
			height = h + paddingTop + paddingBottom
		}
	case "button":
		if len(node.Children) > 0 {
			contentW, contentH := intrinsicSizeFromChildren(node, ss, css, viewportW, viewportH, width, paddingLeft, paddingRight, gap)
			if width < 0 {
				width = contentW + paddingLeft + paddingRight
			}
			if height < 0 {
				height = contentH + paddingTop + paddingBottom
			}
			break
		}
		fontSize := uicore.CSSFontSize(css, 13)
		text := uicore.CSSApplyLetterSpacing(uicore.CSSTextTransform(nodePropString(node, "text", ""), css), css)
		w, h := uicore.EstimateTextLayoutBox(text, fontSize, css)
		iconExtra := 0
		if nodePropString(node, "icon", "") != "" {
			iconExtra = max(18, int(fontSize)+4)
		}
		if width < 0 {
			width = w + iconExtra + paddingLeft + paddingRight + 28
		}
		if height < 0 {
			height = max(30, h+paddingTop+paddingBottom+10)
		}
	case "input":
		inputType := strings.ToLower(nodePropString(node, "type", "text"))
		fontSize := uicore.CSSFontSize(css, 13)
		if width < 0 {
			if inputType == "checkbox" || inputType == "check" {
				width = 120
			} else if inputType == "range" || inputType == "slider" {
				width = 220
			} else {
				width = 180
			}
		}
		if height < 0 {
			if inputType == "textarea" || inputType == "multiline" {
				minLine := max(18, int(float64(fontSize)*1.35))
				height = max(92, minLine*3+paddingTop+paddingBottom+8)
			} else {
				minLine := max(16, int(float64(fontSize)*1.35))
				height = max(36, minLine+paddingTop+paddingBottom+2)
			}
		}
	case "select":
		fontSize := uicore.CSSFontSize(css, 13)
		if width < 0 {
			width = 180
		}
		if height < 0 {
			minLine := max(16, int(float64(fontSize)*1.35))
			height = max(36, minLine+paddingTop+paddingBottom+2)
		}
	case "native":
		component := strings.ToLower(nodePropString(node, "component", nodePropString(node, "native", "label")))
		if width < 0 {
			switch component {
			case "image", "img", "svg":
				width = 180
			case "icon":
				width = 18
			case "slider":
				width = 220
			case "progress", "progressbar":
				width = 220
			case "select", "dropdown":
				width = 180
			default:
				width = 120
			}
		}
		if height < 0 {
			switch component {
			case "image", "img", "svg":
				height = 120
			case "icon":
				height = 18
			case "progress", "progressbar":
				height = 14
			default:
				height = 36
			}
		}
	default:
		contentW, contentH := intrinsicSizeFromChildren(node, ss, css, viewportW, viewportH, width, paddingLeft, paddingRight, gap)
		if width < 0 {
			width = contentW + paddingLeft + paddingRight
		}
		if height < 0 {
			height = contentH + paddingTop + paddingBottom
		}
	}

	if ar, ok := uicore.ParseAspectRatio(css); ok && ar > 0.001 {
		switch {
		case width > 0 && !hasExplicitHeight:
			height = max(1, int(float64(width)/ar))
		case height > 0 && !hasExplicitWidth:
			width = max(1, int(float64(height)*ar))
		case width <= 0 && height <= 0:
			baseW := max(96, paddingLeft+paddingRight+96)
			width = baseW
			height = max(1, int(float64(baseW)/ar))
		}
	}

	width, height = applyIntrinsicMinMax(node, css, viewportW, viewportH, width, height)

	return max(1, width), max(1, height)
}

func intrinsicSizeFromChildren(node *Node, ss *uicore.StyleSheet, css map[string]string, viewportW int, viewportH int, width int, paddingLeft int, paddingRight int, gap int) (int, int) {
	direction := NodeLayoutString(node, css, "direction", "direction", strings.ToLower(nodePropString(node, "layout", "column")))
	if direction != "row" {
		direction = "column"
	}
	contentW := 0
	contentH := 0
	visibleChildren := 0
	childViewportW := viewportW
	if width > 0 {
		childViewportW = max(1, width-paddingLeft-paddingRight)
	}
	for _, child := range node.Children {
		childCSS := ResolveNodeStyle(child, ss, viewportW)
		childCSS = MergeInheritedTextCSS(childCSS, css)
		if strings.EqualFold(strings.TrimSpace(childCSS["display"]), "none") {
			continue
		}
		if isAbsoluteOrFixedPositioned(childCSS) {
			continue
		}
		cw, ch := IntrinsicNodeSize(child, ss, childViewportW, viewportH, css)
		mw := NodeLayoutLength(child, childCSS, "marginLeft", "margin-left", viewportW, viewportW, viewportH, 0) + NodeLayoutLength(child, childCSS, "marginRight", "margin-right", viewportW, viewportW, viewportH, 0)
		mh := NodeLayoutLength(child, childCSS, "marginTop", "margin-top", viewportH, viewportW, viewportH, 0) + NodeLayoutLength(child, childCSS, "marginBottom", "margin-bottom", viewportH, viewportW, viewportH, 0)
		if direction == "row" {
			contentW += cw + mw
			contentH = max(contentH, ch+mh)
		} else {
			contentH += ch + mh
			contentW = max(contentW, cw+mw)
		}
		visibleChildren++
	}
	if visibleChildren > 1 {
		if direction == "row" {
			contentW += gap * (visibleChildren - 1)
		} else {
			contentH += gap * (visibleChildren - 1)
		}
	}
	return contentW, contentH
}

func applyIntrinsicMinMax(node *Node, css map[string]string, viewportW int, viewportH int, width int, height int) (int, int) {
	minW := NodeLayoutLength(node, css, "minWidth", "min-width", viewportW, viewportW, viewportH, -1)
	maxW := NodeLayoutLength(node, css, "maxWidth", "max-width", viewportW, viewportW, viewportH, -1)
	minH := NodeLayoutLength(node, css, "minHeight", "min-height", viewportH, viewportW, viewportH, -1)
	maxH := NodeLayoutLength(node, css, "maxHeight", "max-height", viewportH, viewportW, viewportH, -1)

	if width >= 0 {
		if minW >= 0 && width < minW {
			width = minW
		}
		if maxW >= 0 && width > maxW {
			width = maxW
		}
	}
	if height >= 0 {
		if minH >= 0 && height < minH {
			height = minH
		}
		if maxH >= 0 && height > maxH {
			height = maxH
		}
	}

	return width, height
}

func nodePropString(node *Node, name string, fallback string) string {
	if node == nil {
		return fallback
	}
	if strings.EqualFold(strings.TrimSpace(name), "text") {
		if strings.TrimSpace(node.Text) != "" {
			return node.Text
		}
	}
	v := anyToString(node.GetProp(name), "")
	if v == "" {
		return fallback
	}
	return v
}

func nodePropInt(node *Node, name string, fallback int) int {
	if node == nil {
		return fallback
	}
	switch typed := node.GetProp(name).(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case float32:
		return int(typed)
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
			return parsed
		}
	}
	return fallback
}
