package main

import (
	"image"
	"strconv"
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
  background: linear-gradient(160deg, #f8fafc 0%, #ecfeff 100%);
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 24px;
}

.shell {
  width: 100%;
  max-width: 860px;
  background-color: #ffffff;
  border-radius: 14px;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.12);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 18px;
  border-bottom: 1px solid #e2e8f0;
}

.title {
  font-size: 24px;
  font-weight: bold;
  color: #0f172a;
	margin: 0;
	line-height: 1.1;
}

.subtitle {
  font-size: 13px;
  color: #64748b;
}

.accordion {
  display: flex;
  flex-direction: column;
}

.item {
  display: flex;
  flex-direction: column;
  border-top: 1px solid #f1f5f9;
}

.item-head {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
	padding: 0 16px;
	min-height: 56px;
}

.item-btn {
	display: flex;
	align-items: center;
	justify-content: flex-start;
	width: 100%;
	min-width: 0;
  background-color: transparent;
  border: 0;
  color: #0f172a;
  font-size: 15px;
  font-weight: bold;
  text-align: left;
  cursor: pointer;
	padding: 0;
	line-height: 1.2;
}
.item-btn:hover {
  color: #2563eb;
}

.chev {
	display: flex;
	align-items: center;
	justify-content: center;
  font-size: 16px;
  color: #64748b;
  transform: rotate(0deg);
	flex-shrink: 0;
	min-width: 16px;
}

.chev-open {
  transform: rotate(90deg);
  color: #1d4ed8;
}

.panel {
  padding: 0 18px 16px 18px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.panel-text {
  font-size: 14px;
  color: #475569;
  line-height: 1.45;
}

.tag-row {
  display: flex;
  flex-direction: row;
  gap: 8px;
	flex-wrap: wrap;
	align-items: center;
	margin-top: 6px;
}

.badge {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	line-height: 1;
	white-space: nowrap;
}

.tag {
  background-color: #e2e8f0;
  color: #334155;
  border-radius: 999px;
  padding: 4px 8px;
  font-size: 11px;
  font-weight: bold;
}
`

type section struct {
	Title string
	Body  string
	Tags  []string
}

var sections = []section{
	{
		Title: "Rendering pipeline",
		Body:  "The renderer resolves styles, computes layout, and draws the tree in one frame snapshot. Open sections update state and trigger a fresh snapshot.",
		Tags:  []string{"layout", "paint", "snapshot"},
	},
	{
		Title: "Event dispatch",
		Body:  "Each header button dispatches an event. The runtime mutates app state in DispatchEvent, then Snapshot rebuilds the node tree from the new state.",
		Tags:  []string{"events", "state", "runtime"},
	},
	{
		Title: "Transform support",
		Body:  "The chevron uses transform: rotate(90deg) when open. This sample validates transform usage in small UI interactions.",
		Tags:  []string{"transform", "rotate", "interaction"},
	},
}

type appState struct {
	mu   sync.Mutex
	open int
}

func newAppState() *appState {
	return &appState{open: 0}
}

func (s *appState) onEvent(eventName string, _ map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	eventName = strings.TrimSpace(eventName)
	if !strings.HasPrefix(eventName, "accordion.toggle.") {
		return nil
	}
	idxRaw := strings.TrimPrefix(eventName, "accordion.toggle.")
	idx, err := strconv.Atoi(idxRaw)
	if err != nil || idx < 0 || idx >= len(sections) {
		return nil
	}
	if s.open == idx {
		s.open = -1
	} else {
		s.open = idx
	}
	return nil
}

func (s *appState) snapshot() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.open
}

func buildSection(index int, isOpen bool, sec section) *giocss.Node {
	item := components.Column()
	item.AddClass("item")

	head := components.Row()
	head.AddClass("item-head")

	btn := components.Button(sec.Title, "item-btn")
	btn.SetProp("event", "accordion.toggle."+strconv.Itoa(index))
	head.AddChild(btn)

	chev := components.Text(">", "chev")
	if isOpen {
		chev.AddClass("chev-open")
	}
	head.AddChild(chev)
	item.AddChild(head)

	if isOpen {
		panel := components.Column(
			components.Text(sec.Body, "panel-text"),
		)
		panel.AddClass("panel")
		tagRow := components.Row()
		tagRow.AddClass("tag-row")
		for _, t := range sec.Tags {
			tagRow.AddChild(components.Badge(strings.ToUpper(t), "tag"))
		}
		panel.AddChild(tagRow)
		item.AddChild(panel)
	}

	return item
}

func buildUI(open int) *giocss.Node {
	root := giocss.NewNode("body")

	shell := components.Column(
		components.Column(
			components.Heading(1, "Sample 14 - Accordion", "title"),
			components.Text("Event-driven expandable panels with transform rotation.", "subtitle"),
		),
	)
	shell.AddClass("shell")
	shell.Children[0].AddClass("header")

	acc := components.Column()
	acc.AddClass("accordion")
	for i, sec := range sections {
		acc.AddChild(buildSection(i, i == open, sec))
	}
	shell.AddChild(acc)

	root.AddChild(shell)
	return root
}

func main() {
	ss := giocss.NewStyleSheet()
	ss.ParseCSSText(css)
	app := newAppState()

	rt := giocss.NewWindowRuntime(
		giocss.WindowOptions{Title: "Sample 14 - Accordion", Width: 980, Height: 580},
		giocss.WindowRuntimeHooks{
			DispatchEvent: app.onEvent,
			Snapshot: func(size image.Point) giocss.WindowRuntimeSnapshot {
				root := buildUI(app.snapshot())
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
