package render

import (
	"image"
	"image/color"
	"strconv"
	"strings"

	"gioui.org/op"

	coreengine "github.com/ArubikU/giocss/core/engine"
	coreinput "github.com/ArubikU/giocss/core/input"
	corelayout "github.com/ArubikU/giocss/core/layout"
	pkgcore "github.com/ArubikU/giocss/pkg"
	uicore "github.com/ArubikU/giocss/ui"
)

type selectOverlayInsets struct {
	PadTop       int
	PadRight     int
	PadBottom    int
	PadLeft      int
	BorderTop    int
	BorderRight  int
	BorderBottom int
	BorderLeft   int
}

func (in selectOverlayInsets) InsetX() int {
	return max(0, in.PadLeft) + max(0, in.PadRight) + max(0, in.BorderLeft) + max(0, in.BorderRight)
}

func (in selectOverlayInsets) InsetY() int {
	return max(0, in.PadTop) + max(0, in.PadBottom) + max(0, in.BorderTop) + max(0, in.BorderBottom)
}

func (in selectOverlayInsets) ContentRect(menuRect image.Rectangle) image.Rectangle {
	if menuRect.Empty() {
		return image.Rectangle{}
	}
	left := menuRect.Min.X + max(0, in.BorderLeft) + max(0, in.PadLeft)
	right := menuRect.Max.X - max(0, in.BorderRight) - max(0, in.PadRight)
	top := menuRect.Min.Y + max(0, in.BorderTop) + max(0, in.PadTop)
	bottom := menuRect.Max.Y - max(0, in.BorderBottom) - max(0, in.PadBottom)
	if right <= left {
		right = left + 1
	}
	if bottom <= top {
		bottom = top + 1
	}
	return image.Rect(left, top, right, bottom)
}

