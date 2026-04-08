package main

import (
	"image"
	"sort"
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
  padding: 24px;
}

.shell {
  width: 100%;
  max-width: 900px;
  background-color: #ffffff;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  box-shadow: 0 16px 36px rgba(15, 23, 42, 0.1);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.header {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 18px;
  border-bottom: 1px solid #e2e8f0;
}

.eyebrow {
  align-self: flex-start;
  background-color: #ede9fe;
  color: #6d28d9;
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
  line-height: 1.5;
}

.search-shell {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 18px;
}

.search-row {
  display: flex;
  flex-direction: row;
  gap: 10px;
  align-items: center;
}

.search-input {
  flex: 1;
  background-color: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 12px;
  padding: 12px 14px;
  font-size: 14px;
  color: #0f172a;
}

.clear-btn {
  padding: 12px 14px;
  border-radius: 12px;
  border: 1px solid #cbd5e1;
  background-color: #f8fafc;
  color: #334155;
  cursor: pointer;
  font-size: 13px;
  font-weight: bold;
}
.clear-btn:hover {
  background-color: #eef2ff;
}

.dropdown {
	margin-top: 2px;
  border: 1px solid #cbd5e1;
  border-radius: 14px;
  background-color: #ffffff;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.12);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.suggestion {
	display: flex;
	flex-direction: column;
	width: 100%;
  text-align: left;
  border: 0;
  border-bottom: 1px solid #e2e8f0;
  background-color: #ffffff;
  padding: 12px 14px;
  cursor: pointer;
}
.suggestion:hover {
  background-color: #f8fafc;
}
.suggestion:last-child {
  border-bottom: 0;
}

.suggestion-active {
  background-color: #eff6ff;
}

.suggestion-name {
  display: block;
  font-size: 14px;
  font-weight: bold;
  color: #0f172a;
}

.suggestion-meta {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: #64748b;
}

.empty {
  padding: 18px;
  font-size: 13px;
  color: #64748b;
}

.details {
  margin: 18px;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  background-color: #f8fafc;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.details-title {
  font-size: 18px;
  font-weight: bold;
  color: #0f172a;
}

.details-copy {
  font-size: 14px;
  color: #475569;
  line-height: 1.5;
}

.tag-row {
  display: flex;
  flex-direction: row;
  gap: 8px;
  flex-wrap: wrap;
}

.tag {
  background-color: #e2e8f0;
  color: #334155;
  border-radius: 999px;
  padding: 4px 8px;
  font-size: 11px;
  font-weight: bold;
}

