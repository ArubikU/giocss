package main

import (
	"image"
	"strings"
	"sync"

	giocss "github.com/ArubikU/giocss"
	"github.com/ArubikU/giocss/components"
)

const css = `
body {
	background: linear-gradient(180deg, #f4f7fb 0%, #eef3fa 100%);
	width: 100%;
	height: 100%;
	display: flex;
	flex-direction: column;
	overflow: hidden;
	font-family: Segoe UI, Tahoma, sans-serif;
}

.app-header {
	height: 74px;
	width: 100%;
	display: flex;
	flex-direction: row;
	align-items: center;
	justify-content: space-between;
	padding: 0 18px;
	background: #0b223d;
	border-bottom: 1px solid #123a66;
}

.brand {
	font-size: 20px;
	font-weight: 800;
	color: #f6fbff;
	letter-spacing: 0.3px;
}

.header-pill {
	font-size: 11px;
	font-weight: 700;
	padding: 6px 10px;
	border-radius: 999px;
	color: #0b223d;
	background: #8cd8ff;
	border: 1px solid #5dc3f8;
}

.nav {
	display: flex;
	flex-direction: row;
	gap: 10px;
	min-width: 0;
	overflow-x: auto;
}

.nav-btn {
	padding: 9px 13px;
	font-size: 13px;
	font-weight: 700;
	border-radius: 10px;
	border: 1px solid #275381;
	background: #123459;
	color: #d9ebff;
	cursor: pointer;
}

.nav-btn:hover {
	background: #184670;
}

.nav-btn-active {
	background: #e9f6ff;
	color: #0b223d;
	border-color: #b9def7;
}

.main {
	width: 100%;
	flex: 1;
	min-height: 0;
	overflow: auto;
	padding: 20px;
	display: flex;
	flex-direction: column;
	gap: 14px;
}

.hero {
	background: #ffffff;
	border: 1px solid #d8e6f3;
	border-radius: 14px;
	padding: 18px;
	display: flex;
	flex-direction: column;
	gap: 6px;
}

.hero-title {
	font-size: 31px;
	font-weight: 800;
	color: #0f2f50;
	line-height: 1.16;
}

.hero-text {
	font-size: 14px;
	color: #446382;
	line-height: 1.45;
}

.panel {
	background: #ffffff;
	border: 1px solid #d8e6f3;
	border-radius: 14px;
	padding: 16px;
	display: flex;
	flex-direction: column;
	gap: 10px;
}

.panel-title {
	font-size: 18px;
	font-weight: 700;
	color: #123459;
}

.panel-text {
	font-size: 13px;
	line-height: 1.45;
	color: #4b6783;
}

.matrix-row {
	display: flex;
	flex-direction: row;
	align-items: center;
	justify-content: space-between;
	gap: 10px;
	border: 1px solid #e4edf6;
	border-radius: 10px;
	padding: 10px;
	background: #fbfdff;
}

.matrix-feature {
	font-size: 13px;
	font-weight: 700;
	color: #1d456b;
}

.matrix-notes {
	font-size: 12px;
	line-height: 1.35;
	color: #52708c;
}

.badge {
	font-size: 11px;
	font-weight: 800;
	padding: 5px 8px;
	border-radius: 999px;
	border: 1px solid transparent;
}

.badge-ok {
	background: #e8f8ef;
	color: #175f34;
	border-color: #c4ebd2;
}

.badge-mid {
	background: #fff8e5;
	color: #7a5a00;
	border-color: #f0e0a8;
}

.badge-no {
	background: #fdeaea;
	color: #7c1f1f;
	border-color: #f2c4c4;
}

.forms-grid {
	display: flex;
	flex-direction: row;
	gap: 14px;
	flex-wrap: wrap;
}

.forms-col {
	flex: 1;
	min-width: 280px;
	display: flex;
	flex-direction: column;
	gap: 10px;
}

.field-label {
	font-size: 12px;
	font-weight: 700;
	color: #30567a;
}

.field-input {
	padding: 8px 10px;
	border-radius: 9px;
	border: 1px solid #bcd2e5;
	background: #ffffff;
	color: #17344f;
}

.field-input:focus {
	border-color: #2f89c9;
	background: #f6fbff;
}

.field-input:invalid {
	border-color: #c63232;
	background: #fff4f4;
}

.field-input:disabled {
	background: #edf3f8;
	color: #7b92a8;
	border-color: #d4e1ec;
}

.demo-check {
	width: 18px;
	height: 18px;
	border-radius: 5px;
	border: 1px solid #9eb9d3;
	background: #ffffff;
}

.demo-check:checked {
	background: #1f7ab8;
	border-color: #1f7ab8;
}

.row-inline {
	display: flex;
	flex-direction: row;
	align-items: center;
	gap: 8px;
}
`