func DrawSelectDropdownOverlay(rctx *coreengine.RenderContext, host *GioRenderHost, state *GioWindowState, screenW, screenH int) {
	if state == nil || state.selectMenuOpen == "" {
		return
	}
	selectPath := state.selectMenuOpen
	selectProps, ok := state.propsForPath[selectPath]
	if !ok {
		state.selectMenuOpen = ""
		return
	}
	selectBounds, ok := state.boundsForPath[selectPath]
	if !ok || selectBounds.Empty() {
		state.selectMenuOpen = ""
		return
	}
	selectCSS := state.cssForPath[selectPath]
	selectModel := ResolveSelectModel(selectPath, selectProps, state.inputValues)
	if len(selectModel.Entries) == 0 {
		state.selectMenuOpen = ""
		return
	}

	gtx := rctx.Gtx
	backdropPath := selectPath + "__menu_backdrop"
	state.handlers[backdropPath] = func() {
		state.selectMenuOpen = ""
		invalidateHost(host)
	}
	registerGioEventArea(gtx.Ops, 0, 0, screenW, screenH, state.GetTag(backdropPath))

	menuCSS := selectOverlayMenuPaintCSS(selectCSS)
	optionH := selectOverlayOptionHeight(selectCSS, max(32, selectBounds.Dy()), state.frameViewW, state.frameViewH)
	baseMenuW := max(1, selectBounds.Dx())
	contentH := max(1, optionH*len(selectModel.Entries))
	insets := selectOverlayMenuInsetsFromCSS(menuCSS, baseMenuW, contentH, state.frameViewW, state.frameViewH)
	menuW := selectOverlayMenuWidth(menuCSS, baseMenuW, insets, state.frameViewW, state.frameViewH)
	insets = selectOverlayMenuInsetsFromCSS(menuCSS, menuW, contentH, state.frameViewW, state.frameViewH)
	menuH := selectOverlayMenuHeight(menuCSS, contentH, insets, state.frameViewW, state.frameViewH)

	constraintRect := selectOverlayConstraintRect(state, selectPath, screenW, screenH)
	menuRect := resolveSelectOverlayMenuRectWithSize(selectBounds, constraintRect, screenW, screenH, menuW, menuH)
	menuW = menuRect.Dx()
	menuH = menuRect.Dy()
	menuX := menuRect.Min.X
	menuY := menuRect.Min.Y
	insets = selectOverlayMenuInsetsFromCSS(menuCSS, menuW, menuH, state.frameViewW, state.frameViewH)
	contentRect := insets.ContentRect(menuRect)
	rowH := optionH
	if len(selectModel.Entries) > 0 {
		maxRowH := max(1, contentRect.Dy()/len(selectModel.Entries))
		if rowH > maxRowH {
			rowH = maxRowH
		}
	}
	radius := roundedFromProps(selectProps, menuCSS, menuW, menuH)
	radii := cssBorderRadiiValues(menuCSS, menuW, menuH)
	drawGioBoxShadow(gtx.Ops, menuX, menuY, menuW, menuH, radius, menuCSS)
	drawGioBackground(gtx.Ops, menuX, menuY, menuW, menuH, menuCSS)
	drawGioInsetBoxShadow(gtx.Ops, menuX, menuY, menuW, menuH, radius, menuCSS)
	drawGioElementBorder(gtx.Ops, menuX, menuY, menuW, menuH, radii, menuCSS)
	applySelectOverlayCursor(gtx.Ops, state, backdropPath, "default")

	for i, entry := range selectModel.Entries {
		optionPath := coreengine.BuildSelectOptionPath(selectPath, i)
		optionY := contentRect.Min.Y + i*rowH
		rowRect := image.Rect(contentRect.Min.X, optionY, contentRect.Max.X, optionY+rowH)
		optionProps := cloneAnyMap(entry.Props)
		if optionProps == nil {
			optionProps = map[string]any{}
		}
		optionProps["tag"] = "option"
		optionProps["text"] = entry.Label
		if entry.Disabled {
			optionProps["disabled"] = true
		}
		if i == selectModel.Index {
			optionProps["selected"] = true
		}

		hovered := selectOverlayOptionHovered(state.pointerKnown, state.pointerPos, rowRect)
		optionCSS := resolveSelectOverlayOptionCSS(state, selectPath, optionPath, selectProps, optionProps, selectCSS, selectModel, i, hovered)
		if hovered {
			applySelectOverlayCursor(gtx.Ops, state, optionPath, selectOverlayCursorValue(optionCSS, entry.Disabled))
		}
		rowX := contentRect.Min.X
		rowW := max(1, contentRect.Dx())
		rowCSS := selectOverlayOptionPaintCSS(optionCSS, i > 0, i == selectModel.Index)
		rowRadius := roundedFromProps(optionProps, rowCSS, rowW, rowH)
		rowRadii := cssBorderRadiiValues(rowCSS, rowW, rowH)
		drawGioBoxShadow(gtx.Ops, rowX, optionY, rowW, rowH, rowRadius, rowCSS)
		drawGioBackground(gtx.Ops, rowX, optionY, rowW, rowH, rowCSS)
		drawGioInsetBoxShadow(gtx.Ops, rowX, optionY, rowW, rowH, rowRadius, rowCSS)
		drawGioElementBorder(gtx.Ops, rowX, optionY, rowW, rowH, rowRadii, rowCSS)

		fg := toNRGBA(uicore.ApplyCSSOpacity(uicore.ParseHexColor(uicore.CSSGetColor(optionCSS, "color", "#17344f"), color.NRGBA{R: 0x17, G: 0x34, B: 0x4F, A: 0xFF}), optionCSS))
		textX, textY, textW, textH := selectOverlayOptionTextRect(optionCSS, rowX, optionY, rowW, rowH)
		textPlan := BuildTextContentPlan(optionProps, optionCSS, textW, 13)
		align := pkgcore.CSSTextAlign(optionCSS)
		drawGioText(gtx, state, textX, textY, textW, textH, textPlan.Text, fg, textPlan.FontSize, textPlan.Bold, textPlan.Italic, textPlan.UseMono, align, textPlan.MaxLines, textPlan.WrapAllowed)
		drawGioTextDecorations(gtx, textX, textY, textW, textH, textPlan.FontSize, align, textPlan.Text, fg, uicore.ParseTextDecoration(optionCSS))

		if entry.Disabled {
			continue
		}
		selectedValue := entry.Value
		if strings.TrimSpace(selectedValue) == "" {
			selectedValue = entry.Label
		}
		registerSelectOptionHandler(state, host, optionPath, selectPath, selectModel.Key, coreengine.ResolveComponentEventName(selectProps, "change"), i, selectedValue)
		state.propsForPath[optionPath] = optionProps
		state.cssForPath[optionPath] = optionCSS
		registerGioEventArea(gtx.Ops, rowX, optionY, rowW, rowH, state.GetTag(optionPath))
	}
}

