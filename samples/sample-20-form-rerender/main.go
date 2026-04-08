package main

import (
	"fmt"
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
	padding: 24px;
	background: linear-gradient(180deg, #f7fafc 0%, #eef4fb 100%);
	display: flex;
	flex-direction: column;
	gap: 18px;
	font-family: Segoe UI, Tahoma, sans-serif;
}

.hero {
	padding: 18px;
	border-radius: 16px;
	border: 1px solid #d8e4ef;
	background: #ffffff;
	display: flex;
	flex-direction: column;
	gap: 8px;
}

.hero-title {
	font-size: 28px;
	font-weight: 800;
	color: #16324d;
}

.hero-copy {
	font-size: 14px;
	line-height: 1.45;
	color: #4d6a86;
}

.layout {
	display: flex;
	flex-direction: row;
	gap: 18px;
	flex-wrap: wrap;
}

.panel {
	flex: 1;
	min-width: 320px;
	padding: 16px;
	border-radius: 16px;
	border: 1px solid #d8e4ef;
	background: #ffffff;
	display: flex;
	flex-direction: column;
	gap: 10px;
}

.panel-title {
	font-size: 18px;
	font-weight: 800;
	color: #173652;
}

.panel-copy {
	font-size: 13px;
	line-height: 1.45;
	color: #5a7691;
}

.field-label {
	font-size: 12px;
	font-weight: 700;
	color: #315678;
}

.field-input {
	padding: 8px 10px;
	border-radius: 10px;
	border: 1px solid #bfd4e6;
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

option {
	padding: 8px 10px;
	background: #ffffff;
	color: #17344f;
}

option:hover {
	background: #e8f3fb;
	color: #123f68;
}

.submit-btn {
	width: 100%;
	max-width: 100%;
	box-sizing: border-box;
	padding: 10px 14px;
	border-radius: 10px;
	background: #123f68;
	color: #f3fbff;
	font-size: 13px;
	font-weight: 700;
	border: 1px solid #0f3658;
	text-align: center;
	cursor: pointer;
}

.submit-btn:hover {
	background: #16507f;
}

.status-chip {
	padding: 10px 12px;
	border-radius: 12px;
	background: #eef6fd;
	border: 1px solid #d0e4f4;
	font-size: 12px;
	line-height: 1.4;
	color: #315678;
}

.summary-box {
	padding: 12px;
	border-radius: 12px;
	background: #f9fcff;
	border: 1px solid #dbe8f3;
	display: flex;
	flex-direction: column;
	gap: 4px;
	font-size: 12px;
	line-height: 1.45;
	color: #47637d;
}

.summary-line {
	font-size: 12px;
	line-height: 1.45;
	color: #47637d;
}
`

type appState struct {
	mu            sync.Mutex
	renderCount   int
	status        string
	lastSubmitted string
}

func newAppState() *appState {
	return &appState{status: "Start typing. Every oninput/onchange event rerenders this view without controlling the fields."}
}

func (s *appState) onEvent(eventName string, payload map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.renderCount++

	switch strings.TrimSpace(eventName) {
	case "demo.form.live":
		value, _ := payload["value"].(string)
		value = strings.TrimSpace(value)
		if len(value) > 24 {
			value = value[:24] + "..."
		}
		if value == "" {
			s.status = fmt.Sprintf("Rerender #%d: field changed, current payload is empty.", s.renderCount)
		} else {
			s.status = fmt.Sprintf("Rerender #%d: latest value snapshot = %q", s.renderCount, value)
		}
	case "demo.form.submit":
		values, _ := payload["values"].(map[string]any)
		fullName, _ := values["full-name"].(string)
		email, _ := values["email"].(string)
		role, _ := values["role"].(string)
		s.status = fmt.Sprintf("Submit received after %d runtime updates.", s.renderCount)
		s.lastSubmitted = fmt.Sprintf("full-name=%q\nemail=%q\nrole=%q", fullName, email, role)
	}
	return nil
}

func (s *appState) snapshot() (string, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status, s.lastSubmitted
}

func buildUI(status string, submitted string) *giocss.Node {
	root := giocss.NewNode("body")
	root.AddChild(components.Column(
		components.Heading(1, "Form rerender stability", "hero-title"),
		components.Span("These controls are intentionally uncontrolled. Typing triggers state updates and a full snapshot rebuild, but the editor text should remain intact now.", "hero-copy"),
	))
	root.Children[0].AddClass("hero")

	nameInput := components.Input("full-name", "Ada Lovelace", "text")
	nameInput.AddClass("field-input")
	nameInput.SetProp("required", "true")
	nameInput.SetProp("oninput", "demo.form.live")

	emailInput := components.Input("email", "ada@example.com", "email")
	emailInput.AddClass("field-input")
	emailInput.SetProp("required", "true")
	emailInput.SetProp("oninput", "demo.form.live")

	roleSelect := components.Select("role", []*giocss.Node{
		components.Option("Choose a role", "", true, false),
		components.Option("Reader", "reader", false, false),
		components.Option("Editor", "editor", false, false),
		components.Option("Owner", "owner", false, false),
	})
	roleSelect.AddClass("field-input")
	roleSelect.SetProp("required", "true")
	roleSelect.SetProp("onchange", "demo.form.live")
	roleSelect.SetProp("placeholder", "Choose a role")

	form := components.Form("rerender-demo", []string{"panel"},
		components.Heading(2, "Live form", "panel-title"),
		components.Span("Type in the fields and keep an eye on the status card. Before the fix, these rerenders cleared the input on every keystroke.", "panel-copy"),
		components.Fieldset("Profile", []string{},
			components.Label("Full name", "field-label"),
			nameInput,
			components.Label("Email", "field-label"),
			emailInput,
			components.Label("Role", "field-label"),
			roleSelect,
		),
		components.SubmitButton("Submit snapshot", "", "submit-btn"),
	)
	form.SetProp("onsubmit", "demo.form.submit")

	submittedText := submitted
	if strings.TrimSpace(submittedText) == "" {
		submittedText = "No submit yet. Fill the form, trigger rerenders while typing, then submit to inspect the collected values."
	}

	sidebar := components.Column(
		components.Heading(2, "Runtime status", "panel-title"),
		components.Span(status, "status-chip"),
		components.Heading(3, "Last submitted payload", "panel-title"),
		buildSummaryBox(submittedText),
	)
	sidebar.AddClass("panel")

	root.AddChild(components.Row(form, sidebar))
	root.Children[1].AddClass("layout")
	return root
}

func buildSummaryBox(content string) *giocss.Node {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	children := make([]*giocss.Node, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			line = " "
		}
		children = append(children, components.Span(line, "summary-line"))
	}
	return components.Div([]string{"summary-box"}, children...)
}

func main() {
	ss := giocss.NewStyleSheet()
	ss.ParseCSSText(css)
	app := newAppState()

	rt := giocss.NewWindowRuntime(
		giocss.WindowOptions{Title: "Sample 20 - Form Rerender", Width: 1120, Height: 700},
		giocss.WindowRuntimeHooks{
			DispatchEvent: app.onEvent,
			Snapshot: func(size image.Point) giocss.WindowRuntimeSnapshot {
				status, submitted := app.snapshot()
				root := buildUI(status, submitted)
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
