package render

import (
	"image"
	"image/color"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"

	coredebug "github.com/ArubikU/giocss/core/debug"
	coreengine "github.com/ArubikU/giocss/core/engine"
	coreinput "github.com/ArubikU/giocss/core/input"
	corelayout "github.com/ArubikU/giocss/core/layout"
	pkgcore "github.com/ArubikU/giocss/pkg"
	uicore "github.com/ArubikU/giocss/ui"
)

var (
	renderGPUEnabled = true
	renderLowPower   = false
)

func initRenderConfig() {
	uicore.RenderConfigOnce.Do(func() {
		gpuEnv := strings.TrimSpace(strings.ToLower(os.Getenv("POLYLOFT_GIO_GPU")))
		if gpuEnv == "0" || gpuEnv == "false" || gpuEnv == "off" {
			renderGPUEnabled = false
		}
		lowEnv := strings.TrimSpace(strings.ToLower(os.Getenv("POLYLOFT_GIO_LOW_POWER")))
		if lowEnv == "1" || lowEnv == "true" || lowEnv == "on" {
			renderLowPower = true
		}
	})
}

func InitRenderConfig() {
	initRenderConfig()
}

func RenderGPUEnabled() bool {
	initRenderConfig()
	return renderGPUEnabled
}

func RenderLowPowerEnabled() bool {
	initRenderConfig()
	return renderLowPower
}

func backgroundLayers(raw string) []string {
	return BackgroundLayers(raw)
}

type PointerTag struct {
	path string
}

type GioRenderHost struct {
	HoverState           map[string]bool
	ActiveState          map[string]bool
	StyleSheet           *uicore.StyleSheet
	RenderStore          *uicore.RenderStore
	DispatchEvent        func(string, map[string]any) error
	EmitRuntimeError     func(error)
	Invalidate           func()
	RecordDebugComponent func(path string, kind string, props map[string]any, w int, h int, totalDur time.Duration, selfDur time.Duration)
	CanUseWindowEffects  bool
}

type GioWindowState struct {
	shaper             *text.Shaper
	tags               map[string]*PointerTag
	handlers           map[string]func()
	propsForPath       map[string]map[string]any
	cssForPath         map[string]map[string]string
	scrollOffsets      map[string]image.Point
	scrollCarry        map[string]f32.Point
	scrollHits         map[string]coreinput.ScrollHitInfo
	scrollFocusPath    string
	scrollDragPath     string
	scrollDragAxis     string
	scrollDragGrab     int
	pointerCapturePath string
	boundsForPath      map[string]image.Rectangle
	lastPointer        map[string]image.Point
	pointerPos         image.Point
	pointerKnown       bool
	editors            map[string]*widget.Editor
	inputValues        map[string]string
	inputExternal      map[string]string
	inputScrollX       map[string]int
	inputContentW      map[string]int
	inputVisibleW      map[string]int
	sliderValues       map[string]float64
	boolValues         map[string]bool
	inputFocused       map[string]bool
	focusedInputPath   string
	selectMenuOpen     string
	pickerModalOpen    string
	pickerType         string
	pickerValue        string
	frameViewW         int
	frameViewH         int
	frameViewportRect  image.Rectangle
	frameViewport48    image.Rectangle
	frameViewport96    image.Rectangle
	frameHoverState    map[string]bool
	frameActiveState   map[string]bool
	frameDebug         bool
	frameStyleSheet    *uicore.StyleSheet
	frameNumber        int64
	profileSampleFrame bool
	frameCursorPath    string
	frameCursorValue   string
	profileComponents  bool
	profileFull        bool
	resolvedCSS        map[string]uicore.ResolvedCSSCacheEntry
	zChildrenHint      map[string]uicore.ZChildrenHintCacheEntry
}

func NewGioWindowState() *GioWindowState {
	return &GioWindowState{
		shaper:           uicore.NewGioShaper(),
		tags:             make(map[string]*PointerTag),
		handlers:         make(map[string]func()),
		propsForPath:     make(map[string]map[string]any),
		cssForPath:       make(map[string]map[string]string),
		scrollOffsets:    make(map[string]image.Point),
		scrollCarry:      make(map[string]f32.Point),
		scrollHits:       make(map[string]coreinput.ScrollHitInfo),
		boundsForPath:    make(map[string]image.Rectangle),
		lastPointer:      make(map[string]image.Point),
		editors:          make(map[string]*widget.Editor),
		inputValues:      make(map[string]string),
		inputExternal:    make(map[string]string),
		inputScrollX:     make(map[string]int),
		inputContentW:    make(map[string]int),
		inputVisibleW:    make(map[string]int),
		sliderValues:     make(map[string]float64),
		boolValues:       make(map[string]bool),
		inputFocused:     make(map[string]bool),
		frameHoverState:  make(map[string]bool),
		frameActiveState: make(map[string]bool),
		resolvedCSS:      make(map[string]uicore.ResolvedCSSCacheEntry),
		zChildrenHint:    make(map[string]uicore.ZChildrenHintCacheEntry),
	}
}

func (gs *GioWindowState) BeginFrame(size image.Point, frameNumber int64, hoverState map[string]bool, activeState map[string]bool, debug bool, ss *uicore.StyleSheet, profileComponents bool, profileFull bool) {
	gs.frameViewW = size.X
	gs.frameViewH = size.Y
	gs.frameViewportRect = image.Rect(0, 0, size.X, size.Y)
	gs.frameViewport48 = coreengine.ExpandedViewportRect(size.X, size.Y, 48)
	gs.frameViewport96 = coreengine.ExpandedViewportRect(size.X, size.Y, 96)
	gs.frameNumber = frameNumber
	gs.profileSampleFrame = frameNumber%30 == 0
	if frameNumber%120 == 0 {
		gs.PurgeStaleCaches(frameNumber)
	}
	gs.frameHoverState = coreengine.RefreshBoolMap(gs.frameHoverState, hoverState)
	gs.frameActiveState = coreengine.RefreshBoolMap(gs.frameActiveState, activeState)
	gs.frameDebug = debug
	gs.frameStyleSheet = ss
	gs.profileComponents = profileComponents
	gs.profileFull = profileFull
}

func (gs *GioWindowState) PurgeStaleCaches(frame int64) {
	uicore.PurgeStaleRenderCaches(gs.resolvedCSS, gs.zChildrenHint, frame)
}

func (gs *GioWindowState) GetTag(path string) *PointerTag {
	if tag, ok := gs.tags[path]; ok {
		return tag
	}
	tag := &PointerTag{path: path}
	gs.tags[path] = tag
	return tag
}

func (gs *GioWindowState) PrepareRenderFrame() {
	clear(gs.handlers)
	clear(gs.propsForPath)
	clear(gs.cssForPath)
	clear(gs.scrollHits)
	clear(gs.boundsForPath)
	gs.frameCursorPath = ""
	gs.frameCursorValue = ""
}

func mayContainOutOfFlowChildren(children []any) bool {
	childMaps := make([]map[string]any, 0, len(children))
	for _, child := range children {
		childMaps = append(childMaps, anyToMap(child))
	}
	return ChildrenMayContainOutOfFlow(childMaps)
}

func anyToMap(candidate any) map[string]any {
	if mapped, ok := candidate.(map[string]any); ok {
		return mapped
	}
	return map[string]any{}
}

func anyToSlice(candidate any) []any {
	if values, ok := candidate.([]any); ok {
		return values
	}
	return nil
}