func resolveSelectOverlayOptionCSS(state *GioWindowState, selectPath, optionPath string, selectProps map[string]any, optionProps map[string]any, selectCSS map[string]string, selectModel SelectModel, optionIndex int, hovered bool) map[string]string {
	ss := state.frameStyleSheet
	css := uicore.ResolveStyle(optionProps, ss, state.frameViewW)
	css = corelayout.MergeInheritedTextCSS(css, selectCSS)
	if ss != nil {
		hovered = hovered || (state.frameHoverState != nil && state.frameHoverState[optionPath])
		active := state.frameActiveState != nil && state.frameActiveState[optionPath]
		disabled := propTruthy(optionProps["disabled"], "disabled")
		selected := optionIndex == selectModel.Index
		uicore.ApplyHoverStyles(css, optionProps, ss, hovered)
		uicore.ApplyActiveStyles(css, optionProps, ss, active)
		uicore.ApplyDisabledStyles(css, optionProps, ss, disabled)
		if selected {
			uicore.ApplyCheckedStyles(css, optionProps, ss, true)
		}
		if ss.HasAdvancedSelectors() {
			uicore.ApplyAdvancedSelectorStyles(css, ss, uicore.AdvancedSelectorContext{
				Path:  optionPath,
				Props: optionProps,
				LookupProps: func(target string) (map[string]any, bool) {
					switch {
					case target == optionPath:
						return optionProps, true
					case target == selectPath:
						return selectProps, true
					default:
						props, ok := state.propsForPath[target]
						return props, ok
					}
				},
				ParentPath: func(target string) string {
					if strings.HasPrefix(target, selectPath+"__opt/") {
						return selectPath
					}
					return parentPath(target)
				},
				PreviousSiblingPath: func(target string) string {
					paths := syntheticSelectOptionSiblings(selectPath, target)
					if len(paths) == 0 {
						return ""
					}
					return paths[0]
				},
				PreviousSiblingPaths: func(target string) []string {
					return syntheticSelectOptionSiblings(selectPath, target)
				},
				PseudoState: func(target string, targetProps map[string]any, pseudo string) bool {
					if strings.HasPrefix(target, selectPath+"__opt/") {
						switch pseudo {
						case "hover":
							return state.frameHoverState != nil && state.frameHoverState[target]
						case "active":
							return state.frameActiveState != nil && state.frameActiveState[target]
						case "disabled":
							return propTruthy(targetProps["disabled"], "disabled")
						case "checked":
							return target == optionPath && selected
						default:
							return false
						}
					}
					return selectorPseudoState(state, target, targetProps, pseudo)
				},
			})
		}
	}
	return css
}

func selectOverlayConstraintRect(state *GioWindowState, selectPath string, screenW, screenH int) image.Rectangle {
	fallback := image.Rect(0, 0, screenW, screenH)
	if state == nil {
		return fallback
	}
	for parent := parentPath(selectPath); parent != ""; parent = parentPath(parent) {
		rect, ok := state.boundsForPath[parent]
		if !ok || rect.Empty() {
			continue
		}
		constrained := rect.Intersect(fallback)
		if constrained.Empty() {
			continue
		}
		return constrained
	}
	return fallback
}

func resolveSelectOverlayMenuRect(selectBounds, constraintRect image.Rectangle, screenW, screenH, optionH, optionCount int) image.Rectangle {
	menuW := max(1, selectBounds.Dx())
	menuH := max(1, optionH*optionCount)
	return resolveSelectOverlayMenuRectWithSize(selectBounds, constraintRect, screenW, screenH, menuW, menuH)
}

func resolveSelectOverlayMenuRectWithSize(selectBounds, constraintRect image.Rectangle, screenW, screenH, menuW, menuH int) image.Rectangle {
	if constraintRect.Empty() {
		constraintRect = image.Rect(0, 0, screenW, screenH)
	}
	minX := max(0, constraintRect.Min.X)
	maxX := min(screenW, constraintRect.Max.X)
	minY := max(0, constraintRect.Min.Y)
	maxY := min(screenH, constraintRect.Max.Y)
	if maxX <= minX {
		minX, maxX = 0, screenW
	}
	if maxY <= minY {
		minY, maxY = 0, screenH
	}

	menuW = max(1, menuW)
	menuH = max(1, menuH)
	availableW := max(1, maxX-minX)
	if menuW > availableW {
		menuW = availableW
	}

	menuX := selectBounds.Min.X
	if menuX < minX {
		menuX = minX
	}
	if menuX+menuW > maxX {
		menuX = max(minX, maxX-menuW)
	}

	menuY := selectBounds.Max.Y + 4
	if menuY+menuH > maxY && selectBounds.Min.Y-menuH-4 >= minY {
		menuY = selectBounds.Min.Y - menuH - 4
	}
	if menuY+menuH > maxY {
		menuY = max(minY, maxY-menuH)
	}
	if menuY < minY {
		menuY = minY
	}

	return image.Rect(menuX, menuY, menuX+menuW, menuY+menuH)
}

