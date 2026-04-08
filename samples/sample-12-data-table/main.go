package main

import (
	"image"
	"sort"
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

.table-shell {
  width: 100%;
  max-width: 920px;
  background-color: #ffffff;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  padding: 18px;
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.1);
  gap: 12px;
}

.title {
  font-size: 24px;
  font-weight: bold;
  color: #0f172a;
}

.subtitle {
  font-size: 13px;
  color: #64748b;
}

.toolbar {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 10px;
}

.filter-input {
  width: 280px;
  background-color: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 8px 10px;
  color: #0f172a;
}

.clear-btn {
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid #cbd5e1;
  background-color: #f8fafc;
  color: #334155;
  cursor: pointer;
}
.clear-btn:hover {
  background-color: #eef2ff;
}

.grid {
  display: flex;
  flex-direction: column;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  overflow: auto;
  max-height: 420px;
}

.row {
  display: flex;
  flex-direction: row;
  align-items: center;
  min-height: 42px;
}

.header {
  background-color: #0f172a;
}

.th-btn {
  flex: 1;
  text-align: left;
  background-color: transparent;
  color: #e2e8f0;
  border: 0;
  border-right: 1px solid #1e293b;
  padding: 11px 12px;
  font-size: 13px;
  font-weight: bold;
  cursor: pointer;
}
.th-btn:hover {
  background-color: #111f3d;
}
.th-btn:last-child {
  border-right: 0;
}

.cell {
  flex: 1;
  padding: 10px 12px;
  border-right: 1px solid #e2e8f0;
  font-size: 13px;
  color: #334155;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.cell:last-child {
  border-right: 0;
}

.data-row {
  border-top: 1px solid #f1f5f9;
}
.data-row:hover {
  background-color: #f8fafc;
}