func anyToInt(candidate any, fallback int) int {
	switch typed := candidate.(type) {
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

func childSliceSignatureAny(children []any) uint64 {
	if len(children) == 0 {
		return 0
	}
	childMaps := make([]map[string]any, 0, len(children))
	for _, child := range children {
		childMaps = append(childMaps, anyToMap(child))
	}
	return coredebug.ChildSliceSignature(childMaps)
}

func stylePropSignature(kind string, props map[string]any) string {
	if len(props) == 0 {
		return strings.TrimSpace(kind)
	}
	keys := []string{"tag", "class", "id", "style", "type", "inputtype"}
	var b strings.Builder
	b.Grow(128)
	b.WriteString(strings.TrimSpace(kind))
	for _, k := range keys {
		v := strings.TrimSpace(anyToString(props[k], ""))
		if v == "" {
			continue
		}
		b.WriteByte('|')
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v)
	}
	return b.String()
}

func propTruthy(raw any, attrName string) bool {
	attrName = strings.ToLower(strings.TrimSpace(attrName))
	switch typed := raw.(type) {
	case bool:
		return typed
	case int:
		return typed != 0
	case int64:
		return typed != 0
	case float64:
		return typed != 0
	case float32:
		return typed != 0
	case string:
		v := strings.ToLower(strings.TrimSpace(typed))
		if v == "" {
			// HTML-like boolean attrs may appear as present-empty.
			return true
		}
		switch v {
		case "1", "true", "yes", "on", attrName:
			return true
		default:
			return false
		}
	default:
		return raw != nil
	}
}

func isControlDisabledProps(props map[string]any) bool {
	if props == nil {
		return false
	}
	return propTruthy(props["disabled"], "disabled")
}

func resolveCheckedState(path string, props map[string]any, state *GioWindowState) bool {
	if props == nil {
		return false
	}
	tag := strings.ToLower(strings.TrimSpace(anyToString(props["tag"], "")))
	inputType := strings.ToLower(strings.TrimSpace(anyToString(props["type"], anyToString(props["inputtype"], ""))))

	if inputType == "checkbox" || inputType == "check" {
		if state != nil {
			if checked, seen := state.boolValues[path]; seen {
				return checked
			}
		}
		return propTruthy(props["checked"], "checked")
	}

	if inputType == "radio" {
		groupName := anyToString(props["name"], path)
		groupKey := "radio:" + groupName
		myValue := anyToString(props["value"], path)
		if state != nil {
			if selected := strings.TrimSpace(state.inputValues[groupKey]); selected != "" {
				return selected == myValue
			}
		}
		return propTruthy(props["checked"], "checked")
	}

	if tag == "option" {
		return propTruthy(props["selected"], "selected") || propTruthy(props["checked"], "checked")
	}

	return propTruthy(props["checked"], "checked")
}

func resolveControlValue(path string, props map[string]any, state *GioWindowState) string {
	if state != nil {
		if v, ok := state.inputValues[path]; ok {
			return strings.TrimSpace(v)
		}
	}
	return strings.TrimSpace(coreinput.InputExternalValue(props))
}

func resolveInvalidState(path string, props map[string]any, state *GioWindowState, checked bool, disabled bool) bool {
	if props == nil || disabled {
		return false
	}
	tag := strings.ToLower(strings.TrimSpace(anyToString(props["tag"], "")))
	inputType := strings.ToLower(strings.TrimSpace(anyToString(props["type"], anyToString(props["inputtype"], ""))))
	component := strings.ToLower(strings.TrimSpace(anyToString(props["component"], anyToString(props["native"], ""))))
	required := propTruthy(props["required"], "required")
	readonly := propTruthy(props["readonly"], "readonly")

	if readonly {
		return false
	}

	if inputType == "checkbox" || inputType == "check" {
		return required && !checked
	}

	if inputType == "radio" {
		if !required {
			return false
		}
		groupName := anyToString(props["name"], path)
		groupKey := "radio:" + groupName
		if state != nil {
			return strings.TrimSpace(state.inputValues[groupKey]) == ""
		}
		return !checked
	}

	value := resolveControlValue(path, props, state)
	if required && value == "" {
		return true
	}
	if value == "" {
		return false
	}

	switch inputType {
	case "number", "date", "time":
		return coreinput.NormalizeTypedInputValue(inputType, value, props, true) == ""
	case "email":
		at := strings.Index(value, "@")
		if at <= 0 || at >= len(value)-1 {
			return true
		}
		domain := value[at+1:]
		return !strings.Contains(domain, ".")
	case "url":
		lower := strings.ToLower(value)
		return !(strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://"))
	}

	if tag == "select" || component == "select" || component == "dropdown" {
		if !required {
			return false
		}
		if state == nil {
			return value == ""
		}
		model := ResolveSelectModel(path, props, state.inputValues)
		selectedValue := strings.TrimSpace(model.SelectedValue)
		selectedLabel := strings.TrimSpace(model.SelectedLabel)
		if model.SelectedDisabled {
			return true
		}
		placeholderValue := strings.TrimSpace(anyToString(props["emptyvalue"], anyToString(props["placeholdervalue"], "")))
		placeholder := strings.TrimSpace(anyToString(props["placeholder"], anyToString(props["emptylabel"], "")))
		if selectedValue == "" {
			return true
		}
		if placeholderValue != "" && strings.EqualFold(selectedValue, placeholderValue) {
			return true
		}
		if placeholder != "" && strings.EqualFold(selectedLabel, placeholder) {
			return true
		}
		return false
	}

	return false
}

func hasSelectOptions(raw any) bool {
	switch typed := raw.(type) {
	case []any:
		return len(typed) > 0
	case string:
		return strings.TrimSpace(typed) != ""
	default:
		return false
	}
}

func extractSelectOptionsFromChildren(children []any) []any {
	if len(children) == 0 {
		return nil
	}
	options := make([]any, 0, len(children))
	for _, child := range children {
		childMap := anyToMap(child)
		if strings.ToLower(strings.TrimSpace(anyToString(childMap["kind"], ""))) != "option" {
			continue
		}
		childProps := anyToMap(childMap["props"])
		label := strings.TrimSpace(anyToString(childProps["text"], anyToString(childProps["label"], anyToString(childProps["value"], ""))))
		if label == "" {
			continue
		}
		value := strings.TrimSpace(anyToString(childProps["value"], label))
		entry := map[string]any{
			"label": label,
			"value": value,
		}
		if propTruthy(childProps["disabled"], "disabled") {
			entry["disabled"] = true
		}
		if propTruthy(childProps["selected"], "selected") {
			entry["selected"] = true
		}
		options = append(options, entry)
	}
	if len(options) == 0 {
		return nil
	}
	return options
}

func dispatchHostEvent(host *GioRenderHost, eventName string, payload map[string]any) error {
	if host == nil || host.DispatchEvent == nil || strings.TrimSpace(eventName) == "" {
		return nil
	}
	return host.DispatchEvent(eventName, payload)
}

func invalidateHost(host *GioRenderHost) {
	if host != nil && host.Invalidate != nil {
		host.Invalidate()
	}
}

func registerCheckboxToggleHandler(state *GioWindowState, host *GioRenderHost, path, eventName string) {
	state.handlers[path] = func() {
		payload := coreengine.ToggleCheckboxState(path, state.boolValues, time.Now().UnixMilli())
		_ = dispatchHostEvent(host, eventName, payload)
		invalidateHost(host)
	}
}

func registerRadioSelectHandler(state *GioWindowState, host *GioRenderHost, path, groupKey, radioValue, eventName string) {
	state.handlers[path] = func() {
		payload := coreengine.SelectRadioState(path, groupKey, radioValue, state.inputValues, time.Now().UnixMilli())
		_ = dispatchHostEvent(host, eventName, payload)
		invalidateHost(host)
	}
}

func registerSliderPointerHandler(state *GioWindowState, host *GioRenderHost, path string, minV, maxV float64, eventName, component string) {
	state.handlers[path] = func() {
		payload, ok := coreengine.ResolveSliderPointerState(path, state.boundsForPath, state.lastPointer, state.sliderValues, minV, maxV, component, time.Now().UnixMilli())
		if !ok {
			return
		}
		_ = dispatchHostEvent(host, eventName, payload)
		invalidateHost(host)
	}
}

func dispatchPointerEventForProps(host *GioRenderHost, path string, props map[string]any, kind string, pos image.Point, delta image.Point, pressed bool, timestamp int64) bool {
	if props == nil {
		return false
	}
	eventName := ""
	dragging := false
	switch kind {
	case "pointerdown":
		eventName = coreengine.ResolveOnPointerDownEvent(props)
	case "pointermove":
		eventName = coreengine.ResolveOnPointerMoveEvent(props)
	case "pointerdrag":
		eventName = coreengine.ResolveOnPointerMoveEvent(props)
		dragging = true
	case "pointerup":
		eventName = coreengine.ResolveOnPointerUpEvent(props)
	}
	if eventName == "" {
		return false
	}
	payload := coreengine.BuildPointerEventPayload(path, kind, pos.X, pos.Y, delta.X, delta.Y, dragging, pressed, timestamp)
	if err := dispatchHostEvent(host, eventName, payload); err != nil {
		if host != nil && host.EmitRuntimeError != nil {
			host.EmitRuntimeError(err)
		}
		return false
	}
	return true
}

func resolveCapturedPointerPath(state *GioWindowState, currentPath string, kind pointer.Kind) string {
	if state == nil || state.pointerCapturePath == "" {
		return currentPath
	}
	switch kind {
	case pointer.Move, pointer.Drag, pointer.Release:
		if state.pointerCapturePath != currentPath {
			return state.pointerCapturePath
		}
	}
	return currentPath
}

func registerSelectCycleHandler(state *GioWindowState, host *GioRenderHost, path, key, eventName string, options []string, values []string, enabledIndexes []int, currentIndex int) {
	state.handlers[path] = func() {
		payload, ok := coreengine.CycleSelectState(path, key, state.inputValues, options, values, enabledIndexes, currentIndex, time.Now().UnixMilli())
		if !ok {
			return
		}
		if indexRaw, ok := payload["index"]; ok {
			switch indexVal := indexRaw.(type) {
			case float64:
				state.inputValues[key] = strconv.Itoa(int(indexVal))
			case int:
				state.inputValues[key] = strconv.Itoa(indexVal)
			}
		}
		if selectedValue, ok := payload["value"].(string); ok {
			state.inputValues[path] = selectedValue
		}
		_ = dispatchHostEvent(host, eventName, payload)
		invalidateHost(host)
	}
}

func registerSelectToggleHandler(state *GioWindowState, host *GioRenderHost, path string) {
	state.handlers[path] = func() {
		if state.selectMenuOpen == path {
			state.selectMenuOpen = ""
		} else {
			state.selectMenuOpen = path
		}
		invalidateHost(host)
	}
}

func registerSelectOptionHandler(state *GioWindowState, host *GioRenderHost, optionPath, selectPath, key, eventName string, index int, selectedValue string) {
	state.handlers[optionPath] = func() {
		state.inputValues[key] = strconv.Itoa(index)
		state.inputValues[selectPath] = selectedValue
		state.selectMenuOpen = ""
		payload := coreengine.BuildSelectEventPayload(selectPath, selectedValue, index, time.Now().UnixMilli())
		_ = dispatchHostEvent(host, eventName, payload)
		invalidateHost(host)
	}
}

func parentPath(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx <= 0 {
		return ""
	}
	return path[:idx]
}

func shouldCaptureSelectOverlayEvent(state *GioWindowState, path string) bool {
	if state == nil || state.selectMenuOpen == "" {
		return false
	}
	selectPath := state.selectMenuOpen
	if path == selectPath || path == selectPath+"__menu_backdrop" {
		return false
	}
	return !strings.HasPrefix(path, selectPath+"__opt/")
}

func previousSiblingPath(path string, state *GioWindowState) string {
	if state == nil {
		return ""
	}
	parent := parentPath(path)
	idx := strings.LastIndex(path, "/")
	if parent == "" || idx < 0 || idx+1 >= len(path) {
		return ""
	}
	childIndex, err := strconv.Atoi(path[idx+1:])
	if err != nil {
		return ""
	}
	for i := childIndex - 1; i >= 0; i-- {
		sibling := coreengine.ChildPath(parent, i)
		if _, ok := state.propsForPath[sibling]; ok {
			return sibling
		}
	}
	return ""
}

func previousSiblingPaths(path string, state *GioWindowState) []string {
	if state == nil {
		return nil
	}
	parent := parentPath(path)
	idx := strings.LastIndex(path, "/")
	if parent == "" || idx < 0 || idx+1 >= len(path) {
		return nil
	}
	childIndex, err := strconv.Atoi(path[idx+1:])
	if err != nil || childIndex <= 0 {
		return nil
	}
	out := make([]string, 0, childIndex)
	for i := childIndex - 1; i >= 0; i-- {
		sibling := coreengine.ChildPath(parent, i)
		if _, ok := state.propsForPath[sibling]; ok {
			out = append(out, sibling)
		}
	}
	return out
}

func selectorPseudoState(state *GioWindowState, path string, props map[string]any, pseudo string) bool {
	if state == nil {
		return false
	}
	switch pseudo {
	case "hover":
		return state.frameHoverState != nil && state.frameHoverState[path]
	case "active":
		return state.frameActiveState != nil && state.frameActiveState[path]
	case "focus":
		return state.inputFocused[path] || state.focusedInputPath == path
	case "disabled":
		return isControlDisabledProps(props)
	case "checked":
		return resolveCheckedState(path, props, state)
	case "invalid":
		checked := resolveCheckedState(path, props, state)
		disabled := isControlDisabledProps(props)
		return resolveInvalidState(path, props, state, checked, disabled)
	default:
		return false
	}
}

func applyInlineStyleOverrides(css map[string]string, props map[string]any) {
	for k, v := range props {
		if !strings.HasPrefix(k, "style.") {
			continue
		}
		prop := uicore.CanonicalName(strings.ToLower(strings.TrimPrefix(k, "style.")))
		if s := anyToString(v, ""); s != "" {
			css[prop] = s
		}
	}
	uicore.CSSExpandBoxShorthand(css, "padding")
	uicore.CSSExpandBoxShorthand(css, "margin")
	uicore.CSSExpandBorderShorthand(css)
	uicore.CSSResolveVariables(css)
}

func resolveFormIDFromProps(formProps map[string]any) string {
	if formProps == nil {
		return ""
	}
	return strings.TrimSpace(anyToString(formProps["id"], ""))
}

func controlReferencesFormID(props map[string]any, formID string) bool {
	if props == nil || formID == "" {
		return false
	}
	controlFormID := strings.TrimSpace(anyToString(props["form"], anyToString(props["form-id"], "")))
	if controlFormID == "" {
		return false
	}
	return controlFormID == formID
}

func isFieldsetNodeProps(props map[string]any) bool {
	if props == nil {
		return false
	}
	tag := strings.ToLower(strings.TrimSpace(anyToString(props["tag"], "")))
	if tag == "fieldset" {
		return true
	}
	component := strings.ToLower(strings.TrimSpace(anyToString(props["component"], anyToString(props["native"], ""))))
	return component == "fieldset"
}

func isLegendNodeProps(props map[string]any) bool {
	if props == nil {
		return false
	}
	tag := strings.ToLower(strings.TrimSpace(anyToString(props["tag"], "")))
	if tag == "legend" {
		return true
	}
	component := strings.ToLower(strings.TrimSpace(anyToString(props["component"], anyToString(props["native"], ""))))
	return component == "legend"
}

func firstLegendPathForFieldset(fieldsetPath string, state *GioWindowState) string {
	if state == nil || state.propsForPath == nil || fieldsetPath == "" {
		return ""
	}
	first := ""
	for candidatePath, candidateProps := range state.propsForPath {
		if parentPath(candidatePath) != fieldsetPath {
			continue
		}
		if !isLegendNodeProps(candidateProps) {
			continue
		}
		if first == "" || candidatePath < first {
			first = candidatePath
		}
	}
	return first
}

func isControlWithinDisabledFieldset(path string, state *GioWindowState) bool {
	if state == nil || path == "" {
		return false
	}
	cur := path
	for {
		cur = parentPath(cur)
		if cur == "" {
			break
		}
		ancestorProps, ok := state.propsForPath[cur]
		if !ok {
			continue
		}
		if isFieldsetNodeProps(ancestorProps) && propTruthy(ancestorProps["disabled"], "disabled") {
			legendPath := firstLegendPathForFieldset(cur, state)
			if legendPath != "" && pathIsDescendant(path, legendPath) {
				// HTML-like behavior: descendants of the first legend are not disabled by the fieldset.
				continue
			}
			return true
		}
	}
	return false
}

func isControlDisabledForSubmit(path string, props map[string]any, state *GioWindowState) bool {
	if isControlDisabledProps(props) {
		return true
	}
	return isControlWithinDisabledFieldset(path, state)
}

func pathIsDescendant(path, ancestor string) bool {
	if path == "" || ancestor == "" {
		return false
	}
	if path == ancestor {
		return true
	}
	return strings.HasPrefix(path, ancestor+"/")
}

func isFormNodeProps(props map[string]any) bool {
	tag := strings.ToLower(strings.TrimSpace(anyToString(props["tag"], "")))
	if tag == "form" {
		return true
	}
	component := strings.ToLower(strings.TrimSpace(anyToString(props["component"], anyToString(props["native"], ""))))
	return component == "form"
}

func isSubmitButtonProps(props map[string]any) bool {
	if props == nil {
		return false
	}
	tag := strings.ToLower(strings.TrimSpace(anyToString(props["tag"], "")))
	component := strings.ToLower(strings.TrimSpace(anyToString(props["component"], anyToString(props["native"], ""))))
	t := strings.ToLower(strings.TrimSpace(anyToString(props["type"], anyToString(props["buttonType"], ""))))
	if t == "submit" {
		return true
	}
	// HTML-like default: <button> without explicit type behaves as submit.
	if t == "" && (tag == "button" || component == "button") {
		return true
	}
	return false
}

func resolveSubmitterValue(path string, props map[string]any, state *GioWindowState) string {
	if props == nil {
		return ""
	}
	if state != nil {
		if v, ok := state.inputValues[path]; ok {
			return strings.TrimSpace(v)
		}
	}
	if raw := strings.TrimSpace(anyToString(props["value"], "")); raw != "" {
		return raw
	}
	return strings.TrimSpace(coreinput.InputExternalValue(props))
}

func appendSubmitterValue(values map[string]any, buttonPath string, buttonProps map[string]any, state *GioWindowState) {
	if !isSubmitButtonProps(buttonProps) {
		return
	}
	if isControlDisabledForSubmit(buttonPath, buttonProps, state) {
		return
	}
	name := strings.TrimSpace(anyToString(buttonProps["name"], ""))
	if name == "" {
		return
	}
	value := resolveSubmitterValue(buttonPath, buttonProps, state)
	appendFormFieldValue(values, name, value)
}

func resolveFormPathForSubmit(buttonPath string, buttonProps map[string]any, state *GioWindowState) string {
	if state == nil {
		return ""
	}
	formRef := strings.TrimSpace(anyToString(buttonProps["form"], anyToString(buttonProps["form-id"], "")))
	if formRef != "" {
		for path, props := range state.propsForPath {
			if !isFormNodeProps(props) {
				continue
			}
			if strings.TrimSpace(anyToString(props["id"], "")) == formRef {
				return path
			}
		}
	}

	cur := buttonPath
	for {
		cur = parentPath(cur)
		if cur == "" {
			break
		}
		if props, ok := state.propsForPath[cur]; ok && isFormNodeProps(props) {
			return cur
		}
	}
	return ""
}

func collectFormValues(formPath string, state *GioWindowState) map[string]any {
	values := map[string]any{}
	if state == nil || strings.TrimSpace(formPath) == "" {
		return values
	}
	if state.propsForPath == nil {
		return values
	}

	paths := make([]string, 0, len(state.propsForPath))
	seenPaths := make(map[string]struct{}, len(state.propsForPath))
	formID := resolveFormIDFromProps(state.propsForPath[formPath])

	for path := range state.propsForPath {
		if pathIsDescendant(path, formPath) {
			if _, seen := seenPaths[path]; !seen {
				paths = append(paths, path)
				seenPaths[path] = struct{}{}
			}
		}
	}
	if formID != "" {
		for path, props := range state.propsForPath {
			if pathIsDescendant(path, formPath) {
				continue
			}
			if !controlReferencesFormID(props, formID) {
				continue
			}
			if _, seen := seenPaths[path]; !seen {
				paths = append(paths, path)
				seenPaths[path] = struct{}{}
			}
		}
	}
	sort.Strings(paths)

	for _, path := range paths {
		props := state.propsForPath[path]
		if isControlDisabledForSubmit(path, props, state) {
			continue
		}
		if !isSubmittableControlProps(props) {
			continue
		}
		name := strings.TrimSpace(anyToString(props["name"], anyToString(props["id"], path)))
		if name == "" {
			continue
		}
		value, include := resolveSubmittableControlValue(path, props, state)
		if !include {
			continue
		}
		appendFormFieldValue(values, name, value)
	}

	return values
}

func appendFormFieldValue(values map[string]any, name string, value any) {
	if values == nil || strings.TrimSpace(name) == "" {
		return
	}
	if existing, ok := values[name]; ok {
		switch current := existing.(type) {
		case []any:
			values[name] = append(current, value)
		default:
			values[name] = []any{current, value}
		}
		return
	}
	values[name] = value
}

func isSubmittableControlProps(props map[string]any) bool {
	if props == nil || isControlDisabledProps(props) {
		return false
	}
	tag := strings.ToLower(strings.TrimSpace(anyToString(props["tag"], "")))
	component := strings.ToLower(strings.TrimSpace(anyToString(props["component"], anyToString(props["native"], ""))))
	inputType := strings.ToLower(strings.TrimSpace(anyToString(props["type"], anyToString(props["inputtype"], ""))))
	if inputType == "file" {
		return false
	}

	if tag == "input" {
		return inputType != "submit" && inputType != "reset" && inputType != "button" && inputType != "image"
	}
	if tag == "select" || tag == "textarea" {
		return true
	}
	if component == "select" || component == "dropdown" || component == "textarea" {
		return true
	}
	return inputType == "checkbox" || inputType == "check" || inputType == "radio"
}

func resolveSubmittableControlValue(path string, props map[string]any, state *GioWindowState) (any, bool) {
	tag := strings.ToLower(strings.TrimSpace(anyToString(props["tag"], "")))
	component := strings.ToLower(strings.TrimSpace(anyToString(props["component"], anyToString(props["native"], ""))))
	inputType := strings.ToLower(strings.TrimSpace(anyToString(props["type"], anyToString(props["inputtype"], ""))))

	if tag == "select" || component == "select" || component == "dropdown" {
		if state == nil {
			return strings.TrimSpace(anyToString(props["value"], "")), true
		}
		model := ResolveSelectModel(path, props, state.inputValues)
		if model.SelectedDisabled {
			return nil, false
		}
		return model.SelectedValue, true
	}

	if inputType == "checkbox" || inputType == "check" {
		if !resolveCheckedState(path, props, state) {
			return nil, false
		}
		value := strings.TrimSpace(anyToString(props["value"], "on"))
		if value == "" {
			value = "on"
		}
		return value, true
	}

	if inputType == "radio" {
		if !resolveCheckedState(path, props, state) {
			return nil, false
		}
		value := strings.TrimSpace(anyToString(props["value"], "on"))
		if value == "" {
			value = "on"
		}
		return value, true
	}

	if tag == "textarea" || component == "textarea" || tag == "input" {
		return resolveControlValue(path, props, state), true
	}

	return nil, false
}

func dispatchFormSubmitFromButton(state *GioWindowState, host *GioRenderHost, buttonPath string, buttonProps map[string]any) {
	if state == nil {
		return
	}
	if !isSubmitButtonProps(buttonProps) {
		return
	}
	formPath := resolveFormPathForSubmit(buttonPath, buttonProps, state)
	if formPath == "" {
		return
	}
	formProps := state.propsForPath[formPath]
	eventName := coreengine.ResolveOnSubmitEvent(formProps)
	formID := strings.TrimSpace(anyToString(formProps["id"], ""))
	values := collectFormValues(formPath, state)
	appendSubmitterValue(values, buttonPath, buttonProps, state)
	payload := coreengine.BuildFormSubmitEventPayload(buttonPath, formPath, formID, values, time.Now().UnixMilli())
	if err := dispatchHostEvent(host, eventName, payload); err != nil && host != nil && host.EmitRuntimeError != nil {
		host.EmitRuntimeError(err)
	}
}

func scrollTargetForPoint(state *GioWindowState, preferredPath string, p image.Point) string {
	return coreinput.SelectScrollTarget(state.scrollHits, preferredPath, p)
}

func clampScrollForPath(path string, state *GioWindowState) bool {
	hit, ok := state.scrollHits[path]
	if !ok {
		return false
	}
	cur, changed := coreinput.ClampScrollOffset(state.scrollOffsets[path], hit)
	if changed {
		state.scrollOffsets[path] = cur
		return true
	}
	return false
}

func applyKeyboardScrollState(path string, kev key.Event, state *GioWindowState) bool {
	hit, ok := state.scrollHits[path]
	if !ok {
		return false
	}
	next, changed := coreinput.ApplyKeyboardScroll(state.scrollOffsets[path], hit, kev.Name)
	state.scrollOffsets[path] = next
	return changed
}

func updateScrollByThumbState(path string, p image.Point, state *GioWindowState, axis string) bool {
	hit, ok := state.scrollHits[path]
	if !ok {
		return false
	}
	cur, changed := coreinput.UpdateScrollByThumb(state.scrollOffsets[path], hit, p, axis, state.scrollDragGrab)
	if changed {
		state.scrollOffsets[path] = cur
	}
	return changed
}

func handleScrollbarPressState(path string, p image.Point, state *GioWindowState) (handled bool, changed bool) {
	hit, ok := state.scrollHits[path]
	if !ok {
		return false, false
	}
	outcome, next := coreinput.HandleScrollbarPress(state.scrollOffsets[path], hit, p)
	if !outcome.Handled {
		return false, false
	}
	state.scrollDragPath = path
	state.scrollDragAxis = outcome.DragAxis
	state.scrollDragGrab = outcome.DragGrab
	state.scrollOffsets[path] = next
	return true, outcome.Changed
}

// processGioEvents reads all pending pointer events for registered tags.
// Returns true if a re-render is needed.
func ProcessGioEvents(rctx *coreengine.RenderContext, host *GioRenderHost, state *GioWindowState, frameTime time.Time) bool {
	gtx := rctx.Gtx
	needInvalidate := false
	for path, tag := range state.tags {
		for {
			ev, ok := gtx.Source.Event(pointer.Filter{
				Target:  tag,
				Kinds:   pointer.Enter | pointer.Leave | pointer.Move | pointer.Press | pointer.Release | pointer.Drag | pointer.Scroll,
				ScrollX: pointer.ScrollRange{Min: -1 << 20, Max: 1 << 20},
				ScrollY: pointer.ScrollRange{Min: -1 << 20, Max: 1 << 20},
			})
			if !ok {
				break
			}
			pe, ok := ev.(pointer.Event)
			if !ok {
				continue
			}
			dispatchPath := resolveCapturedPointerPath(state, path, pe.Kind)
			previousPointer := state.lastPointer[dispatchPath]
			pointerPoint := image.Pt(int(pe.Position.X), int(pe.Position.Y))
			state.pointerPos = image.Pt(int(pe.Position.X), int(pe.Position.Y))
			state.pointerKnown = true
			switch pe.Kind {
			case pointer.Enter:
				if host == nil || !coreinput.PathState(host.HoverState, path) {
					if host != nil {
						if host.HoverState == nil {
							host.HoverState = make(map[string]bool)
						}
						host.HoverState[path] = true
					}
					needInvalidate = true
					invalidateHost(host)
				}
			case pointer.Leave:
				if host != nil && coreinput.PathState(host.HoverState, path) {
					if host.HoverState != nil {
						delete(host.HoverState, path)
					}
					needInvalidate = true
					invalidateHost(host)
				}
			case pointer.Move:
				state.lastPointer[dispatchPath] = pointerPoint
				if dispatchPointerEventForProps(host, dispatchPath, state.propsForPath[dispatchPath], "pointermove", pointerPoint, image.Pt(pointerPoint.X-previousPointer.X, pointerPoint.Y-previousPointer.Y), pe.Buttons != 0, frameTime.UnixMilli()) {
					needInvalidate = true
					invalidateHost(host)
				}
				if state.selectMenuOpen != "" {
					needInvalidate = true
					invalidateHost(host)
				}
			case pointer.Press:
				state.lastPointer[path] = pointerPoint
				if handled, changed := handleScrollbarPressState(path, state.lastPointer[path], state); handled {
					if changed {
						needInvalidate = true
						invalidateHost(host)
					}
					continue
				}
				if isControlDisabledProps(state.propsForPath[path]) {
					continue
				}
				if shouldCaptureSelectOverlayEvent(state, path) {
					continue
				}
				if dispatchPointerEventForProps(host, path, state.propsForPath[path], "pointerdown", pointerPoint, image.Point{}, true, frameTime.UnixMilli()) {
					needInvalidate = true
					invalidateHost(host)
				}
				if coreengine.HasPointerEventHandlers(state.propsForPath[path]) {
					state.pointerCapturePath = path
				}
				// Track focus transitions for input elements.
				nextFocusedInput := ""
				if tag.path != "" {
					candidatePath := coreengine.NormalizeFocusableInputPath(tag.path)
					if inputProps, ok := state.propsForPath[candidatePath]; ok {
						if coreengine.IsInputLikeProps(inputProps) && !isControlDisabledProps(inputProps) {
							nextFocusedInput = candidatePath
						}
					}
				}
				if nextFocusedInput != state.focusedInputPath {
					previousFocusedInput := state.focusedInputPath
					for k := range state.inputFocused {
						state.inputFocused[k] = false
					}
					if nextFocusedInput != "" {
						state.inputFocused[nextFocusedInput] = true
					}
					state.focusedInputPath = nextFocusedInput

					if previousFocusedInput != "" {
						if prevProps, ok := state.propsForPath[previousFocusedInput]; ok {
							evBlur := coreengine.ResolveOnBlurEvent(prevProps)
							if evBlur != "" {
								blurPayload := coreengine.BuildInputValueEventPayload(previousFocusedInput, coreengine.ResolveInputComponentType(prevProps, "input"), state.inputValues[previousFocusedInput], false, frameTime.UnixMilli())
								_ = dispatchHostEvent(host, evBlur, blurPayload)
							}
						}
					}

					if nextFocusedInput != "" {
						if nextProps, ok := state.propsForPath[nextFocusedInput]; ok {
							evFocus := coreengine.ResolveOnFocusEvent(nextProps)
							if evFocus != "" {
								focusPayload := coreengine.BuildInputValueEventPayload(nextFocusedInput, coreengine.ResolveInputComponentType(nextProps, "input"), state.inputValues[nextFocusedInput], true, frameTime.UnixMilli())
								_ = dispatchHostEvent(host, evFocus, focusPayload)
							}
						}
					}
					needInvalidate = true
				}
				if nextFocusedInput != "" {
					if ed := state.editors[nextFocusedInput]; ed != nil {
						gtx.Execute(key.FocusCmd{Tag: ed})
						gtx.Execute(key.SoftKeyboardCmd{Show: true})
					}
				} else {
					gtx.Execute(key.FocusCmd{Tag: nil})
				}
				if h, ok := state.handlers[path]; ok {
					// Check active CSS selector
					props := state.propsForPath[path]
					if props != nil {
						appSS := state.frameStyleSheet
						hasActive := appSS != nil && uicore.HasActiveSelector(props, appSS)
						if hasActive {
							if host != nil {
								if host.ActiveState == nil {
									host.ActiveState = make(map[string]bool)
								}
								host.ActiveState[path] = true
							}
							capturedPath := path
							capturedHost := host
							time.AfterFunc(140*time.Millisecond, func() {
								if capturedHost != nil && capturedHost.ActiveState != nil {
									delete(capturedHost.ActiveState, capturedPath)
								}
								invalidateHost(capturedHost)
							})
							needInvalidate = true
						}
					}
					h()
					// Event handlers usually mutate external app state via DispatchEvent;
					// force a repaint so snapshot-based UIs reflect the new state immediately.
					needInvalidate = true
					invalidateHost(host)
				}
			case pointer.Drag:
				state.lastPointer[dispatchPath] = pointerPoint
				if isControlDisabledProps(state.propsForPath[dispatchPath]) {
					continue
				}
				if dispatchPointerEventForProps(host, dispatchPath, state.propsForPath[dispatchPath], "pointerdrag", pointerPoint, image.Pt(pointerPoint.X-previousPointer.X, pointerPoint.Y-previousPointer.Y), pe.Buttons != 0, frameTime.UnixMilli()) {
					needInvalidate = true
					invalidateHost(host)
				}
				if state.scrollDragPath == path && (state.scrollDragAxis == "x" || state.scrollDragAxis == "y") {
					if updateScrollByThumbState(path, state.lastPointer[path], state, state.scrollDragAxis) {
						needInvalidate = true
						invalidateHost(host)
					}
					continue
				}
				if h, ok := state.handlers[dispatchPath]; ok {
					props := state.propsForPath[dispatchPath]
					if props != nil {
						t := strings.ToLower(anyToString(props["__interactive"], ""))
						if t == "slider" || t == "range" {
							h()
						}
					}
				}
			case pointer.Release:
				state.lastPointer[dispatchPath] = pointerPoint
				if dispatchPointerEventForProps(host, dispatchPath, state.propsForPath[dispatchPath], "pointerup", pointerPoint, image.Pt(pointerPoint.X-previousPointer.X, pointerPoint.Y-previousPointer.Y), false, frameTime.UnixMilli()) {
					needInvalidate = true
					invalidateHost(host)
				}
				delete(state.lastPointer, path)
				if dispatchPath != path {
					delete(state.lastPointer, dispatchPath)
				}
				state.pointerCapturePath = ""
				if state.scrollDragPath == path {
					state.scrollDragPath = ""
					state.scrollDragAxis = ""
					state.scrollDragGrab = 0
				}
			case pointer.Scroll:
				if ed := state.editors[path]; ed != nil && ed.SingleLine {
					contentW := state.inputContentW[path]
					visibleW := state.inputVisibleW[path]
					maxScroll := max(0, contentW-visibleW)
					if maxScroll > 0 {
						axis := pe.Scroll.X
						if math.Abs(float64(pe.Scroll.Y)) > math.Abs(float64(axis)) {
							axis = pe.Scroll.Y
						}
						cur := state.inputScrollX[path]
						next := cur + int(-axis*12)
						if next < 0 {
							next = 0
							invalidateHost(host)
						}
						continue
					}
				}
				pointerPos := image.Pt(int(pe.Position.X), int(pe.Position.Y))
				targetPath := scrollTargetForPoint(state, path, pointerPos)
				if targetPath == "" {
					continue
				}
				state.scrollFocusPath = targetPath
				css := state.cssForPath[targetPath]
				overflow := strings.ToLower(strings.TrimSpace(css["overflow"]))
				overflowX := strings.ToLower(strings.TrimSpace(css["overflow-x"]))
				overflowY := strings.ToLower(strings.TrimSpace(css["overflow-y"]))
				if overflowX == "" {
					overflowX = overflow
				}
				if overflowY == "" {
					overflowY = overflow
				}
				if overflowX == "auto" || overflowX == "scroll" || overflowY == "auto" || overflowY == "scroll" {
					cur := state.scrollOffsets[targetPath]
					carry := state.scrollCarry[targetPath]
					next, carryX, carryY := coreinput.ApplyPointerScroll(cur, carry.X, carry.Y, pe.Scroll.X, pe.Scroll.Y)
					state.scrollCarry[targetPath] = f32.Pt(carryX, carryY)
					cur = next
					state.scrollOffsets[targetPath] = cur
					_ = clampScrollForPath(targetPath, state)
					needInvalidate = true
					invalidateHost(host)
				}
			}
		}

		if state.scrollFocusPath == path {
			for {
				ev, ok := gtx.Source.Event(key.Filter{Focus: tag, Name: key.NameDownArrow})
				if !ok {
					break
				}
				if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
					if applyKeyboardScrollState(path, kev, state) {
						needInvalidate = true
					}
				}
			}
			for {
				ev, ok := gtx.Source.Event(key.Filter{Focus: tag, Name: key.NameUpArrow})
				if !ok {
					break
				}
				if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
					if applyKeyboardScrollState(path, kev, state) {
						needInvalidate = true
					}
				}
			}
			for {
				ev, ok := gtx.Source.Event(key.Filter{Focus: tag, Name: key.NameRightArrow})
				if !ok {
					break
				}
				if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
					if applyKeyboardScrollState(path, kev, state) {
						needInvalidate = true
					}
				}
			}
			for {
				ev, ok := gtx.Source.Event(key.Filter{Focus: tag, Name: key.NameLeftArrow})
				if !ok {
					break
				}
				if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
					if applyKeyboardScrollState(path, kev, state) {
						needInvalidate = true
					}
				}
			}
			for {
				ev, ok := gtx.Source.Event(key.Filter{Focus: tag, Name: key.NamePageDown})
				if !ok {
					break
				}
				if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
					if applyKeyboardScrollState(path, kev, state) {
						needInvalidate = true
					}
				}
			}
			for {
				ev, ok := gtx.Source.Event(key.Filter{Focus: tag, Name: key.NamePageUp})
				if !ok {
					break
				}
				if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
					if applyKeyboardScrollState(path, kev, state) {
						needInvalidate = true
					}
				}
			}
			for {
				ev, ok := gtx.Source.Event(key.Filter{Focus: tag, Name: key.NameHome})
				if !ok {
					break
				}
				if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
					if applyKeyboardScrollState(path, kev, state) {
						needInvalidate = true
					}
				}
			}
			for {
				ev, ok := gtx.Source.Event(key.Filter{Focus: tag, Name: key.NameEnd})
				if !ok {
					break
				}
				if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
					if applyKeyboardScrollState(path, kev, state) {
						needInvalidate = true
					}
				}
			}
		}
	}
	return needInvalidate
}

