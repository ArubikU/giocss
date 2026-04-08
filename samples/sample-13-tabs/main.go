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
  width: 100%;
  height: 100%;
  margin: 0;
  background: linear-gradient(135deg, #f8fafc 0%, #eef2ff 100%);
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 26px;
}

.shell {
  width: 100%;
  max-width: 920px;
  background-color: #ffffff;
  border-radius: 14px;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.12);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.header {
  padding: 16px 18px;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
}

.title {
  font-size: 22px;
  color: #0f172a;
  font-weight: bold;
}

.beta {
  background-color: #f43f5e;
  color: #ffffff;
  padding: 5px 10px;
  border-radius: 8px;
  font-size: 11px;
  font-weight: bold;
  transform: rotate(-10deg);
}

.tabs {
  display: flex;
  flex-direction: row;
  gap: 8px;
  padding: 12px 14px;
  border-bottom: 1px solid #e2e8f0;
}

.tab-btn {
  padding: 10px 14px;
  border-radius: 8px;
  border: 1px solid #cbd5e1;
  background-color: #f8fafc;
  color: #334155;
  font-size: 13px;
  cursor: pointer;
}
.tab-btn:hover {
  background-color: #eef2ff;
}
.tab-btn-active {
  background-color: #1d4ed8;
  color: #ffffff;
  border-color: #1d4ed8;
}

.panel {
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.panel-title {
  font-size: 21px;
  font-weight: bold;
  color: #0f172a;
}

.panel-text {
  font-size: 14px;
  color: #475569;
  line-height: 1.5;
}

.kpi-row {
  display: flex;
  flex-direction: row;
  gap: 10px;
}

.kpi {
  flex: 1;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.kpi-label {
  font-size: 12px;
  color: #64748b;
}

.kpi-value {
  font-size: 20px;
  color: #0f172a;
  font-weight: bold;
}
`

type tabsState struct {
	mu     sync.Mutex
	active string
}

func newTabsState() *tabsState {
	return &tabsState{active: "overview"}
}

func (s *tabsState) Select(name string) {
	s.active = name
}

func (s *tabsState) onEvent(eventName string, _ map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch strings.TrimSpace(eventName) {
	case "tabs.overview":
		s.Select("overview")
	case "tabs.metrics":
		s.Select("metrics")
	case "tabs.activity":
		s.Select("activity")
	}
	return nil
}

func (s *tabsState) snapshot() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

func tabButton(label, eventName, activeTab, current string) *giocss.Node {
	btn := components.Button(label, "tab-btn")
	btn.SetProp("event", eventName)
	if activeTab == current {
		btn.AddClass("tab-btn-active")
	}
	return btn
}

func buildPanel(active string) *giocss.Node {
	panel := components.Column()
	panel.AddClass("panel")

	switch active {
	case "metrics":
		panel.AddChild(components.Heading(2, "Metrics", "panel-title"))
		panel.AddChild(components.Text("Quick status of the project KPIs.", "panel-text"))
		row := components.Row(
			components.Column(components.Text("Coverage", "kpi-label"), components.Text("91%", "kpi-value")),
			components.Column(components.Text("Latency", "kpi-label"), components.Text("42ms", "kpi-value")),
			components.Column(components.Text("Errors", "kpi-label"), components.Text("0.24%", "kpi-value")),
		)
		row.AddClass("kpi-row")
		for _, ch := range row.Children {
			ch.AddClass("kpi")
		}
		panel.AddChild(row)
	case "activity":
		panel.AddChild(components.Heading(2, "Activity", "panel-title"))
		panel.AddChild(components.Text("Recent actions for the active session.", "panel-text"))
		panel.AddChild(components.Text("- User signed in from desktop.", "panel-text"))
		panel.AddChild(components.Text("- Dashboard refreshed 12 seconds ago.", "panel-text"))
		panel.AddChild(components.Text("- 2 comments were added to release note.", "panel-text"))
	default:
		panel.AddChild(components.Heading(2, "Overview", "panel-title"))
		panel.AddChild(components.Text("This sample validates tabs and transform rotate support using the red beta badge in the header.", "panel-text"))
		panel.AddChild(components.Text("Switch tabs to verify stateful content swaps and active button styles.", "panel-text"))
	}

	return panel
}

func buildUI(active string) *giocss.Node {
	root := giocss.NewNode("body")

	shell := components.Column()
	shell.AddClass("shell")

	head := components.Row(components.Heading(1, "Sample 13 - Tabs", "title"), components.Badge("BETA", "beta"))
	head.AddClass("header")
	shell.AddChild(head)

	tabs := components.Row(
		tabButton("Overview", "tabs.overview", active, "overview"),
		tabButton("Metrics", "tabs.metrics", active, "metrics"),
		tabButton("Activity", "tabs.activity", active, "activity"),
	)
	tabs.AddClass("tabs")
	shell.AddChild(tabs)

	shell.AddChild(buildPanel(active))
	root.AddChild(shell)
	return root
}

func main() {
	ss := giocss.NewStyleSheet()
	ss.ParseCSSText(css)
	app := newTabsState()

	rt := giocss.NewWindowRuntime(
		giocss.WindowOptions{Title: "Sample 13 - Tabs", Width: 980, Height: 560},
		giocss.WindowRuntimeHooks{
			DispatchEvent: app.onEvent,
			Snapshot: func(size image.Point) giocss.WindowRuntimeSnapshot {
				active := app.snapshot()
				root := buildUI(active)
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