type docsState struct {
	mu   sync.Mutex
	view string
}

func newDocsState() *docsState {
	return &docsState{view: "overview"}
}

func (s *docsState) onEvent(eventName string, _ map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch strings.TrimSpace(eventName) {
	case "docs.nav.overview":
		s.view = "overview"
	case "docs.nav.forms":
		s.view = "forms"
	case "docs.nav.coverage":
		s.view = "coverage"
	}
	return nil
}

func (s *docsState) snapshotView() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.view
}

func navButton(label, eventName, activeView, currentView string) *giocss.Node {
	btn := components.Button(label, "nav-btn")
	btn.SetProp("event", eventName)
	if activeView == currentView {
		btn.AddClass("nav-btn-active")
	}
	return btn
}

func matrixRow(feature, badgeLabel, badgeClass, notes string) *giocss.Node {
	left := components.Column(
		components.Span(feature, "matrix-feature"),
		components.Span(notes, "matrix-notes"),
	)
	right := components.Badge(badgeLabel, "badge", badgeClass)
	row := components.Row(left, right)
	row.AddClass("matrix-row")
	return row
}

func buildOverviewPage() *giocss.Node {
	panel := components.Column(
		components.Heading(2, "Roadmap in this phase", "panel-title"),
		components.Span("This docs view behaves like an in-app route and tracks implementation coverage for forms and runtime parity.", "panel-text"),
		matrixRow("Pseudo-states for forms", "Implemented", "badge-ok", ":focus, :disabled, :checked and :invalid now resolve and cache correctly."),
		matrixRow("Disabled interaction semantics", "Implemented", "badge-ok", "Pointer press/drag now ignore disabled controls, so events and focus do not leak."),
		matrixRow("Advanced selectors", "Implemented", "badge-ok", "Combinators and :nth-child now resolve at render time; see sample-19 for a focused visual demo."),
		matrixRow("Components API additions", "Implemented", "badge-ok", "New constructors: Select, Option, Fieldset and Legend for HTML-like authoring."),
		matrixRow("Select semantics parity", "Partial", "badge-mid", "Renderer supports cycling options, but no full popup/menu list semantics yet."),
	)
	panel.AddClass("panel")
	return panel
}

func buildFormsPage() *giocss.Node {
	nameInput := components.Input("display-name", "Type your name", "text")
	nameInput.AddClass("field-input")
	nameInput.SetProp("required", "true")

	emailInput := components.Input("email", "name@example.com", "email")
	emailInput.AddClass("field-input")
	emailInput.SetProp("required", "true")

	disabledInput := components.Input("api-key", "Auto-generated key", "text")
	disabledInput.AddClass("field-input")
	disabledInput.SetProp("disabled", "true")
	disabledInput.SetProp("value", "pk_live_51_demo")

	roleSelect := components.Select("role", []*giocss.Node{
		components.Option("Choose a role", "", true, false),
		components.Option("Reader", "Reader", false, false),
		components.Option("Editor", "Editor", false, false),
		components.Option("Owner", "Owner", false, false),
	})
	roleSelect.AddClass("field-input")
	roleSelect.SetProp("placeholder", "Choose a role")
	roleSelect.SetProp("required", "true")

	checkbox := giocss.NewNode("input")
	checkbox.SetProp("type", "checkbox")
	checkbox.SetProp("checked", "true")
	checkbox.AddClass("demo-check")

	leftCol := components.Column(
		components.Heading(3, "Fieldset + validation states", "panel-title"),
		components.Fieldset("Account", []string{"panel"},
			components.Label("Display name", "field-label"),
			nameInput,
			components.Label("Email", "field-label"),
			emailInput,
			components.Label("Role (required)", "field-label"),
			roleSelect,
		),
	)
	leftCol.AddClass("forms-col")

	rightCol := components.Column(
		components.Heading(3, "Disabled + checked styles", "panel-title"),
		components.Fieldset("Runtime states", []string{"panel"},
			components.Label("Disabled input", "field-label"),
			disabledInput,
			components.Div([]string{"row-inline"}, checkbox, components.Span("Checked selector demo", "panel-text")),
			components.Span("Tip: click fields to trigger :focus, leave required fields empty to trigger :invalid, and try select cycling.", "panel-text"),
		),
	)
	rightCol.AddClass("forms-col")

	grid := components.Row(leftCol, rightCol)
	grid.AddClass("forms-grid")

	panel := components.Column(
		components.Heading(2, "Forms HTML-like integration", "panel-title"),
		components.Span("This page intentionally demonstrates the new pseudo-states with live controls.", "panel-text"),
		components.Span("Inputs also preserve typed text across state-driven re-renders now, which unblocks more realistic form flows for upcoming samples.", "panel-text"),
		grid,
	)
	panel.AddClass("panel")
	return panel
}