func roundedFromProps(props map[string]any, css map[string]string, w int, h int) int {
	return ResolveRoundedFromProps(props, css, w, h)
}

func childrenContentBounds(children []any, fallbackRight int, fallbackBottom int) (int, int) {
	childMaps := make([]map[string]any, 0, len(children))
	for _, child := range children {
		childMaps = append(childMaps, anyToMap(child))
	}
	return ChildrenContentBounds(childMaps, fallbackRight, fallbackBottom)
}

func drawScrollbars(gtx layout.Context, x int, y int, w int, h int, contentRight int, contentBottom int, css map[string]string, radius int, scrollX int, scrollY int) coreinput.ScrollHitInfo {
	data := coreinput.ComputeScrollbarRenderData(x, y, w, h, contentRight, contentBottom, css, scrollX, scrollY)
	info := data.Hit
	if info.HasV && !info.TrackV.Empty() {
		drawGioRRect(gtx.Ops, info.TrackV.Min.X, info.TrackV.Min.Y, info.TrackV.Dx(), info.TrackV.Dy(), data.Radius, data.TrackCol)
		if !info.ThumbV.Empty() {
			drawGioRRect(gtx.Ops, info.ThumbV.Min.X, info.ThumbV.Min.Y, info.ThumbV.Dx(), info.ThumbV.Dy(), data.Radius, data.ThumbCol)
		}
	}
	if info.HasH && !info.TrackH.Empty() {
		drawGioRRect(gtx.Ops, info.TrackH.Min.X, info.TrackH.Min.Y, info.TrackH.Dx(), info.TrackH.Dy(), data.Radius, data.TrackCol)
		if !info.ThumbH.Empty() {
			drawGioRRect(gtx.Ops, info.ThumbH.Min.X, info.ThumbH.Min.Y, info.ThumbH.Dx(), info.ThumbH.Dy(), data.Radius, data.ThumbCol)
		}
	}
	_ = radius
	return info
}

