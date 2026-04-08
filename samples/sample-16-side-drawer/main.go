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
  overflow: hidden;
  background: radial-gradient(circle at top, #e0f2fe 0%, #f8fafc 45%, #e2e8f0 100%);
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 24px;
}

.shell {
  width: 100%;
  max-width: 980px;
  background-color: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 18px;
  box-shadow: 0 18px 42px rgba(15, 23, 42, 0.12);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.hero {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 18px;
  border-bottom: 1px solid #e2e8f0;
}

.eyebrow {
  align-self: flex-start;
  background-color: #dbeafe;
  color: #1d4ed8;
  border-radius: 999px;
  padding: 4px 10px;
  font-size: 11px;
  font-weight: bold;
}

.title {
  font-size: 26px;
  font-weight: bold;
  color: #0f172a;
}

.subtitle {
  font-size: 14px;
  color: #475569;
  line-height: 1.55;
}

.toolbar {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 10px;
}

.menu-btn,
.section-btn,
.close-btn {
  padding: 10px 14px;
  border-radius: 10px;
  border: 1px solid #cbd5e1;
  background-color: #ffffff;
  color: #0f172a;
  font-size: 13px;
  cursor: pointer;
}
.menu-btn:hover,
.section-btn:hover,
.close-btn:hover {
  background-color: #eff6ff;
}

.section-btn-active {
  background-color: #1d4ed8;
  border-color: #1d4ed8;
  color: #ffffff;
}

.content {
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.content-card {
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.card-title {
  font-size: 20px;
  font-weight: bold;
  color: #0f172a;
}

.card-copy {
  font-size: 14px;
  color: #475569;
  line-height: 1.5;
}

.stats {
  display: flex;
  flex-direction: row;
  gap: 10px;
}

.stat {
  flex: 1;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background-color: #f8fafc;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.stat-label {
  font-size: 12px;
  color: #64748b;
}

.stat-value {
  font-size: 22px;
  font-weight: bold;
  color: #0f172a;
}

.stat-note {
  font-size: 12px;
  color: #475569;
}

.drawer-overlay {
  position: fixed;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  border: 0;
  background-color: rgba(15, 23, 42, 0.46);
  z-index: 20;
  color: transparent;
  font-size: 0;
}

.drawer-panel {
  position: fixed;
  left: 0;
  top: 0;
  width: 320px;
  height: 100%;
  z-index: 30;
  background: linear-gradient(180deg, #0f172a 0%, #111827 100%);
  color: #e2e8f0;
  box-shadow: 18px 0 40px rgba(15, 23, 42, 0.35);
  display: flex;
  flex-direction: column;
  padding: 18px;
  gap: 16px;
}

.drawer-head {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
}

.drawer-title {
  font-size: 20px;
  font-weight: bold;
  color: #ffffff;
}

.drawer-copy {
  font-size: 13px;
  color: #cbd5e1;
  line-height: 1.5;
}

.drawer-nav {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
}

.nav-item {
	align-self: stretch;
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	min-height: 62px;
  text-align: left;
  padding: 12px 14px;
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  background-color: rgba(255, 255, 255, 0.04);
  color: #e2e8f0;
  cursor: pointer;
}
.nav-item:hover {
  background-color: rgba(255, 255, 255, 0.08);
}

.nav-item-active {
  background-color: #1d4ed8;
  border-color: #1d4ed8;
}

.nav-label {
  display: block;
  font-size: 14px;
  font-weight: bold;
	line-height: 1.35;
}

.nav-note {
  display: block;
  margin-top: 4px;
  font-size: 12px;
	line-height: 1.35;
  color: #cbd5e1;
}

.drawer-footer {
  margin-top: auto;
  border-top: 1px solid rgba(148, 163, 184, 0.18);
  padding-top: 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.drawer-chip {
  align-self: flex-start;
  border-radius: 999px;
  background-color: rgba(96, 165, 250, 0.16);
  color: #bfdbfe;
  padding: 4px 8px;
  font-size: 11px;
  font-weight: bold;
  transform: rotate(-8deg);
}
`

type drawerSection struct {
	Title   string
	Copy    string
	MetricA string
	MetricB string
	MetricC string
	NoteA   string
	NoteB   string
	NoteC   string
}

var drawerSections = map[string]drawerSection{
	"inbox": {
		Title:   "Inbox",
		Copy:    "A compact side drawer with a fixed overlay. Click the dark veil to close it or jump to another section.",
		MetricA: "12",
		MetricB: "3",
		MetricC: "97%",
		NoteA:   "New items today",
		NoteB:   "Pinned threads",
		NoteC:   "Delivery rate",
	},
	"projects": {
		Title:   "Projects",
		Copy:    "The drawer shows that fixed overlays, stacking order, and active navigation states all behave together in a snapshot render.",
		MetricA: "8",
		MetricB: "4",
		MetricC: "21",
		NoteA:   "Active boards",
		NoteB:   "Owners",
		NoteC:   "Open tasks",
	},
	"settings": {
		Title:   "Settings",
		Copy:    "This view keeps the drawer open state deterministic and validates that content behind the overlay remains untouched until the next snapshot.",
		MetricA: "6",
		MetricB: "2",
		MetricC: "1",
		NoteA:   "Profile sections",
		NoteB:   "Security checks",
		NoteC:   "Saved themes",
	},
}

type drawerState struct {
	mu     sync.Mutex
	open   bool
	active string
}

func newDrawerState() *drawerState {
	return &drawerState{active: "inbox"}
}

func (s *drawerState) onEvent(eventName string, _ map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch strings.TrimSpace(eventName) {
	case "drawer.toggle":
		s.open = !s.open
	case "drawer.close":
		s.open = false
	case "drawer.section.inbox":
		s.active = "inbox"
		s.open = false
	case "drawer.section.projects":
		s.active = "projects"
		s.open = false
	case "drawer.section.settings":
		s.active = "settings"
		s.open = false
	}
	return nil
}

func (s *drawerState) snapshot() (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.open, s.active
}

func navButton(label, note, eventName, active, current string) *giocss.Node {
	btn := components.Button("", "nav-item")
	btn.SetProp("event", eventName)
	btn.AddChild(components.Span(label, "nav-label"))
	btn.AddChild(components.Span(note, "nav-note"))
	if active == current {
		btn.AddClass("nav-item-active")
	}
	return btn
}

func statBox(label, value, note string) *giocss.Node {
	box := components.Column(
		components.Text(label, "stat-label"),
		components.Text(value, "stat-value"),
		components.Text(note, "stat-note"),
	)
	box.AddClass("stat")
	return box
}

func buildDrawer(active string) *giocss.Node {
	section := drawerSections[active]
	panel := components.Column()
	panel.AddClass("drawer-panel")

	head := components.Row(
		components.Heading(2, "Navigation", "drawer-title"),
		components.Button("Close", "close-btn"),
	)
	head.AddClass("drawer-head")
	head.Children[1].SetProp("event", "drawer.close")
	panel.AddChild(head)
	panel.AddChild(components.Text(section.Copy, "drawer-copy"))

	nav := components.Column(
		navButton("Inbox", "Messages and updates", "drawer.section.inbox", active, "inbox"),
		navButton("Projects", "Boards and milestones", "drawer.section.projects", active, "projects"),
		navButton("Settings", "Preferences and security", "drawer.section.settings", active, "settings"),
	)
	nav.AddClass("drawer-nav")
	panel.AddChild(nav)

	footer := components.Column(
		components.Badge("Fixed + overlay", "drawer-chip"),
		components.Text("The panel stays pinned to the viewport while the overlay darkens the rest of the scene.", "drawer-copy"),
	)
	footer.AddClass("drawer-footer")
	panel.AddChild(footer)

	return panel
}

func buildUI(open bool, active string) *giocss.Node {
	root := giocss.NewNode("body")

	shell := components.Column()
	shell.AddClass("shell")

	hero := components.Column(
		components.Badge("Sample 16", "eyebrow"),
		components.Heading(1, "Side Drawer", "title"),
		components.Text("This example exercises fixed positioning, z-index layering, and overlay click handling with real app state.", "subtitle"),
	)
	hero.AddClass("hero")
	menu := components.Button("Open drawer", "menu-btn")
	menu.SetProp("event", "drawer.toggle")
	hero.AddChild(menu)
	stateLabel := "Drawer state: closed · active section: " + active
	if open {
		stateLabel = "Drawer state: open · active section: " + active
	}
	hero.AddChild(components.Text(stateLabel, "subtitle"))
	shell.AddChild(hero)

	content := components.Column(
		components.Column(
			components.Heading(2, drawerSections[active].Title, "card-title"),
			components.Text("The main surface stays static while the drawer animates conceptually through state changes and stacking order.", "card-copy"),
		),
	)
	content.AddClass("content-card")
	stats := components.Row(
		statBox("Primary", drawerSections[active].MetricA, drawerSections[active].NoteA),
		statBox("Secondary", drawerSections[active].MetricB, drawerSections[active].NoteB),
		statBox("Coverage", drawerSections[active].MetricC, drawerSections[active].NoteC),
	)
	stats.AddClass("stats")
	content.AddChild(stats)
	content.AddChild(components.Text("Try the menu button, click the shaded backdrop, and switch sections from inside the drawer.", "card-copy"))
	shell.AddChild(content)

	root.AddChild(shell)

	if open {
		overlay := components.Button("", "drawer-overlay")
		overlay.SetProp("event", "drawer.close")
		root.AddChild(overlay)
		root.AddChild(buildDrawer(active))
	}

	return root
}

func main() {
	ss := giocss.NewStyleSheet()
	ss.ParseCSSText(css)
	app := newDrawerState()

	rt := giocss.NewWindowRuntime(
		giocss.WindowOptions{Title: "Sample 16 - Side Drawer", Width: 980, Height: 620},
		giocss.WindowRuntimeHooks{
			DispatchEvent: app.onEvent,
			Snapshot: func(size image.Point) giocss.WindowRuntimeSnapshot {
				open, active := app.snapshot()
				root := buildUI(open, active)
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