func selectOverlayMenuInsetsFromCSS(css map[string]string, w, h, viewportW, viewportH int) selectOverlayInsets {
	plan := BuildElementBorderPlan(css, w, h)
	basisW := max(1, w)
	basisH := max(1, h)
	return selectOverlayInsets{
		PadTop:       uicore.CSSLengthValue(css["padding-top"], 0, basisH, viewportW, viewportH),
		PadRight:     uicore.CSSLengthValue(css["padding-right"], 0, basisW, viewportW, viewportH),
		PadBottom:    uicore.CSSLengthValue(css["padding-bottom"], 0, basisH, viewportW, viewportH),
		PadLeft:      uicore.CSSLengthValue(css["padding-left"], 0, basisW, viewportW, viewportH),
		BorderTop:    plan.Top.Width,
		BorderRight:  plan.Right.Width,
		BorderBottom: plan.Bottom.Width,
		BorderLeft:   plan.Left.Width,
	}
}

func selectOverlayMenuWidth(css map[string]string, fallback int, insets selectOverlayInsets, viewportW, viewportH int) int {
	w := max(1, fallback)
	basis := max(1, fallback)
	if raw := strings.TrimSpace(css["width"]); raw != "" {
		w = max(1, uicore.CSSLengthValue(raw, w, basis, viewportW, viewportH))
	}
	if raw := strings.TrimSpace(css["min-width"]); raw != "" {
		w = max(w, max(1, uicore.CSSLengthValue(raw, w, basis, viewportW, viewportH)))
	}
	if raw := strings.TrimSpace(css["max-width"]); raw != "" {
		w = min(w, max(1, uicore.CSSLengthValue(raw, w, basis, viewportW, viewportH)))
	}
	return max(1+insets.InsetX(), w)
}

func selectOverlayMenuHeight(css map[string]string, contentH int, insets selectOverlayInsets, viewportW, viewportH int) int {
	h := max(1, contentH+insets.InsetY())
	basis := max(1, h)
	if raw := strings.TrimSpace(css["height"]); raw != "" {
		h = max(1, uicore.CSSLengthValue(raw, h, basis, viewportW, viewportH))
	}
	if raw := strings.TrimSpace(css["min-height"]); raw != "" {
		h = max(h, max(1, uicore.CSSLengthValue(raw, h, basis, viewportW, viewportH)))
	}
	if raw := strings.TrimSpace(css["max-height"]); raw != "" {
		h = min(h, max(1, uicore.CSSLengthValue(raw, h, basis, viewportW, viewportH)))
	}
	return max(1+insets.InsetY(), h)
}

func selectOverlayOptionHovered(pointerKnown bool, pointerPos image.Point, rowRect image.Rectangle) bool {
	return pointerKnown && !rowRect.Empty() && pointerPos.In(rowRect)
}

func selectOverlayOptionHeight(css map[string]string, fallback, viewportW, viewportH int) int {
	if fallback <= 0 {
		fallback = 32
	}
	fontSize := uicore.CSSFontSize(css, 13)
	lineH := uicore.CSSLineHeightPx(css, fontSize)
	padT := uicore.CSSLengthValue(css["padding-top"], 0, fallback, viewportW, viewportH)
	padB := uicore.CSSLengthValue(css["padding-bottom"], 0, fallback, viewportW, viewportH)
	borderT := uicore.CSSLengthValue(css["border-top-width"], 0, fallback, viewportW, viewportH)
	borderB := uicore.CSSLengthValue(css["border-bottom-width"], 0, fallback, viewportW, viewportH)

	h := max(fallback, lineH+max(0, padT)+max(0, padB)+max(0, borderT)+max(0, borderB))
	if raw := strings.TrimSpace(css["height"]); raw != "" {
		h = max(1, uicore.CSSLengthValue(raw, h, fallback, viewportW, viewportH))
	}
	if raw := strings.TrimSpace(css["min-height"]); raw != "" {
		h = max(h, max(1, uicore.CSSLengthValue(raw, h, fallback, viewportW, viewportH)))
	}
	if raw := strings.TrimSpace(css["max-height"]); raw != "" {
		h = min(h, max(1, uicore.CSSLengthValue(raw, h, fallback, viewportW, viewportH)))
	}
	return max(1, h)
}

