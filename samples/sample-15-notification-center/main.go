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
  background-color: #f8fafc;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 24px;
}

.shell {
  width: 100%;
  max-width: 920px;
  background-color: #ffffff;
  border-radius: 12px;
  box-shadow: 0 12px 26px rgba(15, 23, 42, 0.1);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.header {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  padding: 16px 18px;
  border-bottom: 1px solid #e2e8f0;
}

.title {
  font-size: 22px;
  font-weight: bold;
  color: #0f172a;
	margin: 0;
	line-height: 1.1;
	flex: 1 1 auto;
	min-width: 0;
}

.stats {
  font-size: 12px;
  color: #64748b;
	flex-shrink: 0;
}

.actions {
  display: flex;
  flex-direction: row;
  gap: 8px;
  padding: 12px 18px;
  border-bottom: 1px solid #e2e8f0;
}

.action-btn {
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid #cbd5e1;
  background-color: #f8fafc;
  color: #334155;
  cursor: pointer;
  font-size: 12px;
  font-weight: bold;
}
.action-btn:hover {
  background-color: #eef2ff;
}

.list {
  display: flex;
  flex-direction: column;
  padding: 12px;
  gap: 10px;
}

.card {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background-color: #ffffff;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.card-unread {
  border-color: #93c5fd;
  box-shadow: inset 3px 0 0 #2563eb;
}

.row-top {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
	align-items: flex-start;
  gap: 10px;
}

.card-title {
  font-size: 14px;
  color: #0f172a;
  font-weight: bold;
	line-height: 1.2;
}

.card-body {
  font-size: 13px;
  color: #475569;
  line-height: 1.4;
}

.level {
	display: inline-flex;
	align-items: center;
	justify-content: center;
  border-radius: 999px;
  padding: 4px 8px;
  font-size: 11px;
  font-weight: bold;
	line-height: 1;
	white-space: nowrap;
}

.level-info { background-color: #dbeafe; color: #1d4ed8; }
.level-warn { background-color: #fef3c7; color: #b45309; }
.level-error { background-color: #fee2e2; color: #b91c1c; }

.row-actions {
  display: flex;
  flex-direction: row;
  gap: 8px;
	flex-wrap: wrap;
}

.small-btn {
  padding: 6px 10px;
  border-radius: 8px;
  border: 1px solid #cbd5e1;
  background-color: #ffffff;
  color: #334155;
  cursor: pointer;
  font-size: 12px;
	min-height: 32px;
}
.small-btn:hover {
  background-color: #f8fafc;
}

.badge {
	display: inline-flex;
	align-items: center;
	justify-content: center;
}

.empty {
  padding: 22px;
  font-size: 13px;
  color: #64748b;
}
`

type notification struct {
	ID    int
	Title string
	Body  string
	Level string
	Read  bool
}

func seedNotifications() []notification {
	return []notification{
		{ID: 1, Title: "Build succeeded", Body: "The main branch build completed in 2m 13s.", Level: "info", Read: false},
		{ID: 2, Title: "Usage spike", Body: "API traffic increased by 34% in the last hour.", Level: "warn", Read: false},
		{ID: 3, Title: "Payment retry failed", Body: "A payment provider rejected a retry attempt.", Level: "error", Read: true},
		{ID: 4, Title: "Release note ready", Body: "Version 0.2.0 release notes were generated.", Level: "info", Read: true},
	}
}

type appState struct {
	mu     sync.Mutex
	items  []notification
	nextID int
}

func newAppState() *appState {
	items := seedNotifications()
	maxID := 0
	for _, item := range items {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return &appState{items: items, nextID: maxID + 1}
}

func (s *appState) onEvent(eventName string, _ map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	eventName = strings.TrimSpace(eventName)
	switch {
	case eventName == "notif.markall":
		for i := range s.items {
			s.items[i].Read = true
		}
	case eventName == "notif.reset":
		s.items = seedNotifications()
	case strings.HasPrefix(eventName, "notif.mark."):
		id, ok := parseID(eventName, "notif.mark.")
		if ok {
			for i := range s.items {
				if s.items[i].ID == id {
					s.items[i].Read = !s.items[i].Read
					break
				}
			}
		}
	case strings.HasPrefix(eventName, "notif.remove."):
		id, ok := parseID(eventName, "notif.remove.")
		if ok {
			filtered := make([]notification, 0, len(s.items))
			for _, item := range s.items {
				if item.ID != id {
					filtered = append(filtered, item)
				}
			}
			s.items = filtered
		}
	}

	return nil
}

func parseID(eventName string, prefix string) (int, bool) {
	idRaw := strings.TrimPrefix(eventName, prefix)
	id, err := strconv.Atoi(idRaw)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (s *appState) snapshot() []notification {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]notification, len(s.items))
	copy(out, s.items)
	return out
}

func unreadCount(items []notification) int {
	count := 0
	for _, item := range items {
		if !item.Read {
			count++
		}
	}
	return count
}

func levelClass(level string) string {
	switch level {
	case "warn":
		return "level-warn"
	case "error":
		return "level-error"
	default:
		return "level-info"
	}
}

func cardNode(item notification) *giocss.Node {
	card := components.Column()
	card.AddClass("card")
	if !item.Read {
		card.AddClass("card-unread")
	}

	top := components.Row(
		components.Text(item.Title, "card-title"),
		components.Badge(strings.ToUpper(item.Level), "level", levelClass(item.Level)),
	)
	top.AddClass("row-top")
	card.AddChild(top)
	card.AddChild(components.Text(item.Body, "card-body"))

	actions := components.Row(
		components.Button("Toggle Read", "small-btn"),
		components.Button("Dismiss", "small-btn"),
	)
	actions.AddClass("row-actions")
	actions.Children[0].SetProp("event", "notif.mark."+strconv.Itoa(item.ID))
	actions.Children[1].SetProp("event", "notif.remove."+strconv.Itoa(item.ID))
	card.AddChild(actions)

	return card
}

func buildUI(items []notification) *giocss.Node {
	root := giocss.NewNode("body")
	shell := components.Column()
	shell.AddClass("shell")

	head := components.Row(
		components.Heading(1, "Sample 15 - Notification Center", "title"),
		components.Text(strconv.Itoa(unreadCount(items))+" unread", "stats"),
	)
	head.AddClass("header")
	shell.AddChild(head)

	actions := components.Row(
		components.Button("Mark all as read", "action-btn"),
		components.Button("Reset", "action-btn"),
	)
	actions.AddClass("actions")
	actions.Children[0].SetProp("event", "notif.markall")
	actions.Children[1].SetProp("event", "notif.reset")
	shell.AddChild(actions)

	list := components.Column()
	list.AddClass("list")
	if len(items) == 0 {
		list.AddChild(components.Text("No notifications available.", "empty"))
	} else {
		for _, item := range items {
			list.AddChild(cardNode(item))
		}
	}
	shell.AddChild(list)

	root.AddChild(shell)
	return root
}

func main() {
	ss := giocss.NewStyleSheet()
	ss.ParseCSSText(css)
	app := newAppState()

	rt := giocss.NewWindowRuntime(
		giocss.WindowOptions{Title: "Sample 15 - Notification Center", Width: 980, Height: 640},
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