// drawGioText renders text at absolute position (x,y) with size (w,h).
func drawGioText(gtx layout.Context, state *GioWindowState, x, y, w, h int,
	txt string, col color.NRGBA, fontSize float32, bold, italic, mono bool,
	align text.Alignment, maxLines int, wrapText bool) {
	if txt == "" || w <= 0 || h <= 0 {
		return
	}
	// Record the color for text material
	macro := op.Record(gtx.Ops)
	paint.ColorOp{Color: col}.Add(gtx.Ops)
	colorCall := macro.Stop()

	w2 := font.Normal
	if bold {
		w2 = font.Bold
	}
	style := font.Regular
	if italic {
		style = font.Italic
	}
	// Use word wrapping as the default CSS-like behavior; heuristic wrapping can
	// split short tokens (e.g. "50%") in tight boxes and cause visual clipping.
	wrapPolicy := text.WrapWords
	label := widget.Label{
		Alignment:  align,
		MaxLines:   maxLines,
		WrapPolicy: wrapPolicy,
		Truncator:  "",
	}

	// Browser-like centering for one-line text: measure shaped bounds first,
	// then place exact text box in the center of the available area.
	runeCount := len([]rune(strings.TrimSpace(txt)))
	shortToken := !strings.Contains(txt, " ") && runeCount > 0 && runeCount <= 8
	fitsSingleLine := false
	if !strings.Contains(txt, "\n") {
		estW, _ := uicore.EstimateTextLayoutBox(txt, fontSize, map[string]string{})
		fitsSingleLine = estW <= max(1, w)
	}
	singleLine := maxLines == 1 || !wrapText || fitsSingleLine
	lineLike := singleLine && !strings.Contains(txt, "\n") && (maxLines == 1 || h <= max(18, int(float64(fontSize)*1.9)) || shortToken)
	if lineLike {
		// Give single-line labels a slightly taller box so glyph descenders
		// like g, j, p, q, y are not clipped by tight exact constraints.
		th := max(1, min(h, int(math.Ceil(float64(fontSize)*1.65))+3))
		ty := y + max(0, (h-th)/2)

		tr := op.Offset(image.Pt(x, ty)).Push(gtx.Ops)
		gtx2 := gtx
		gtx2.Constraints = layout.Exact(image.Pt(w, th))
		widget.Label{
			Alignment:  align,
			MaxLines:   1,
			WrapPolicy: wrapPolicy,
			Truncator:  "",
		}.Layout(gtx2, state.shaper, font.Font{Weight: w2, Style: style}, unit.Sp(fontSize), txt, colorCall)
		tr.Pop()
		return
	}

	// Translate to position
	tr := op.Offset(image.Pt(x, y)).Push(gtx.Ops)

	// Set exact constraints
	gtx2 := gtx
	gtx2.Constraints = layout.Exact(image.Pt(w, h))

	label.Layout(gtx2, state.shaper, font.Font{Weight: w2, Style: style}, unit.Sp(fontSize), txt, colorCall)

	tr.Pop()
}

func DrawGioText(rctx *coreengine.RenderContext, state *GioWindowState, x, y, w, h int,
	txt string, col color.NRGBA, fontSize float32, bold, italic, mono bool,
	align text.Alignment, maxLines int, wrapText bool) {
	drawGioText(rctx.Gtx, state, x, y, w, h, txt, col, fontSize, bold, italic, mono, align, maxLines, wrapText)
}

func drawDecorationLine(ops *op.Ops, x, y, w, thickness int, col color.NRGBA, style string) {
	if col.A == 0 {
		return
	}
	for _, r := range BuildDecorationLineRects(x, y, w, thickness, style) {
		if r.Rounded {
			drawGioRRect(ops, r.X, r.Y, r.W, r.H, r.Radius, col)
			continue
		}
		drawGioRect(ops, r.X, r.Y, r.W, r.H, col)
	}
}

func drawGioTextDecorations(gtx layout.Context, x, y, w, h int, fontSize float32, align text.Alignment, txt string, base color.NRGBA, deco uicore.TextDecorationInfo) {
	alignKey := "start"
	switch align {
	case text.Middle:
		alignKey = "middle"
	case text.End:
		alignKey = "end"
	}
	plan := BuildTextDecorationPlan(x, y, w, h, fontSize, alignKey, txt, base, deco)
	for _, r := range plan.Rects {
		if r.Rounded {
			drawGioRRect(gtx.Ops, r.X, r.Y, r.W, r.H, r.Radius, plan.Color)
			continue
		}
		drawGioRect(gtx.Ops, r.X, r.Y, r.W, r.H, plan.Color)
	}
}

// drawGioImage renders a raster image at absolute position with scaling.
func drawGioImage(ops *op.Ops, x, y, w, h int, img image.Image) {
	if img == nil || w <= 0 || h <= 0 {
		return
	}
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()
	plan, ok := BuildImagePaintPlan(x, y, w, h, imgW, imgH)
	if !ok {
		return
	}

	// Clip to destination
	cs := clip.Rect(plan.Clip).Push(ops)

	tr := op.Offset(plan.Offset).Push(ops)

	aff := f32.Affine2D{}.Scale(f32.Pt(0, 0), f32.Pt(plan.ScaleX, plan.ScaleY))
	op.Affine(aff).Add(ops)

	imgOp := paint.NewImageOp(img)
	imgOp.Add(ops)
	paint.PaintOp{}.Add(ops)

	tr.Pop()
	cs.Pop()
}

// registerGioEventArea registers a clip area with stable tag for pointer events.
func registerGioEventArea(ops *op.Ops, x, y, w, h int, tag *PointerTag) {
	area := clip.Rect(image.Rect(x, y, x+w, y+h)).Push(ops)
	event.Op(ops, tag)
	area.Pop()
}

// applyFrameCursorOverride keeps cursor style stable while pointer is stationary.
func applyFrameCursorOverride(ops *op.Ops, state *GioWindowState) {
	if state == nil {
		return
	}
	if state.frameCursorValue != "" {
		coreinput.CSSCursorToGio(state.frameCursorValue).Add(ops)
		return
	}
	if len(state.frameHoverState) == 0 || len(state.cssForPath) == 0 {
		return
	}
	bestPathLen := -1
	bestCursor := ""
	for path, hovered := range state.frameHoverState {
		if !hovered {
			continue
		}
		css, ok := state.cssForPath[path]
		if !ok || css == nil {
			continue
		}
		cursorRaw := uicore.ParseCursor(css)
		if cursorRaw == "" || cursorRaw == "default" || cursorRaw == "auto" {
			continue
		}
		if len(path) > bestPathLen {
			bestPathLen = len(path)
			bestCursor = cursorRaw
		}
	}
	if bestCursor != "" {
		coreinput.CSSCursorToGio(bestCursor).Add(ops)
	}
}

func ApplyFrameCursorOverride(rctx *coreengine.RenderContext, state *GioWindowState) {
	applyFrameCursorOverride(rctx.Gtx.Ops, state)
}