func buildCoveragePage() *giocss.Node {
	panel := components.Column(
		components.Heading(2, "Coverage matrix", "panel-title"),
		components.Span("High-level parity snapshot for current giocss runtime.", "panel-text"),
		matrixRow("Layout primitives (flex, gap, abs/fixed)", "Implemented", "badge-ok", "Core layout reconciliation and intrinsic sizing are in place with recent bug fixes."),
		matrixRow("Text rendering and clipping resilience", "Implemented", "badge-ok", "Descender clipping and placeholder alignment regressions were addressed in render/runtime."),
		matrixRow("Pseudo-classes :hover / :active", "Implemented", "badge-ok", "Stable and used by multiple interactive samples."),
		matrixRow("Pseudo-classes :focus / :disabled / :checked / :invalid", "Implemented", "badge-ok", "Newly added in this phase with cache-aware invalidation."),
		matrixRow("Native select dropdown menu parity", "Partial", "badge-mid", "Single-control cycle model exists; popup list semantics remain future work."),
		matrixRow("Advanced selectors (:nth-child, combinators)", "Implemented", "badge-ok", "Descendant, child and sibling combinators plus :nth-child are now evaluated in render-time matching."),
		matrixRow("Animation timing/function parity", "Partial", "badge-mid", "Transitions exist, but full browser timing feature parity is not complete."),
	)
	panel.AddClass("panel")
	return panel
}

func buildUI(view string) *giocss.Node {
	root := giocss.NewNode("body")

	header := components.Row(
		components.Span("giocss docs", "brand"),
		components.Row(
			navButton("Overview", "docs.nav.overview", "overview", view),
			navButton("Forms", "docs.nav.forms", "forms", view),
			navButton("Coverage", "docs.nav.coverage", "coverage", view),
		),
		components.Badge("state-driven route", "header-pill"),
	)
	header.AddClass("app-header")
	header.Children[1].AddClass("nav")
	root.AddChild(header)

	hero := components.Column(
		components.Heading(1, "In-app docs route", "hero-title"),
		components.Span("This sample behaves like /docs in a browser app, but implemented with giocss state snapshots.", "hero-text"),
	)
	hero.AddClass("hero")

	mainCol := components.Column(hero)
	mainCol.AddClass("main")

	switch view {
	case "forms":
		mainCol.AddChild(buildFormsPage())
	case "coverage":
		mainCol.AddChild(buildCoveragePage())
	default:
		mainCol.AddChild(buildOverviewPage())
	}

	root.AddChild(mainCol)
	return root
}

func main() {
	ss := giocss.NewStyleSheet()
	ss.ParseCSSText(css)
	app := newDocsState()

	rt := giocss.NewWindowRuntime(
		giocss.WindowOptions{Title: "Sample 18 - Docs Viewer", Width: 1024, Height: 640},
		giocss.WindowRuntimeHooks{
			DispatchEvent: app.onEvent,
			Snapshot: func(size image.Point) giocss.WindowRuntimeSnapshot {
				root := buildUI(app.snapshotView())
				return giocss.WindowRuntimeSnapshot{
					RootLayout:   giocss.LayoutNodeToNative(root, size.X, size.Y, ss),
					RootCSS:      giocss.ResolveNodeStyle(root, ss, size.X),
					StyleSheet:   ss,
					ScreenWidth:  size.X,
					ScreenHeight: size.Y,
				}
			},
		},
	)
	rt.Run()
}