func selectOverlayOptionTextRect(css map[string]string, x, y, w, h int) (int, int, int, int) {
	if w <= 0 || h <= 0 {
		return x, y, 1, 1
	}
	padL := uicore.CSSLengthValue(css["padding-left"], 10, w, w, h)
	padR := uicore.CSSLengthValue(css["padding-right"], 12, w, w, h)
	padT := uicore.CSSLengthValue(css["padding-top"], 0, h, w, h)
	padB := uicore.CSSLengthValue(css["padding-bottom"], 0, h, w, h)

	padL = min(max(0, padL), max(0, w-1))
	padR = min(max(0, padR), max(0, w-1))
	padT = min(max(0, padT), max(0, h-1))
	padB = min(max(0, padB), max(0, h-1))

	textX := x + padL
	textY := y + padT
	textW := max(1, w-padL-padR)
	textH := max(1, h-padT-padB)
	return textX, textY, textW, textH
}

func selectOverlayMenuPaintCSS(selectCSS map[string]string) map[string]string {
	css := cloneStringMapLocal(selectCSS)
	if css == nil {
		css = map[string]string{}
	}
	if strings.TrimSpace(selectOverlayBackgroundColor(css, "")) == "" {
		css["background-color"] = "#ffffff"
	}
	if !BuildElementBorderPlan(css, 120, 40).HasAny {
		css["border-width"] = "1px"
		css["border-style"] = "solid"
		css["border-color"] = "#bfd4e6"
	}
	return css
}

func selectOverlayOptionPaintCSS(optionCSS map[string]string, addDivider bool, selected bool) map[string]string {
	css := cloneStringMapLocal(optionCSS)
	if css == nil {
		css = map[string]string{}
	}
	if selected && strings.TrimSpace(selectOverlayBackgroundColor(css, "")) == "" {
		css["background-color"] = "#e7f2fb"
	}
	if addDivider && !selectOverlayHasExplicitTopBorder(css) {
		css["border-top-width"] = "1px"
		if strings.TrimSpace(css["border-top-style"]) == "" {
			css["border-top-style"] = "solid"
		}
		if strings.TrimSpace(css["border-top-color"]) == "" {
			css["border-top-color"] = "#e2ecf5"
		}
	}
	return css
}

func selectOverlayHasExplicitTopBorder(css map[string]string) bool {
	if css == nil {
		return false
	}
	for _, key := range []string{
		"border", "border-width", "border-style", "border-color",
		"border-top", "border-top-width", "border-top-style", "border-top-color",
	} {
		if strings.TrimSpace(css[key]) != "" {
			return true
		}
	}
	return false
}

func cloneStringMapLocal(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func selectOverlayBackgroundColor(css map[string]string, fallback string) string {
	if v := strings.TrimSpace(uicore.CSSGetColor(css, "background-color", "")); v != "" {
		return v
	}
	if v := strings.TrimSpace(uicore.CSSBackground(css)); v != "" {
		return v
	}
	return fallback
}

func selectOverlayCursorValue(optionCSS map[string]string, disabled bool) string {
	if disabled {
		return "not-allowed"
	}
	if optionCSS != nil {
		cursorRaw := strings.TrimSpace(uicore.ParseCursor(optionCSS))
		if cursorRaw != "" && cursorRaw != "auto" {
			return cursorRaw
		}
	}
	return "default"
}

func applySelectOverlayCursor(ops *op.Ops, state *GioWindowState, path, cursorRaw string) {
	if state == nil || strings.TrimSpace(cursorRaw) == "" {
		return
	}
	state.frameCursorPath = path
	state.frameCursorValue = cursorRaw
	coreinput.CSSCursorToGio(cursorRaw).Add(ops)
}

func syntheticSelectOptionSiblings(selectPath, optionPath string) []string {
	prefix := selectPath + "__opt/"
	if !strings.HasPrefix(optionPath, prefix) {
		return nil
	}
	idx := strings.LastIndex(optionPath, "/")
	if idx < 0 || idx+1 >= len(optionPath) {
		return nil
	}
	current, err := strconv.Atoi(optionPath[idx+1:])
	if err != nil || current <= 0 {
		return nil
	}
	out := make([]string, 0, current)
	for i := current - 1; i >= 0; i-- {
		out = append(out, coreengine.BuildSelectOptionPath(selectPath, i))
	}
	return out
}

func cloneAnyMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