.score-good { color: #166534; font-weight: bold; }
.score-mid { color: #b45309; font-weight: bold; }
.score-low { color: #991b1b; font-weight: bold; }

.empty {
  padding: 24px;
  font-size: 13px;
  color: #64748b;
}
`

type userRow struct {
	Name  string
	Email string
	Role  string
	Score int
}

var baseRows = []userRow{
	{Name: "Ana Ramirez", Email: "ana.ramirez@acme.test", Role: "Design", Score: 98},
	{Name: "Luis Ortega", Email: "luis.ortega@acme.test", Role: "Engineering", Score: 86},
	{Name: "Valeria Soto", Email: "valeria.soto@acme.test", Role: "Product", Score: 91},
	{Name: "Martin Ruiz", Email: "martin.ruiz@acme.test", Role: "Support", Score: 74},
	{Name: "Noah Diaz", Email: "noah.diaz@acme.test", Role: "Engineering", Score: 82},
	{Name: "Camila Vega", Email: "camila.vega@acme.test", Role: "Research", Score: 95},
	{Name: "Sofia Luna", Email: "sofia.luna@acme.test", Role: "Operations", Score: 68},
}

type appState struct {
	mu     sync.Mutex
	sortBy string
	asc    bool
	filter string
}

func newAppState() *appState {
	return &appState{sortBy: "name", asc: true}
}

func (s *appState) toggleSort(field string) {
	if s.sortBy == field {
		s.asc = !s.asc
		return
	}
	s.sortBy = field
	s.asc = true
}

func (s *appState) onEvent(eventName string, payload map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch strings.TrimSpace(eventName) {
	case "table.sort.name":
		s.toggleSort("name")
	case "table.sort.email":
		s.toggleSort("email")
	case "table.sort.role":
		s.toggleSort("role")
	case "table.sort.score":
		s.toggleSort("score")
	case "table.filter.input":
		if v, ok := payload["value"].(string); ok {
			s.filter = v
		}
	case "table.filter.clear":
		s.filter = ""
	}
	return nil
}

func (s *appState) snapshot() (string, bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sortBy, s.asc, s.filter
}

func filterRows(rows []userRow, query string) []userRow {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		out := make([]userRow, len(rows))
		copy(out, rows)
		return out
	}
	out := make([]userRow, 0, len(rows))
	for _, row := range rows {
		if strings.Contains(strings.ToLower(row.Name), query) ||
			strings.Contains(strings.ToLower(row.Email), query) ||
			strings.Contains(strings.ToLower(row.Role), query) {
			out = append(out, row)
		}
	}
	return out
}

func sortRows(rows []userRow, sortBy string, asc bool) {
	less := func(i, j int) bool { return false }
	switch sortBy {
	case "email":
		less = func(i, j int) bool { return rows[i].Email < rows[j].Email }
	case "role":
		less = func(i, j int) bool { return rows[i].Role < rows[j].Role }
	case "score":
		less = func(i, j int) bool { return rows[i].Score < rows[j].Score }
	default:
		less = func(i, j int) bool { return rows[i].Name < rows[j].Name }
	}
	if asc {
		sort.Slice(rows, less)
		return
	}
	sort.Slice(rows, func(i, j int) bool { return !less(i, j) })
}

func sortArrow(field string, active string, asc bool) string {
	if field != active {
		return ""
	}
	if asc {
		return " ↑"
	}
	return " ↓"
}

func scoreClass(score int) string {
	if score >= 90 {
		return "score-good"
	}
	if score >= 75 {
		return "score-mid"
	}
	return "score-low"
}

func cell(textValue string, classes ...string) *giocss.Node {
	n := giocss.NewNode("div")
	n.AddClass("cell")
	for _, c := range classes {
		n.AddClass(c)
	}
	n.Text = textValue
	return n
}

func headerButton(label, eventName string) *giocss.Node {
	btn := components.Button(label, "th-btn")
	btn.SetProp("event", eventName)
	return btn
}

func buildUI(sortBy string, asc bool, filter string) *giocss.Node {
	root := giocss.NewNode("body")

	shell := components.Column(
		components.Heading(1, "Sample 12 - Data Table", "title"),
		components.Text("Sorting + filtering with pure giocss events.", "subtitle"),
	)
	shell.AddClass("table-shell")

	toolbar := components.Row()
	toolbar.AddClass("toolbar")
	filterInput := components.Input("q", "Filter by name, email or role", "text")
	filterInput.AddClass("filter-input")
	filterInput.SetProp("value", filter)
	filterInput.SetProp("oninput", "table.filter.input")
	toolbar.AddChild(filterInput)

	clearBtn := components.Button("Clear", "clear-btn")
	clearBtn.SetProp("event", "table.filter.clear")
	toolbar.AddChild(clearBtn)
	shell.AddChild(toolbar)

	grid := components.Column()
	grid.AddClass("grid")

	head := components.Row(
		headerButton("Name"+sortArrow("name", sortBy, asc), "table.sort.name"),
		headerButton("Email"+sortArrow("email", sortBy, asc), "table.sort.email"),
		headerButton("Role"+sortArrow("role", sortBy, asc), "table.sort.role"),
		headerButton("Score"+sortArrow("score", sortBy, asc), "table.sort.score"),
	)
	head.AddClass("header")
	grid.AddChild(head)

	rows := filterRows(baseRows, filter)
	sortRows(rows, sortBy, asc)
	if len(rows) == 0 {
		empty := components.Text("No rows match your filter.", "empty")
		grid.AddChild(empty)
	} else {
		for _, row := range rows {
			line := components.Row(
				cell(row.Name),
				cell(row.Email),
				cell(row.Role),
				cell(strconv.Itoa(row.Score)+" pts", scoreClass(row.Score)),
			)
			line.AddClass("data-row")
			grid.AddChild(line)
		}
	}
	shell.AddChild(grid)
	root.AddChild(shell)

	return root
}

func main() {
	ss := giocss.NewStyleSheet()
	ss.ParseCSSText(css)
	app := newAppState()

	rt := giocss.NewWindowRuntime(
		giocss.WindowOptions{Title: "Sample 12 - Data Table", Width: 980, Height: 620},
		giocss.WindowRuntimeHooks{
			DispatchEvent: app.onEvent,
			Snapshot: func(size image.Point) giocss.WindowRuntimeSnapshot {
				sortBy, asc, filter := app.snapshot()
				root := buildUI(sortBy, asc, filter)
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