func drawGioDebugOverlay(gtx layout.Context, state *GioWindowState, host *GioRenderHost, path string, kind string, x, y, w, h int, css map[string]string, clipChildren bool, contentRight int, contentBottom int) {
	if state == nil || !state.frameDebug {
		return
	}
	nativeKind := strings.ToLower(anyToString(state.propsForPath[path]["component"], anyToString(state.propsForPath[path]["native"], "")))
	_, hasHandler := state.handlers[path]
	overlay := coredebug.BuildDebugOverlayModel(coredebug.DebugOverlayInput{
		Path:          path,
		Kind:          kind,
		NativeKind:    nativeKind,
		Active:        host != nil && coreinput.PathState(host.ActiveState, path),
		Hovered:       host != nil && coreinput.PathState(host.HoverState, path),
		HasHandler:    hasHandler,
		ClipChildren:  clipChildren,
		X:             x,
		Y:             y,
		W:             w,
		H:             h,
		ContentRight:  contentRight,
		ContentBottom: contentBottom,
	})

	drawDebugOutline(gtx.Ops, x, y, w, h, 1, overlay.OutlineColor)
	if overlay.Interactive {
		drawDebugOutline(gtx.Ops, x+2, y+2, max(1, w-4), max(1, h-4), 1, overlay.InteractiveCol)
	}
	if overlay.ClipChildren {
		drawDebugOutline(gtx.Ops, x+1, y+1, max(1, w-2), max(1, h-2), 1, overlay.ScrollCol)
		if overlay.HasOverflow {
			cw := max(1, contentRight-x)
			ch := max(1, contentBottom-y)
			drawDebugOutline(gtx.Ops, x, y, cw, ch, 1, color.NRGBA{R: 0, G: 255, B: 255, A: 90})
		}
	}
	if overlay.Label != "" {
		drawGioText(gtx, state, x+3, y+2, max(1, w-6), min(14, h), overlay.Label, overlay.OutlineColor, 9, false, false, true, text.Start, 1, false)
	}
	_ = css
}