.selected-chip {
  align-self: flex-start;
  background-color: #dbeafe;
  color: #1d4ed8;
  border-radius: 999px;
  padding: 4px 10px;
  font-size: 11px;
  font-weight: bold;
  transform: rotate(-6deg);
}
`

type commandItem struct {
	ID   string
	Name string
	Kind string
	Copy string
	Tags []string
	Tone string
}

var catalog = []commandItem{
	{ID: "modal", Name: "Open modal", Kind: "ui", Copy: "Shows an overlay dialog with keyboard-safe focus flow.", Tags: []string{"overlay", "focus", "dialog"}, Tone: "blue"},
	{ID: "drawer", Name: "Open drawer", Kind: "layout", Copy: "Slides in a fixed side panel with a dimmed backdrop.", Tags: []string{"fixed", "overlay", "stacking"}, Tone: "violet"},
	{ID: "tabs", Name: "Switch tabs", Kind: "state", Copy: "Changes the active surface without leaving the page.", Tags: []string{"state", "active", "snapshot"}, Tone: "emerald"},
	{ID: "table", Name: "Filter table", Kind: "data", Copy: "Narrow rows by text and sort by clicking headers.", Tags: []string{"input", "sort", "filter"}, Tone: "amber"},
	{ID: "notify", Name: "Read notifications", Kind: "feed", Copy: "Marks items as read or removes them from the list.", Tags: []string{"badge", "list", "interaction"}, Tone: "rose"},
	{ID: "search", Name: "Search autocomplete", Kind: "input", Copy: "Matches commands as you type and can be picked from the list.", Tags: []string{"input", "dropdown", "selection"}, Tone: "cyan"},
}

type autocompleteState struct {
	mu       sync.Mutex
	query    string
	selected string
}

func newAutocompleteState() *autocompleteState {
	return &autocompleteState{query: "search"}
}

func (s *autocompleteState) onEvent(eventName string, payload map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch {
	case strings.TrimSpace(eventName) == "search.clear":
		s.query = ""
		s.selected = ""
	case strings.TrimSpace(eventName) == "search.input":
		if value, ok := payload["value"].(string); ok {
			s.query = value
			if s.selected != "" {
				keepSelected := false
				normalizedValue := strings.ToLower(strings.TrimSpace(value))
				for _, item := range catalog {
					if item.ID != s.selected {
						continue
					}
					if strings.ToLower(strings.TrimSpace(item.Name)) == normalizedValue {
						keepSelected = true
					}
					break
				}
				if !keepSelected {
					s.selected = ""
				}
			}
		}
	case strings.HasPrefix(eventName, "search.select."):
		id := strings.TrimPrefix(strings.TrimSpace(eventName), "search.select.")
		for _, item := range catalog {
			if item.ID == id {
				s.query = item.Name
				s.selected = item.ID
				break
			}
		}
	}
	return nil
}

func (s *autocompleteState) snapshot() (string, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.query, s.selected
}

func matches(item commandItem, query string) int {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return 1
	}
	name := strings.ToLower(item.Name)
	kind := strings.ToLower(item.Kind)
	copy := strings.ToLower(item.Copy)
	score := 0

	if strings.HasPrefix(name, query) {
		score = max(score, 6)
	} else if strings.Contains(name, query) {
		score = max(score, 4)
	}
	if strings.HasPrefix(kind, query) || strings.Contains(kind, query) {
		score = max(score, 3)
	}
	for _, tag := range item.Tags {
		lowTag := strings.ToLower(tag)
		if strings.HasPrefix(lowTag, query) {
			score = max(score, 3)
		} else if strings.Contains(lowTag, query) {
			score = max(score, 2)
		}
	}
	if len(query) >= 3 && strings.Contains(copy, query) {
		score = max(score, 1)
	}
	return score
}

func filterCatalog(query string) []commandItem {
	type scored struct {
		item  commandItem
		score int
	}
	items := make([]scored, 0, len(catalog))
	for _, item := range catalog {
		if score := matches(item, query); score > 0 {
			items = append(items, scored{item: item, score: score})
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].score != items[j].score {
			return items[i].score > items[j].score
		}
		return items[i].item.Name < items[j].item.Name
	})
	out := make([]commandItem, 0, len(items))
	for _, item := range items {
		out = append(out, item.item)
		if len(out) >= 8 {
			break
		}
	}
	return out
}

func toneClass(tone string) string {
	switch tone {
	case "violet":
		return "tag"
	case "emerald":
		return "tag"
	case "amber":
		return "tag"
	case "rose":
		return "tag"
	case "cyan":
		return "tag"
	default:
		return "tag"
	}
}

func suggestionButton(item commandItem, active bool) *giocss.Node {
	btn := components.Button("", "suggestion")
	btn.SetProp("event", "search.select."+item.ID)
	if active {
		btn.AddClass("suggestion-active")
	}
	btn.AddChild(components.Span(item.Name, "suggestion-name"))
	btn.AddChild(components.Span(item.Kind+" · "+item.Copy, "suggestion-meta"))
	return btn
}

func tagRow(tags []string) *giocss.Node {
	row := components.Row()
	row.AddClass("tag-row")
	for _, tag := range tags {
		row.AddChild(components.Badge(strings.ToUpper(tag), "tag"))
	}
	return row
}

func buildUI(query string, selected string) *giocss.Node {
	root := giocss.NewNode("body")

	shell := components.Column()
	shell.AddClass("shell")

	head := components.Column(
		components.Badge("Sample 17", "eyebrow"),
		components.Heading(1, "Search Autocomplete", "title"),
		components.Text("Typing filters a live suggestion list; choosing a row rewrites the input and updates the detail panel.", "subtitle"),
	)
	head.AddClass("header")
	shell.AddChild(head)

	searchShell := components.Column()
	searchShell.AddClass("search-shell")
	row := components.Row()
	row.AddClass("search-row")
	input := components.Input("search", "Search commands, layouts or state", "text")
	input.AddClass("search-input")
	input.SetProp("value", query)
	input.SetProp("oninput", "search.input")
	row.AddChild(input)
	clear := components.Button("Clear", "clear-btn")
	clear.SetProp("event", "search.clear")
	row.AddChild(clear)
	searchShell.AddChild(row)

	results := filterCatalog(query)
	if len(results) > 0 {
		dropdown := components.Column()
		dropdown.AddClass("dropdown")
		for _, item := range results {
			dropdown.AddChild(suggestionButton(item, selected == item.ID))
		}
		searchShell.AddChild(dropdown)
	} else {
		searchShell.AddChild(components.Text("No matches. Try a different term or clear the input.", "empty"))
	}
	shell.AddChild(searchShell)

	active := catalog[0]
	for _, item := range catalog {
		if item.ID == selected {
			active = item
			break
		}
	}

	details := components.Column(
		components.Span(active.Name+" · "+active.Kind, "selected-chip"),
		components.Heading(2, active.Name, "details-title"),
		components.Text(active.Copy, "details-copy"),
		tagRow(active.Tags),
	)
	details.AddClass("details")
	shell.AddChild(details)

	root.AddChild(shell)
	return root
}

func main() {
	ss := giocss.NewStyleSheet()
	ss.ParseCSSText(css)
	app := newAutocompleteState()

	rt := giocss.NewWindowRuntime(
		giocss.WindowOptions{Title: "Sample 17 - Search Autocomplete", Width: 960, Height: 640},
		giocss.WindowRuntimeHooks{
			DispatchEvent: app.onEvent,
			Snapshot: func(size image.Point) giocss.WindowRuntimeSnapshot {
				query, selected := app.snapshot()
				root := buildUI(query, selected)
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
