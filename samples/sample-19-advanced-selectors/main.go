package main

import (
	"image"

	giocss "github.com/ArubikU/giocss"
	"github.com/ArubikU/giocss/components"
)

const css = `
body {
	width: 100%;
	height: 100%;
	box-sizing: border-box;
	padding: 24px;
	display: flex;
	flex-direction: column;
	gap: 18px;
	background: linear-gradient(180deg, #f5f8fc 0%, #ebf1f8 100%);
	font-family: Segoe UI, Tahoma, sans-serif;
}

.hero {
	padding: 20px;
	border-radius: 18px;
	border: 1px solid #d7e3ef;
	background: #ffffff;
	display: flex;
	flex-direction: column;
	gap: 8px;
}

.hero-title {
	font-size: 30px;
	font-weight: 800;
	color: #16324d;
}

.hero-copy {
	font-size: 14px;
	line-height: 1.45;
	color: #496581;
}

.layout {
	display: flex;
	flex-direction: row;
	gap: 18px;
	flex-wrap: wrap;
	align-items: flex-start;
}

.panel {
	flex: 1;
	min-width: 300px;
	padding: 16px;
	border-radius: 16px;
	border: 1px solid #d7e3ef;
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
	line-height: 1.4;
	color: #5a7691;
}

.selector-grid {
	display: flex;
	flex-direction: row;
	gap: 14px;
	flex-wrap: wrap;
}

.selector-grid > .selector-card {
	flex: 1;
	min-width: 220px;
	padding: 14px;
	border-radius: 14px;
	border: 1px solid #d6e5f1;
	background: #f9fcff;
	display: flex;
	flex-direction: column;
	gap: 8px;
}

.selector-card + .selector-card {
	border-color: #9bc3e6;
	box-shadow: 0 10px 26px rgba(31, 84, 129, 0.08);
}

.selector-card ~ .selector-card .card-note {
	color: #235c8b;
	font-weight: 700;
}

.card-title {
	font-size: 16px;
	font-weight: 800;
	color: #1a3d5c;
}

.card-note {
	font-size: 12px;
	line-height: 1.4;
	color: #70879d;
}

.swatch-list {
	display: flex;
	flex-direction: column;
	border-radius: 14px;
	overflow: hidden;
	border: 1px solid #d8e4ef;
}

.swatch-list > .swatch-row {
	display: flex;
	flex-direction: row;
	justify-content: space-between;
	align-items: center;
	padding: 9px 14px;
	gap: 12px;
}

.swatch-list > .swatch-row:nth-child(odd) {
	background: #f9fcff;
}

.swatch-list > .swatch-row:nth-child(even) {
	background: #eef5fb;
}

.swatch-list > .swatch-row:nth-child(3) {
	border-left: 4px solid #2f80c0;
}

.swatch-label {
	font-size: 13px;
	font-weight: 700;
	color: #264865;
}

.swatch-meta {
	font-size: 12px;
	color: #617b93;
}
`

func buildUI() *giocss.Node {
	root := giocss.NewNode("body")
	root.AddChild(components.Column(
		components.Heading(1, "Advanced selectors now live", "hero-title"),
		components.Span("This sample highlights descendant, child, adjacent sibling, general sibling and :nth-child matching without changing the authoring API.", "hero-copy"),
	))
	root.Children[0].AddClass("hero")

	left := components.Column(
		components.Heading(2, "Combinators", "panel-title"),
		components.Span("The first rule styles direct cards, the second targets every card after the first, and the sibling selector intensifies helper copy on later cards.", "panel-copy"),
		components.Div([]string{"selector-grid"},
			components.Div([]string{"selector-card"},
				components.Span("Direct child card", "card-title"),
				components.Span("Base card inside .selector-grid > .selector-card", "card-note"),
			),
			components.Div([]string{"selector-card"},
				components.Span("Adjacent sibling card", "card-title"),
				components.Span("Styled by .selector-card + .selector-card", "card-note"),
			),
			components.Div([]string{"selector-card"},
				components.Span("General sibling card", "card-title"),
				components.Span("Styled by .selector-card ~ .selector-card .card-note", "card-note"),
			),
		),
	)
	left.AddClass("panel")

	right := components.Column(
		components.Heading(2, ":nth-child", "panel-title"),
		components.Span("Odd/even rows are striped, and the third row gets an accent border to make structural matching obvious.", "panel-copy"),
		components.Div([]string{"swatch-list"},
			buildSwatchRow("01", "odd row"),
			buildSwatchRow("02", "even row"),
			buildSwatchRow("03", "explicit nth-child(3)"),
			buildSwatchRow("04", "even row"),
			buildSwatchRow("05", "odd row"),
		),
	)
	right.AddClass("panel")

	root.AddChild(components.Row(left, right))
	root.Children[1].AddClass("layout")
	return root
}

func buildSwatchRow(index, label string) *giocss.Node {
	return components.Div([]string{"swatch-row"},
		components.Span("Row "+index, "swatch-label"),
		components.Span(label, "swatch-meta"),
	)
}

func main() {
	ss := giocss.NewStyleSheet()
	ss.ParseCSSText(css)
	root := buildUI()

	rt := giocss.NewWindowRuntime(
		giocss.WindowOptions{Title: "Sample 19 - Advanced Selectors", Width: 1080, Height: 700},
		giocss.WindowRuntimeHooks{
			Snapshot: func(size image.Point) giocss.WindowRuntimeSnapshot {
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