// drawGioTree traverses a computed layout tree (from layoutNodeToNative) and draws each element.
func DrawGioTree(rctx *coreengine.RenderContext, node map[string]any, host *GioRenderHost, path string, state *GioWindowState, parentClass string, containingBlock image.Rectangle, parentCSS map[string]string, inheritedShiftX int, inheritedShiftY int, frameTime time.Time) {
	gtx := rctx.Gtx
	if node == nil {
		return
	}
	_ = parentClass
	initRenderConfig()

	kind := coreengine.LowerASCIIIfNeeded(anyToString(node["kind"], "view"))
	props := anyToMap(node["props"])
	box := anyToMap(node["layout"])
	x := anyToInt(box["x"], 0)
	y := anyToInt(box["y"], 0)
	w := max(1, anyToInt(box["width"], 100))
	h := max(1, anyToInt(box["height"], 30))
	hovered := state.frameHoverState != nil && state.frameHoverState[path]
	active := state.frameActiveState != nil && state.frameActiveState[path]
	focused := state.inputFocused[path] || state.focusedInputPath == path
	disabled := isControlDisabledProps(props)
	checked := resolveCheckedState(path, props, state)
	invalid := resolveInvalidState(path, props, state, checked, disabled)
	viewW := state.frameViewW
	viewH := state.frameViewH
	if viewW <= 0 || viewH <= 0 {
		return
	}
	childrenSlice := anyToSlice(node["children"])
	if kind == "select" {
		kind = "native"
		if strings.TrimSpace(anyToString(props["component"], "")) == "" {
			props["component"] = "select"
		}
		if !hasSelectOptions(props["options"]) {
			if childOptions := extractSelectOptionsFromChildren(childrenSlice); len(childOptions) > 0 {
				props["options"] = childOptions
			}
		}
		// Native select consumes option children as metadata; they should not render as standalone nodes.
		childrenSlice = nil
	}

	if corelayout.ShouldQuickCullNode(path, x, y, w, h, inheritedShiftX, inheritedShiftY, state.frameViewport96, viewW, viewH, len(childrenSlice), kind) {
		return
	}

	profileThisNode := false
	if state != nil {
		profileThisNode = corelayout.ShouldProfileNode(state.profileComponents, state.profileFull, path, state.profileSampleFrame)
	}
	var childrenDur time.Duration
	var nodeStart time.Time
	if profileThisNode {
		nodeStart = frameTime
		defer func() {
			totalDur := time.Since(nodeStart)
			selfDur := totalDur - childrenDur
			if selfDur < 0 {
				selfDur = 0
			}
			if host != nil && host.RecordDebugComponent != nil {
				host.RecordDebugComponent(path, kind, props, w, h, totalDur, selfDur)
			}
		}()
	}

	// Resolve CSS using per-frame uicore.StyleSheet snapshot to avoid per-node locking.
	appSS := state.frameStyleSheet
	hasAdvancedSelectors := appSS != nil && appSS.HasAdvancedSelectors()
	parentSig := coreengine.InheritedTextCSSSignature(parentCSS)
	propSig := stylePropSignature(kind, props)
	var css map[string]string
	cssWritable := false
	ensureCSSWritable := func() {
		if cssWritable {
			return
		}
		css = coreengine.CloneStringMap(css)
		cssWritable = true
	}
	if !hasAdvancedSelectors {
		if cached, ok := state.resolvedCSS[path]; ok && cached.Hovered == hovered && cached.Active == active && cached.Focused == focused && cached.Disabled == disabled && cached.Checked == checked && cached.Invalid == invalid && cached.ViewW == viewW && cached.LowPower == renderLowPower && cached.ParentSig == parentSig && cached.PropSig == propSig {
			cached.LastSeen = state.frameNumber
			state.resolvedCSS[path] = cached
			css = cached.CSS
		}
	}
	if css == nil {
		css = uicore.ResolveStyle(props, appSS, viewW)
		css = corelayout.MergeInheritedTextCSS(css, parentCSS)
		if hovered && appSS != nil {
			uicore.ApplyHoverStyles(css, props, appSS, hovered)
		}
		if active && appSS != nil {
			uicore.ApplyActiveStyles(css, props, appSS, active)
		}
		if focused && appSS != nil {
			uicore.ApplyFocusStyles(css, props, appSS, focused)
		}
		if disabled && appSS != nil {
			uicore.ApplyDisabledStyles(css, props, appSS, disabled)
		}
		if checked && appSS != nil {
			uicore.ApplyCheckedStyles(css, props, appSS, checked)
		}
		if invalid && appSS != nil {
			uicore.ApplyInvalidStyles(css, props, appSS, invalid)
		}
		if hasAdvancedSelectors {
			uicore.ApplyAdvancedSelectorStyles(css, appSS, uicore.AdvancedSelectorContext{
				Path:  path,
				Props: props,
				LookupProps: func(target string) (map[string]any, bool) {
					matched, ok := state.propsForPath[target]
					return matched, ok
				},
				ParentPath:           parentPath,
				PreviousSiblingPath:  func(target string) string { return previousSiblingPath(target, state) },
				PreviousSiblingPaths: func(target string) []string { return previousSiblingPaths(target, state) },
				PseudoState: func(target string, targetProps map[string]any, pseudo string) bool {
					return selectorPseudoState(state, target, targetProps, pseudo)
				},
			})
			applyInlineStyleOverrides(css, props)
		}
		cssWritable = true
		if !hasAdvancedSelectors {
			state.resolvedCSS[path] = uicore.ResolvedCSSCacheEntry{
				Hovered:   hovered,
				Active:    active,
				Focused:   focused,
				Disabled:  disabled,
				Checked:   checked,
				Invalid:   invalid,
				ViewW:     viewW,
				LowPower:  renderLowPower,
				ParentSig: parentSig,
				PropSig:   propSig,
				LastSeen:  state.frameNumber,
				CSS:       coreengine.CloneStringMap(css),
			}
		}
	}
	if renderLowPower {
		ensureCSSWritable()
		css["filter"] = ""
		css["transition"] = "none"
	} else {
		if host != nil {
			rawTransition := strings.TrimSpace(css["transition"])
			if rawTransition != "" && !strings.EqualFold(rawTransition, "none") {
				if host.RenderStore == nil {
					host.RenderStore = uicore.NewRenderStore()
				}
				prev, ok := host.RenderStore.Previous(path)
				host.RenderStore.Capture(path, css, x, y, w, h)
				if ok && prev.Changed(css, x, y, w, h) {
					_, props := uicore.ParseTransition(css)
					if len(props) > 0 {
						invalidateHost(host)
					}
				}
			}
		}
	}

	// CSS transform values are applied as rendering transforms later,
	// so descendants (like text children) inherit the same motion/scale.
	scale := uicore.CSSScale(css)
	rotateDeg := uicore.ParseTransformRotateDegrees(css)
	filterRaw := strings.TrimSpace(css["filter"])
	if !renderLowPower && corelayout.IsFilterEnabled(filterRaw) && host != nil && host.CanUseWindowEffects {
		for k, v := range corelayout.BuildFilterFactorVars(filterRaw) {
			ensureCSSWritable()
			css[k] = v
		}
	}

	// Apply position: absolute/relative/fixed offsets (convert to pixel adjustments)
	viewport := state.frameViewportRect
	if viewport.Empty() {
		viewport = image.Rect(0, 0, viewW, viewH)
	}
	layoutX := x
	layoutY := y
	positioned := corelayout.ApplyPositionLayout(x, y, w, h, uicore.ParsePositionAndOffset(css, w, h), containingBlock, viewport)
	x = positioned.X
	y = positioned.Y
	positionDeltaX := x - layoutX
	positionDeltaY := y - layoutY

	// Parse CSS transform translate; applied as render transform later.
	translateX, translateY := uicore.ParseTransformTranslate(css)

	// Apply width/height CSS properties
	cssW, cssH, hasWidth, hasHeight := uicore.ParseWidthHeightCSS(css, w, h)
	if hasWidth {
		w = cssW
	}
	if hasHeight {
		h = cssH
	}

	if corelayout.IsDisplayNone(css["display"]) {
		return
	}

	// Keep a per-frame registry of rendered nodes so cross-node interactions
	// (e.g., submit buttons resolving linked forms) can work reliably.
	state.propsForPath[path] = props
	state.cssForPath[path] = css

	nodeRect := positioned.NodeRect
	childContainingBlock := positioned.ChildContainingBox

	visible := !corelayout.IsVisibilityHidden(css["visibility"])
	overflowLayout := corelayout.ResolveOverflowLayout(css)
	overflowX := overflowLayout.OverflowX
	overflowY := overflowLayout.OverflowY
	clipChildren := overflowLayout.ClipChildren
	visibility := corelayout.ComputeRenderVisibility(path, x, y, w, h, inheritedShiftX, inheritedShiftY, translateX, translateY, scale, state.frameViewport48, viewW, viewH)
	renderRect := visibility.RenderRect
	isOnScreen := visibility.IsOnScreen
	if !isOnScreen && len(childrenSlice) == 0 && kind != "input" && kind != "button" {
		return
	}

	var nodeTranslateTransform op.TransformStack
	hasNodeTranslateTransform := false
	var nodeScaleTransform op.TransformStack
	hasNodeScaleTransform := false
	var nodeRotateTransform op.TransformStack
	hasNodeRotateTransform := false
	if translateX != 0 || translateY != 0 {
		nodeTranslateTransform = op.Offset(image.Pt(int(math.Round(translateX)), int(math.Round(translateY)))).Push(gtx.Ops)
		hasNodeTranslateTransform = true
	}
	if math.Abs(rotateDeg) > 0.001 {
		cx := float32(x) + float32(w)/2
		cy := float32(y) + float32(h)/2
		radians := float32(rotateDeg * math.Pi / 180.0)
		aff := f32.Affine2D{}.Rotate(f32.Pt(cx, cy), radians)
		nodeRotateTransform = op.Affine(aff).Push(gtx.Ops)
		hasNodeRotateTransform = true
	}
	if scale > 0 && math.Abs(scale-1.0) > 0.001 {
		cx := float32(x) + float32(w)/2
		cy := float32(y) + float32(h)/2
		aff := f32.Affine2D{}.Scale(f32.Pt(cx, cy), f32.Pt(float32(scale), float32(scale)))
		nodeScaleTransform = op.Affine(aff).Push(gtx.Ops)
		hasNodeScaleTransform = true
	}

	if visible && isOnScreen {
		cursorRaw := uicore.ParseCursor(css)
		if corelayout.ShouldApplyNodeCursor(cursorRaw) {
			if corelayout.ShouldPromoteFrameCursor(path, state.frameCursorPath, state.pointerKnown, state.pointerPos, renderRect) {
				state.frameCursorPath = path
				state.frameCursorValue = cursorRaw
			}
			cursorClip := clip.Rect(nodeRect).Push(gtx.Ops)
			coreinput.CSSCursorToGio(cursorRaw).Add(gtx.Ops)
			cursorClip.Pop()
		}

		radius := roundedFromProps(props, css, w, h)
		radii := cssBorderRadiiValues(css, w, h)

		// Draw box-shadow first so it appears beneath the element.
		drawGioBoxShadow(gtx.Ops, x, y, w, h, radius, css)

		// Draw background
		drawGioBackground(gtx.Ops, x, y, w, h, css)

		// Draw inset shadow over background, clipped to element interior.
		drawGioInsetBoxShadow(gtx.Ops, x, y, w, h, radius, css)

		// Draw border
		drawGioElementBorder(gtx.Ops, x, y, w, h, radii, css)

		// Draw node content
		switch kind {
		case "text", "label", "span", "p", "h1", "h2", "h3", "h4", "h5", "h6":
			fgHex := uicore.CSSGetColor(css, "color", anyToString(props["fg"], "#e8e8e8"))
			fg := toNRGBA(uicore.ApplyCSSOpacity(uicore.ParseHexColor(fgHex, color.NRGBA{R: 0xE8, G: 0xE8, B: 0xE8, A: 0xFF}), css))
			plan := BuildTextContentPlan(props, css, w, DefaultTextFallbackSize(kind))
			drawGioText(gtx, state, x, y, w, h, plan.Text, fg, plan.FontSize,
				plan.Bold, plan.Italic, plan.UseMono,
				pkgcore.CSSTextAlign(css), plan.MaxLines, plan.WrapAllowed)
			drawGioTextDecorations(gtx, x, y, w, h, plan.FontSize, pkgcore.CSSTextAlign(css), plan.Text, fg, uicore.ParseTextDecoration(css))

		case "button":
			plan := BuildButtonContentPlan(props, css, x, y, w, h, 13)
			fgHex := uicore.CSSGetColor(css, "color", anyToString(props["fg"], "#e8e8e8"))
			fg := toNRGBA(uicore.ApplyCSSOpacity(uicore.ParseHexColor(fgHex, color.NRGBA{R: 0xE8, G: 0xE8, B: 0xE8, A: 0xFF}), css))
			drawGioText(gtx, state, plan.TextX, plan.TextY, plan.TextW, plan.TextH, plan.Label, fg, plan.FontSize,
				plan.Bold, plan.Italic, false, text.Middle, 1, false)
			drawGioTextDecorations(gtx, plan.TextX, plan.TextY, plan.TextW, plan.TextH, plan.FontSize, text.Middle, plan.Label, fg, uicore.ParseTextDecoration(css))

			// Register click event
			eventName := coreengine.ResolveComponentEventName(props, "click")
			capturedHost := host
			capturedPath := path
			capturedLabel := plan.Label
			state.handlers[path] = func() {
				payloadMap := coreengine.BuildButtonEventPayload(capturedPath, "button", capturedLabel, time.Now().UnixMilli())
				if err := dispatchHostEvent(capturedHost, eventName, payloadMap); err != nil && capturedHost != nil && capturedHost.EmitRuntimeError != nil {
					capturedHost.EmitRuntimeError(err)
				}

				buttonProps := state.propsForPath[capturedPath]
				if isSubmitButtonProps(buttonProps) {
					dispatchFormSubmitFromButton(state, capturedHost, capturedPath, buttonProps)
				}
			}
			state.propsForPath[path] = props
			state.cssForPath[path] = css
			registerGioEventArea(gtx.Ops, x, y, w, h, state.GetTag(path))

		case "input":
			if _, hasTag := props["tag"]; !hasTag {
				props["tag"] = "input"
			}
			inputType := strings.ToLower(anyToString(props["type"], anyToString(props["inputtype"], "text")))
			padX, padY := 6, 4
			fontSize := uicore.CSSFontSize(css, 13)
			fgHex := uicore.CSSGetColor(css, "color", "#e2e8f0")
			textColor := toNRGBA(uicore.ParseHexColor(fgHex, color.NRGBA{R: 0xE2, G: 0xE8, B: 0xF0, A: 0xFF}))
			phColor := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "placeholder-color", "#64748b"), color.NRGBA{R: 0x64, G: 0x74, B: 0x8B, A: 0xCC}))
			accentCol := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "accent-color", "#4ade80"), color.NRGBA{R: 0x4A, G: 0xDE, B: 0x80, A: 0xFF}))
			borderCol := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "border-color", "#475569"), color.NRGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xFF}))

			switch inputType {
			case "hidden":
				// No visual output â€” skip rendering.

			case "checkbox":
				checked := ResolveCheckboxChecked(path, props, state.boolValues, state.inputValues)
				binaryLayout := BuildBinaryControlLayout(x, y, w, h, 4)
				bg := color.NRGBA{R: 30, G: 41, B: 59, A: 255}
				if checked {
					bg = accentCol
				}
				drawGioRRect(gtx.Ops, binaryLayout.BoxX, binaryLayout.BoxY, binaryLayout.BoxSize, binaryLayout.BoxSize, 3, bg)
				drawGioBorder(gtx.Ops, binaryLayout.BoxX, binaryLayout.BoxY, binaryLayout.BoxSize, binaryLayout.BoxSize, 3, 2, borderCol)
				if checked {
					drawGioText(gtx, state, binaryLayout.BoxX, binaryLayout.BoxY, binaryLayout.BoxSize, binaryLayout.BoxSize, "âœ“",
						color.NRGBA{R: 15, G: 23, B: 25, A: 255}, fontSize, true, false, false, text.Middle, 1, false)
				}
				eventName := coreengine.ResolveComponentEventName(props, "change")
				registerCheckboxToggleHandler(state, host, path, eventName)
				state.propsForPath[path] = props
				state.cssForPath[path] = css
				registerGioEventArea(gtx.Ops, x, y, w, h, state.GetTag(path))

			case "radio":
				radioModel := ResolveRadioGroupModel(path, props, state.inputValues)
				binaryLayout := BuildBinaryControlLayout(x, y, w, h, 4)
				bgRad := color.NRGBA{R: 30, G: 41, B: 59, A: 255}
				drawGioRRect(gtx.Ops, binaryLayout.BoxX, binaryLayout.BoxY, binaryLayout.BoxSize, binaryLayout.BoxSize, binaryLayout.BoxSize/2, bgRad)
				drawGioBorder(gtx.Ops, binaryLayout.BoxX, binaryLayout.BoxY, binaryLayout.BoxSize, binaryLayout.BoxSize, binaryLayout.BoxSize/2, 2, borderCol)
				if radioModel.Selected {
					drawGioRRect(gtx.Ops, binaryLayout.BoxX+(binaryLayout.BoxSize-binaryLayout.Inner)/2, binaryLayout.BoxY+(binaryLayout.BoxSize-binaryLayout.Inner)/2, binaryLayout.Inner, binaryLayout.Inner, binaryLayout.Inner/2, accentCol)
				}
				eventName2 := coreengine.ResolveComponentEventName(props, "change")
				registerRadioSelectHandler(state, host, path, radioModel.GroupKey, radioModel.Value, eventName2)
				state.propsForPath[path] = props
				state.cssForPath[path] = css
				registerGioEventArea(gtx.Ops, x, y, w, h, state.GetTag(path))

			case "color":
				colValue := anyToString(props["value"], anyToString(props["text"], "#4ade80"))
				swatchCol := toNRGBA(uicore.ParseHexColor(colValue, color.NRGBA{R: 0x4A, G: 0xDE, B: 0x80, A: 0xFF}))
				swatch := BuildColorSwatchLayout(x, y, w, h, 4, 4, 4)
				drawGioRRect(gtx.Ops, swatch.X, swatch.Y, swatch.W, swatch.H, swatch.Radius, swatchCol)
				drawGioBorder(gtx.Ops, swatch.X, swatch.Y, swatch.W, swatch.H, swatch.Radius, 2, borderCol)

			case "range":
				sliderState := ResolveSliderState(path, props, state.sliderValues, 0, 100)
				plan := BuildSliderVisualPlan(sliderState.MinV, sliderState.MaxV, sliderState.Value, w, h, 5)
				trackCol := color.NRGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xFF}
				drawGioRRect(gtx.Ops, x, y+h/2-3, w, 6, 3, trackCol)
				if plan.FillW > 0 {
					drawGioRRect(gtx.Ops, x, y+h/2-3, plan.FillW, 6, 3, accentCol)
				}
				drawGioRRect(gtx.Ops, x+plan.ThumbX-plan.ThumbR, y+h/2-plan.ThumbR, plan.ThumbR*2, plan.ThumbR*2, plan.ThumbR, accentCol)
				state.boundsForPath[path] = image.Rect(x, y, x+w, y+h)
				eventName3 := coreengine.ResolveComponentEventName(props, "change")
				registerSliderPointerHandler(state, host, path, sliderState.MinV, sliderState.MaxV, eventName3, "range")
				props["__interactive"] = "slider"
				state.propsForPath[path] = props
				state.cssForPath[path] = css
				registerGioEventArea(gtx.Ops, x, y, w, h, state.GetTag(path))

			case "submit", "reset", "button":
				label := uicore.SanitizeRenderText(anyToString(props["text"], anyToString(props["value"], inputType)))
				fg := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "color", "#e8e8e8"), color.NRGBA{R: 0xE8, G: 0xE8, B: 0xE8, A: 0xFF}))
				drawGioText(gtx, state, x, y, w, h, label, fg, fontSize, uicore.CSSBold(css), uicore.CSSItalic(css), false, text.Middle, 1, false)
				capturedHost4 := host
				eventName4 := coreengine.ResolveComponentEventName(props, "click")
				capturedPath4 := path
				capturedInputType4 := inputType
				capturedLabel4 := label
				state.handlers[path] = func() {
					payloadMap4 := coreengine.BuildButtonEventPayload(capturedPath4, capturedInputType4, capturedLabel4, time.Now().UnixMilli())
					_ = dispatchHostEvent(capturedHost4, eventName4, payloadMap4)
				}
				state.propsForPath[path] = props
				state.cssForPath[path] = css
				registerGioEventArea(gtx.Ops, x, y, w, h, state.GetTag(path))

			default:
				// text, password, email, number, tel, url, search,
				// date, datetime-local, time, month, week, textarea
				editorCfg := BuildInputEditorConfig(inputType)
				isTextarea := !editorCfg.SingleLine
				externalRaw, hasExternalValue := coreinput.InputExternalValueWithPresence(props)
				externalVal := coreinput.NormalizeTypedInputValue(inputType, externalRaw, props, true)
				ed := state.editors[path]
				if ed == nil {
					ed = new(widget.Editor)
					ed.SingleLine = editorCfg.SingleLine
					if editorCfg.UseHeuristicWrap {
						ed.WrapPolicy = text.WrapHeuristically
					} else {
						ed.WrapPolicy = text.WrapWords
					}
					ed.MaxLen = editorCfg.MaxLen
					ed.Mask = editorCfg.Mask
					if externalVal != "" {
						ed.SetText(externalVal)
						state.inputValues[path] = externalVal
					}
					state.inputExternal[path] = externalVal
					state.editors[path] = ed
				}
				// Defensive: ensure SingleLine matches the input type (guards against stale state).
				if ed.SingleLine != editorCfg.SingleLine {
					ed.SingleLine = editorCfg.SingleLine
				}
				ed.MaxLen = editorCfg.MaxLen
				ed.Mask = editorCfg.Mask
				if editorCfg.UseHeuristicWrap {
					ed.WrapPolicy = text.WrapHeuristically
				} else {
					ed.WrapPolicy = text.WrapWords
				}
				// Controlled input sync: if external PF value changes, update the editor.
				if ShouldSyncExternalInput(state.inputExternal, path, externalVal, hasExternalValue) {
					if ed.Text() != externalVal {
						ed.SetText(externalVal)
						state.inputValues[path] = externalVal
					}
					state.inputExternal[path] = externalVal
				}
				// Process editor events (ChangeEvent, SubmitEvent)
				for {
					e, ok := ed.Update(gtx)
					if !ok {
						break
					}
					switch e.(type) {
					case widget.ChangeEvent:
						newTextRaw := ed.Text()
						newText := coreinput.NormalizeTypedInputValue(inputType, newTextRaw, props, false)
						if newText != newTextRaw {
							ed.SetText(newText)
						}
						state.inputValues[path] = newText
						state.inputExternal[path] = newText
						evName := coreengine.ResolveOnInputEvent(props)
						payloadMap := coreengine.BuildInputValueEventPayload(path, inputType, newText, true, time.Now().UnixMilli())
						_ = dispatchHostEvent(host, evName, payloadMap)
					case widget.SubmitEvent:
						finalText := coreinput.NormalizeTypedInputValue(inputType, ed.Text(), props, true)
						ed.SetText(finalText)
						state.inputValues[path] = finalText
						state.inputExternal[path] = finalText
						evSubmit := coreengine.ResolveOnSubmitEvent(props)
						payloadMap2 := coreengine.BuildInputValueEventPayload(path, inputType, finalText, state.inputFocused[path], time.Now().UnixMilli())
						_ = dispatchHostEvent(host, evSubmit, payloadMap2)
						evChange := coreengine.ResolveOnChangeEvent(props)
						if evChange != "" && evChange != evSubmit {
							_ = dispatchHostEvent(host, evChange, payloadMap2)
						}
					}
				}
				metrics := BuildInputBoxMetrics(css, inputType, isTextarea, w, h, padX, padY, 13)
				cssPadL := metrics.PadL
				cssPadT := metrics.PadT
				cssPadB := metrics.PadB
				spinnerW := metrics.SpinnerW
				pickerW := metrics.PickerW
				contentPadR := metrics.ContentPadR
				fontSize = metrics.FontSize
				ph := anyToString(props["placeholder"], "")
				// Build paint materials for the editor
				textRec := op.Record(gtx.Ops)
				paint.ColorOp{Color: textColor}.Add(gtx.Ops)
				textMat := textRec.Stop()
				selRec := op.Record(gtx.Ops)
				paint.ColorOp{Color: color.NRGBA{R: 0x4A, G: 0xDE, B: 0x80, A: 0x60}}.Add(gtx.Ops)
				selMat := selRec.Stop()
				// Layout editor at (x+padX, y+padY) constrained to input content box.
				edW := max(1, w-cssPadL-contentPadR)
				var edH, edOffY int
				contentW := edW
				offX := x + cssPadL
				if isTextarea {
					edH = max(1, h-cssPadT-cssPadB)
					edOffY = cssPadT
				} else {
					// Single-line: keep one visual line, then horizontally pan by caret position.
					edH = max(1, h-cssPadT-cssPadB)
					edOffY = cssPadT
					fullText := ed.Text()
					_, caretCol := ed.CaretPos()
					plan := BuildSingleLineInputLayoutPlan(fullText, fontSize, css, caretCol, edW, state.inputScrollX[path])
					state.inputScrollX[path] = plan.ScrollX
					state.inputContentW[path] = plan.FullWidth
					state.inputVisibleW[path] = edW
					contentW = plan.ContentWidth
					offX = x + cssPadL - plan.ScrollX
				}
				// Draw placeholder in the same inner content box as the editor/caret
				// so vertical alignment matches the text cursor baseline region.
				if ph != "" && ed.Text() == "" {
					drawGioText(gtx, state, x+cssPadL, y+edOffY, max(1, w-cssPadL-contentPadR), edH, ph, phColor,
						fontSize, false, false, false, text.Start, 1, false)
				}
				edClip := clip.Rect(image.Rect(x, y, x+w, y+h)).Push(gtx.Ops)
				offs := op.Offset(image.Pt(offX, y+edOffY)).Push(gtx.Ops)
				edGtx := gtx
				edGtx.Constraints = layout.Exact(image.Pt(contentW, edH))
				ed.Layout(edGtx, state.shaper, font.Font{}, unit.Sp(float32(fontSize)), textMat, selMat)
				offs.Pop()
				edClip.Pop()

				if !isTextarea {
					totalW := state.inputContentW[path]
					scrollPlan := BuildInputScrollIndicatorPlan(x, y, h, cssPadL, cssPadB, edW, totalW, state.inputScrollX[path])
					if scrollPlan.Visible {
						drawGioRRect(gtx.Ops, scrollPlan.TrackX, scrollPlan.TrackY, scrollPlan.TrackW, scrollPlan.TrackH, 1, color.NRGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xB0})
						drawGioRRect(gtx.Ops, scrollPlan.TrackX+scrollPlan.ThumbX, scrollPlan.TrackY, scrollPlan.ThumbW, scrollPlan.TrackH, 1, color.NRGBA{R: 0x94, G: 0xA3, B: 0xB8, A: 0xE0})
					}
				}

				inputHasFocus := state.inputFocused[path]

				// Only show spinner if input is focused
				spinnerPlan := BuildSpinnerLayoutPlan(x, y, w, h, spinnerW, inputHasFocus)
				if spinnerPlan.Visible {
					spinBorder := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "border-color", "#475569"), color.NRGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xFF}))
					spinFg := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "color", "#e8e8e8"), color.NRGBA{R: 0xE8, G: 0xE8, B: 0xE8, A: 0xFF}))
					drawGioRect(gtx.Ops, spinnerPlan.X, spinnerPlan.Y, 1, spinnerPlan.H, spinBorder)
					drawGioRect(gtx.Ops, spinnerPlan.X, spinnerPlan.Y+spinnerPlan.H/2, spinnerPlan.W, 1, spinBorder)
					drawGioText(gtx, state, spinnerPlan.X, spinnerPlan.Y, spinnerPlan.W, spinnerPlan.H/2, "â–²", spinFg, max(9, fontSize-2), false, false, false, text.Middle, 1, false)
					drawGioText(gtx, state, spinnerPlan.X, spinnerPlan.Y+spinnerPlan.H/2, spinnerPlan.W, spinnerPlan.H-spinnerPlan.H/2, "â–¼", spinFg, max(9, fontSize-2), false, false, false, text.Middle, 1, false)

					spinPath := coreengine.BuildInputSpinnerPath(path)
					capturedSpinPath := path
					capturedSpinX := spinnerPlan.X
					capturedSpinY := spinnerPlan.Y
					capturedSpinH := spinnerPlan.H
					capturedSpinHost := host
					capturedSpinProps := props
					capturedSpinEditor := ed
					state.handlers[spinPath] = func() {
						py := state.lastPointer[spinPath].Y
						delta := coreengine.ResolveSpinnerDelta(py, capturedSpinY, capturedSpinH)
						nextVal := coreinput.StepNumberInputValue(capturedSpinEditor.Text(), capturedSpinProps, delta)
						if nextVal == "" {
							return
						}
						capturedSpinEditor.SetText(nextVal)
						state.inputValues[capturedSpinPath] = nextVal
						evName := coreengine.ResolveOnInputEvent(capturedSpinProps)
						payload := coreengine.BuildInputValueEventPayload(capturedSpinPath, "number", nextVal, true, time.Now().UnixMilli())
						_ = dispatchHostEvent(capturedSpinHost, evName, payload)
						evChange := coreengine.ResolveOnChangeEvent(capturedSpinProps)
						if evChange != "" && evChange != evName {
							_ = dispatchHostEvent(capturedSpinHost, evChange, payload)
						}
						invalidateHost(capturedSpinHost)
					}
					state.propsForPath[spinPath] = props
					state.cssForPath[spinPath] = css
					registerGioEventArea(gtx.Ops, capturedSpinX, capturedSpinY, spinnerW, capturedSpinH, state.GetTag(spinPath))
				}

				// Get direction property for date/time inputs
				pickerDirection := coreengine.ResolvePickerDirection(props, "left")

				// Only show picker icon if input is focused
				pickerPlan := BuildPickerLayoutPlan(x, y, w, h, pickerW, pickerDirection, inputHasFocus)
				if pickerPlan.Visible {
					pickerBorder := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "border-color", "#475569"), color.NRGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xFF}))
					pickerFg := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "color", "#e8e8e8"), color.NRGBA{R: 0xE8, G: 0xE8, B: 0xE8, A: 0xFF}))
					drawGioRect(gtx.Ops, pickerPlan.X, pickerPlan.Y, 1, pickerPlan.H, pickerBorder)
					drawGioText(gtx, state, pickerPlan.X, pickerPlan.Y, pickerPlan.W, pickerPlan.H, "â–¼", pickerFg, max(10, fontSize-1), false, false, false, text.Middle, 1, false)

					pickerPath := coreengine.BuildInputPickerPath(path)
					capturedPickerPath := path
					capturedPickerHost := host
					capturedPickerProps := props
					capturedPickerType := inputType
					capturedPickerEditor := ed
					// Also open the modal picker
					state.handlers[pickerPath] = func() {
						// Open modal picker
						state.pickerModalOpen = capturedPickerPath
						state.pickerType = capturedPickerType
						state.pickerValue = coreinput.NormalizeTypedInputValue(capturedPickerType, capturedPickerEditor.Text(), capturedPickerProps, true)

						// Dispatch event
						showPickerEvent := coreengine.ResolveOnShowPickerEvent(capturedPickerProps)
						payload := coreengine.BuildInputValueEventPayload(capturedPickerPath, capturedPickerType, state.pickerValue, true, time.Now().UnixMilli())
						_ = dispatchHostEvent(capturedPickerHost, showPickerEvent, payload)
						invalidateHost(capturedPickerHost)
					}
					state.propsForPath[pickerPath] = props
					state.cssForPath[pickerPath] = css
					registerGioEventArea(gtx.Ops, pickerPlan.X, pickerPlan.Y, pickerPlan.W, pickerPlan.H, state.GetTag(pickerPath))
				}

				// Register the base input area so pointer-based focus transitions
				// work consistently for text/number/date/time inputs.
				state.propsForPath[path] = props
				state.cssForPath[path] = css
				registerGioEventArea(gtx.Ops, x, y, w, h, state.GetTag(path))
			}
			_ = padX
			_ = padY

		case "native":
			component := strings.ToLower(anyToString(props["component"], anyToString(props["native"], "label")))
			switch component {
			case "image", "img", "svg":
				src := anyToString(props["src"], anyToString(props["path"], ""))
				if src != "" {
					img, err := LoadRasterImage(src)
					if err == nil {
						radii := cssBorderRadiiValues(css, w, h)
						filtered := ApplyImageFilters(img, strings.TrimSpace(css["filter"]), strings.TrimSpace(css["opacity"]), CornerRadii{NW: radii.nw, NE: radii.ne, SE: radii.se, SW: radii.sw})
						drawGioImage(gtx.Ops, x, y, w, h, filtered)
					} else {
						drawGioRect(gtx.Ops, x, y, w, h, color.NRGBA{R: 60, G: 30, B: 30, A: 200})
						drawGioText(gtx, state, x, y, w, h, "[img]",
							color.NRGBA{R: 200, G: 100, B: 100, A: 255}, 11,
							false, false, false, text.Middle, 1, false)
					}
				}

			case "progress", "progressbar":
				progressPlan := BuildProgressVisualPlan(props["value"], w, true)
				trackCol := toNRGBA(uicore.ParseHexColor(anyToString(props["track"], "#444444"), color.NRGBA{R: 68, G: 68, B: 68, A: 255}))
				fillCol := toNRGBA(uicore.ParseHexColor(anyToString(props["fill"], "#4caf50"), color.NRGBA{R: 76, G: 175, B: 80, A: 255}))
				drawGioRRect(gtx.Ops, x, y, w, h, radius, trackCol)
				if progressPlan.FillW > 0 {
					drawGioRRect(gtx.Ops, x, y, progressPlan.FillW, h, radius, fillCol)
				}

			case "slider":
				sliderState := ResolveSliderState(path, props, state.sliderValues, 0, 100)
				plan := BuildSliderVisualPlan(sliderState.MinV, sliderState.MaxV, sliderState.Value, w, h, 4)
				accentSlider := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "accent-color", "#4ade80"), color.NRGBA{R: 0x4A, G: 0xDE, B: 0x80, A: 0xFF}))
				trackCol := color.NRGBA{R: 68, G: 68, B: 68, A: 255}
				drawGioRRect(gtx.Ops, x, y+h/2-3, w, 6, 3, trackCol)
				if plan.FillW > 0 {
					drawGioRRect(gtx.Ops, x, y+h/2-3, plan.FillW, 6, 3, accentSlider)
				}
				drawGioRRect(gtx.Ops, x+plan.ThumbX-plan.ThumbR, y+h/2-plan.ThumbR, plan.ThumbR*2, plan.ThumbR*2, plan.ThumbR, accentSlider)
				// Register drag interaction
				state.boundsForPath[path] = image.Rect(x, y, x+w, y+h)
				eventNameSlider := coreengine.ResolveComponentEventName(props, "change")
				registerSliderPointerHandler(state, host, path, sliderState.MinV, sliderState.MaxV, eventNameSlider, "slider")
				props["__interactive"] = "slider"
				state.propsForPath[path] = props
				state.cssForPath[path] = css
				registerGioEventArea(gtx.Ops, x, y, w, h, state.GetTag(path))

			case "checkbox":
				checked := ResolveCheckboxChecked(path, props, state.boolValues, state.inputValues)
				accentChk := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "accent-color", "#4ade80"), color.NRGBA{R: 0x4A, G: 0xDE, B: 0x80, A: 0xFF}))
				borderChk := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "border-color", "#475569"), color.NRGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xFF}))
				binaryLayout := BuildBinaryControlLayout(x, y, w, h, 4)
				bgChk := color.NRGBA{R: 30, G: 41, B: 59, A: 255}
				if checked {
					bgChk = accentChk
				}
				drawGioRRect(gtx.Ops, binaryLayout.BoxX, binaryLayout.BoxY, binaryLayout.BoxSize, binaryLayout.BoxSize, 3, bgChk)
				drawGioBorder(gtx.Ops, binaryLayout.BoxX, binaryLayout.BoxY, binaryLayout.BoxSize, binaryLayout.BoxSize, 3, 2, borderChk)
				if checked {
					drawGioText(gtx, state, binaryLayout.BoxX, binaryLayout.BoxY, binaryLayout.BoxSize, binaryLayout.BoxSize, "âœ“",
						color.NRGBA{R: 15, G: 23, B: 25, A: 255}, uicore.CSSFontSize(css, 13), true, false, false, text.Middle, 1, false)
				}
				chkEvent := coreengine.ResolveComponentEventName(props, "change")
				registerCheckboxToggleHandler(state, host, path, chkEvent)
				state.propsForPath[path] = props
				state.cssForPath[path] = css
				registerGioEventArea(gtx.Ops, x, y, w, h, state.GetTag(path))

			case "radio":
				radioModel := ResolveRadioGroupModel(path, props, state.inputValues)
				accentRad := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "accent-color", "#4ade80"), color.NRGBA{R: 0x4A, G: 0xDE, B: 0x80, A: 0xFF}))
				borderRad := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "border-color", "#475569"), color.NRGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xFF}))
				binaryLayout := BuildBinaryControlLayout(x, y, w, h, 4)
				drawGioRRect(gtx.Ops, binaryLayout.BoxX, binaryLayout.BoxY, binaryLayout.BoxSize, binaryLayout.BoxSize, binaryLayout.BoxSize/2, color.NRGBA{R: 30, G: 41, B: 59, A: 255})
				drawGioBorder(gtx.Ops, binaryLayout.BoxX, binaryLayout.BoxY, binaryLayout.BoxSize, binaryLayout.BoxSize, binaryLayout.BoxSize/2, 2, borderRad)
				if radioModel.Selected {
					drawGioRRect(gtx.Ops, binaryLayout.BoxX+(binaryLayout.BoxSize-binaryLayout.Inner)/2, binaryLayout.BoxY+(binaryLayout.BoxSize-binaryLayout.Inner)/2, binaryLayout.Inner, binaryLayout.Inner, binaryLayout.Inner/2, accentRad)
				}
				radEvent := coreengine.ResolveComponentEventName(props, "change")
				registerRadioSelectHandler(state, host, path, radioModel.GroupKey, radioModel.Value, radEvent)
				state.propsForPath[path] = props
				state.cssForPath[path] = css
				registerGioEventArea(gtx.Ops, x, y, w, h, state.GetTag(path))

			case "select", "dropdown":
				selectModel := ResolveSelectModel(path, props, state.inputValues)
				fgSel := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "color", "#e2e8f0"), color.NRGBA{R: 0xE2, G: 0xE8, B: 0xF0, A: 0xFF}))
				menuOpen := state.selectMenuOpen == path
				// Draw selected value + arrow indicator
				arrowW := 20
				drawGioText(gtx, state, x+6, y, max(1, w-arrowW-6), h, selectModel.SelectedLabel, fgSel, uicore.CSSFontSize(css, 13),
					false, false, false, text.Start, 1, false)
				arrowGlyph := "â–¾"
				if menuOpen {
					arrowGlyph = "â–´"
				}
				drawGioText(gtx, state, x+w-arrowW, y, arrowW, h, arrowGlyph, fgSel, uicore.CSSFontSize(css, 11),
					false, false, false, text.Middle, 1, false)
				selEvent := coreengine.ResolveComponentEventName(props, "change")
				registerSelectToggleHandler(state, host, path)
				state.boundsForPath[path] = image.Rect(x, y, x+w, y+h)
				state.propsForPath[path] = props
				state.cssForPath[path] = css
				registerGioEventArea(gtx.Ops, x, y, w, h, state.GetTag(path))
				_ = selEvent

			case "textarea":
				externalTA, hasExternalTextarea := coreinput.InputExternalValueWithPresence(props)
				editorCfgTA := BuildInputEditorConfig("textarea")
				edTA := state.editors[path]
				if edTA == nil {
					edTA = new(widget.Editor)
					edTA.SingleLine = editorCfgTA.SingleLine
					if editorCfgTA.UseHeuristicWrap {
						edTA.WrapPolicy = text.WrapHeuristically
					} else {
						edTA.WrapPolicy = text.WrapWords
					}
					edTA.MaxLen = editorCfgTA.MaxLen
					edTA.Mask = editorCfgTA.Mask
					if externalTA != "" {
						edTA.SetText(externalTA)
						state.inputValues[path] = externalTA
					}
					state.inputExternal[path] = externalTA
					state.editors[path] = edTA
				}
				if edTA.SingleLine != editorCfgTA.SingleLine {
					edTA.SingleLine = editorCfgTA.SingleLine
				}
				if editorCfgTA.UseHeuristicWrap {
					edTA.WrapPolicy = text.WrapHeuristically
				} else {
					edTA.WrapPolicy = text.WrapWords
				}
				edTA.MaxLen = editorCfgTA.MaxLen
				edTA.Mask = editorCfgTA.Mask
				if ShouldSyncExternalInput(state.inputExternal, path, externalTA, hasExternalTextarea) {
					if edTA.Text() != externalTA {
						edTA.SetText(externalTA)
						state.inputValues[path] = externalTA
					}
					state.inputExternal[path] = externalTA
				}
				for {
					e, ok := edTA.Update(gtx)
					if !ok {
						break
					}
					if _, isChange := e.(widget.ChangeEvent); isChange {
						newTextTA := edTA.Text()
						state.inputValues[path] = newTextTA
						state.inputExternal[path] = newTextTA
					}
				}
				fgTA := toNRGBA(uicore.ParseHexColor(uicore.CSSGetColor(css, "color", "#e2e8f0"), color.NRGBA{R: 0xE2, G: 0xE8, B: 0xF0, A: 0xFF}))
				taRec := op.Record(gtx.Ops)
				paint.ColorOp{Color: fgTA}.Add(gtx.Ops)
				taMat := taRec.Stop()
				taSelRec := op.Record(gtx.Ops)
				paint.ColorOp{Color: color.NRGBA{R: 0x4A, G: 0xDE, B: 0x80, A: 0x60}}.Add(gtx.Ops)
				taSelMat := taSelRec.Stop()
				taLayout := BuildTextareaLayoutPlan(x, y, w, h, 6, 4)
				edClip2 := clip.Rect(image.Rect(x, y, x+w, y+h)).Push(gtx.Ops)
				offs2 := op.Offset(image.Pt(taLayout.OffsetX, taLayout.OffsetY)).Push(gtx.Ops)
				edGtx2 := gtx
				edGtx2.Constraints = layout.Exact(image.Pt(taLayout.EditorW, taLayout.EditorH))
				edTA.Layout(edGtx2, state.shaper, font.Font{}, unit.Sp(float32(uicore.CSSFontSize(css, 13))), taMat, taSelMat)
				offs2.Pop()
				edClip2.Pop()

			default:
				labelText := uicore.SanitizeRenderText(anyToString(props["text"], component))
				fgHex := uicore.CSSGetColor(css, "color", "#e8e8e8")
				fg := toNRGBA(uicore.ParseHexColor(fgHex, color.NRGBA{R: 0xE8, G: 0xE8, B: 0xE8, A: 0xFF}))
				drawGioText(gtx, state, x, y, w, h, labelText, fg, uicore.CSSFontSize(css, 13),
					false, false, false, text.Start, 0, true)
			}

		default:
			// view/container: background already drawn â€” no extra content
		}

		// Register event area only when needed (hover/active selectors or scroll containers)
		// to avoid intercepting pointer focus from child widget.Editor inputs.
		hasHover := appSS != nil && uicore.HasHoverSelector(props, appSS)
		hasActive := appSS != nil && uicore.HasActiveSelector(props, appSS)
		hasPointerHandlers := coreengine.HasPointerEventHandlers(props)
		if corelayout.ShouldRegisterContainerEvents(kind, hasHover, hasActive, overflowX, overflowY) || hasPointerHandlers {
			state.propsForPath[path] = props
			state.cssForPath[path] = css
			registerGioEventArea(gtx.Ops, x, y, w, h, state.GetTag(path))
		}
	}

	var childClip clip.Stack
	var scrollTrStack op.TransformStack
	var hasScrollTr bool
	childPlan := corelayout.BuildChildTraversalPlan(visible, isOnScreen, clipChildren, inheritedShiftX, inheritedShiftY, translateX, translateY, state.scrollOffsets[path])
	childShiftX := childPlan.ChildShiftX + positionDeltaX
	childShiftY := childPlan.ChildShiftY + positionDeltaY
	if childPlan.EnableClip {
		childClip = clip.Rect(image.Rect(x, y, x+w, y+h)).Push(gtx.Ops)
		so := state.scrollOffsets[path]
		if childPlan.ApplyScrollTransform {
			scrollTrStack = op.Offset(image.Pt(-so.X, -so.Y)).Push(gtx.Ops)
			hasScrollTr = true
		}
	}

	if len(childrenSlice) > 0 {
		var positionTrStack op.TransformStack
		hasPositionTr := false
		if positionDeltaX != 0 || positionDeltaY != 0 {
			positionTrStack = op.Offset(image.Pt(positionDeltaX, positionDeltaY)).Push(gtx.Ops)
			hasPositionTr = true
		}
		childMaps := make([]map[string]any, 0, len(childrenSlice))
		for _, child := range childrenSlice {
			childMaps = append(childMaps, anyToMap(child))
		}
		childSig := childSliceSignatureAny(childrenSlice)
		existing := (*ZSortHint)(nil)
		if hint, ok := state.zChildrenHint[path]; ok {
			existing = &ZSortHint{Count: hint.Count, Sig: hint.Sig, Needs: hint.Needs, LastSeen: hint.LastSeen}
		}
		needsZSort, nextHint := ResolveChildrenZSortNeed(childMaps, existing, childSig, state.frameNumber)
		state.zChildrenHint[path] = uicore.ZChildrenHintCacheEntry{Count: nextHint.Count, Sig: nextHint.Sig, Needs: nextHint.Needs, LastSeen: nextHint.LastSeen}

		decision := PlanChildTraversal(childMaps, isOnScreen, needsZSort)
		if decision.Traverse {
			for _, idx := range decision.Order {
				DrawGioTree(rctx, childMaps[idx], host, coreengine.ChildPath(path, idx), state, "", childContainingBlock, css, childShiftX, childShiftY, frameTime)
			}
		}
		if hasPositionTr {
			positionTrStack.Pop()
		}
	}

	showScrollbars := overflowX == "auto" || overflowX == "scroll" || overflowY == "auto" || overflowY == "scroll"
	if visible && isOnScreen && clipChildren {
		if hasScrollTr {
			scrollTrStack.Pop()
		}
		childClip.Pop()
	}

	if visible && isOnScreen && clipChildren && showScrollbars {
		radius := roundedFromProps(props, css, w, h)
		contentRight, contentBottom := childrenContentBounds(childrenSlice, x+w, y+h)
		so := coreinput.ClampScrollToContent(state.scrollOffsets[path], x, y, w, h, contentRight, contentBottom)
		state.scrollOffsets[path] = so
		hitInfo := drawScrollbars(gtx, x, y, w, h, contentRight, contentBottom, css, radius, so.X, so.Y)
		if hitInfo.HasV || hitInfo.HasH {
			state.scrollHits[path] = hitInfo
		}
	}

	if visible && isOnScreen && state.frameDebug {
		contentRight, contentBottom := x+w, y+h
		if clipChildren {
			contentRight, contentBottom = childrenContentBounds(childrenSlice, x+w, y+h)
		}
		drawGioDebugOverlay(gtx, state, host, path, kind, x, y, w, h, css, clipChildren, contentRight, contentBottom)
	}

	if hasNodeScaleTransform {
		nodeScaleTransform.Pop()
	}
	if hasNodeRotateTransform {
		nodeRotateTransform.Pop()
	}
	if hasNodeTranslateTransform {
		nodeTranslateTransform.Pop()
	}
}

// drawPickerModal renders a date/time picker modal overlay that appears on top of all content.
func DrawPickerModal(rctx *coreengine.RenderContext, state *GioWindowState, screenW, screenH int) {
	gtx := rctx.Gtx
	if state.pickerModalOpen == "" {
		return
	}

	model := BuildPickerModalModel(screenW, screenH, state.pickerType, state.pickerValue, state.pickerModalOpen)

	drawGioRect(gtx.Ops, 0, 0, screenW, screenH, model.OverlayColor)
	drawGioRRect(gtx.Ops, model.ModalRect.Min.X, model.ModalRect.Min.Y, model.ModalRect.Dx(), model.ModalRect.Dy(), model.ModalRadius, model.ModalBg)
	drawGioRect(gtx.Ops, model.ModalRect.Min.X, model.ModalRect.Min.Y, model.ModalRect.Dx(), 1, model.BorderColor)
	drawGioRect(gtx.Ops, model.ModalRect.Min.X, model.ModalRect.Max.Y-1, model.ModalRect.Dx(), 1, model.BorderColor)
	drawGioRect(gtx.Ops, model.ModalRect.Min.X, model.ModalRect.Min.Y, 1, model.ModalRect.Dy(), model.BorderColor)
	drawGioRect(gtx.Ops, model.ModalRect.Max.X-1, model.ModalRect.Min.Y, 1, model.ModalRect.Dy(), model.BorderColor)

	drawGioText(gtx, state, model.TitleRect.Min.X, model.TitleRect.Min.Y, model.TitleRect.Dx(), model.TitleRect.Dy(), model.TitleText, model.TitleColor, 16, true, false, false, text.Start, 1, false)
	drawGioRect(gtx.Ops, model.SepRect.Min.X, model.SepRect.Min.Y, model.SepRect.Dx(), model.SepRect.Dy(), model.BorderColor)
	drawGioText(gtx, state, model.ValueRect.Min.X, model.ValueRect.Min.Y, model.ValueRect.Dx(), model.ValueRect.Dy(), model.ValueText, model.ValueColor, 14, false, false, false, text.Middle, 1, false)

	drawGioRRect(gtx.Ops, model.DecRect.Min.X, model.DecRect.Min.Y, model.DecRect.Dx(), model.DecRect.Dy(), model.ButtonRadius, model.DecColor)
	drawGioText(gtx, state, model.DecRect.Min.X, model.DecRect.Min.Y, model.DecRect.Dx(), model.DecRect.Dy(), "âˆ’", model.TitleColor, 16, true, false, false, text.Middle, 1, false)
	state.handlers[model.DecPath] = func() {}
	registerGioEventArea(gtx.Ops, model.DecRect.Min.X, model.DecRect.Min.Y, model.DecRect.Dx(), model.DecRect.Dy(), state.GetTag(model.DecPath))

	drawGioRRect(gtx.Ops, model.IncRect.Min.X, model.IncRect.Min.Y, model.IncRect.Dx(), model.IncRect.Dy(), model.ButtonRadius, model.IncColor)
	drawGioText(gtx, state, model.IncRect.Min.X, model.IncRect.Min.Y, model.IncRect.Dx(), model.IncRect.Dy(), "+", model.TitleColor, 16, true, false, false, text.Middle, 1, false)
	state.handlers[model.IncPath] = func() {}
	registerGioEventArea(gtx.Ops, model.IncRect.Min.X, model.IncRect.Min.Y, model.IncRect.Dx(), model.IncRect.Dy(), state.GetTag(model.IncPath))

	drawGioRRect(gtx.Ops, model.CloseRect.Min.X, model.CloseRect.Min.Y, model.CloseRect.Dx(), model.CloseRect.Dy(), model.ButtonRadius, model.CloseColor)
	drawGioText(gtx, state, model.CloseRect.Min.X, model.CloseRect.Min.Y, model.CloseRect.Dx(), model.CloseRect.Dy(), "Close", model.TitleColor, 12, false, false, false, text.Middle, 1, false)

	state.handlers[model.ClosePath] = func() {
		state.pickerModalOpen = ""
		state.pickerType = ""
		state.pickerValue = ""
	}
	registerGioEventArea(gtx.Ops, model.CloseRect.Min.X, model.CloseRect.Min.Y, model.CloseRect.Dx(), model.CloseRect.Dy(), state.GetTag(model.ClosePath))
}
