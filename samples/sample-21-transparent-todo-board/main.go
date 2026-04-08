package main

import (
	"fmt"
	"image"
	"log"
	"os"
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
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 28px;
  box-sizing: border-box;
  background-color: transparent;
}

.shell {
  width: 100%;
  max-width: 1120px;
  padding: 22px 24px 24px;
  box-sizing: border-box;
  border-radius: 24px;
  background-color: rgba(248, 250, 252, 0.86);
  border: 1px solid rgba(148, 163, 184, 0.34);
  box-shadow: 0 24px 60px rgba(15, 23, 42, 0.24);
}

.header {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  margin-bottom: 18px;
}

.title-wrap {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.eyebrow {
  font-size: 11px;
  font-weight: bold;
  letter-spacing: 1.6px;
  text-transform: uppercase;
  color: #2563eb;
}

.title {
  margin: 0;
  font-size: 30px;
  font-weight: bold;
  line-height: 1.05;
  color: #0f172a;
}

.subtitle {
  font-size: 14px;
  line-height: 1.4;
  color: #334155;
  max-width: 640px;
}

.actions {
  display: flex;
  flex-direction: row;
  gap: 10px;
}

.pill-btn {
  min-width: 96px;
  padding: 10px 14px;
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.42);
  background-color: rgba(255, 255, 255, 0.82);
  color: #0f172a;
  cursor: pointer;
  font-size: 12px;
  font-weight: bold;
}

.pill-btn:hover {
  background-color: rgba(239, 246, 255, 0.96);
}

.status-strip {
  display: flex;
  flex-direction: row;
  gap: 10px;
  margin-bottom: 18px;
  flex-wrap: wrap;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 7px 12px;
  border-radius: 999px;
  background-color: rgba(15, 23, 42, 0.08);
  color: #334155;
  font-size: 12px;
}

.board {
  display: flex;
  flex-direction: row;
  gap: 18px;
  align-items: flex-start;
}

.lane {
  flex: 1;
  min-width: 0;
  min-height: 560px;
  padding: 14px;
  border-radius: 22px;
  background-color: rgba(255, 255, 255, 0.72);
  border: 1px solid rgba(148, 163, 184, 0.28);
  box-sizing: border-box;
}

.lane-target {
  border-color: rgba(59, 130, 246, 0.9);
  box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.2);
}

.lane-head {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.lane-title {
  font-size: 15px;
  font-weight: bold;
  color: #0f172a;
}

.lane-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 26px;
  height: 26px;
  padding: 0 8px;
  border-radius: 999px;
  background-color: rgba(15, 23, 42, 0.08);
  font-size: 12px;
  color: #475569;
}

.lane-stack {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.drop-slot {
  height: 12px;
  border-radius: 999px;
  background-color: rgba(96, 165, 250, 0.9);
  box-shadow: 0 0 0 4px rgba(191, 219, 254, 0.9);
}

.empty-slot {
  padding: 24px 14px;
  border-radius: 16px;
  border: 1px dashed rgba(148, 163, 184, 0.52);
  color: #64748b;
  font-size: 13px;
  line-height: 1.4;
  text-align: center;
}

.card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  border-radius: 18px;
  background-color: rgba(255, 255, 255, 0.96);
  border: 1px solid rgba(148, 163, 184, 0.24);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.08);
  cursor: grab;
  box-sizing: border-box;
}

.card:hover {
  background-color: rgba(255, 255, 255, 1);
}

.card-dragging {
  opacity: 0.42;
}

.drag-ghost {
	display: flex;
	flex-direction: column;
	gap: 10px;
	padding: 14px;
	border-radius: 18px;
	background-color: rgba(255, 255, 255, 0.96);
	border: 1px solid rgba(148, 163, 184, 0.24);
	box-sizing: border-box;
  opacity: 0.84;
  cursor: grabbing;
  box-shadow: 0 26px 48px rgba(15, 23, 42, 0.22);
  border-color: rgba(96, 165, 250, 0.56);
}

.card-top {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.card-title {
  font-size: 15px;
  font-weight: bold;
  line-height: 1.2;
  color: #0f172a;
}

.card-body {
  font-size: 13px;
  line-height: 1.45;
  color: #475569;
}

.card-footer {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 5px 9px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: bold;
  color: #0f172a;
}

.tag-blue { background-color: #dbeafe; }
.tag-amber { background-color: #fef3c7; }
.tag-rose { background-color: #ffe4e6; }
.tag-emerald { background-color: #d1fae5; }

.assignee {
  font-size: 12px;
  color: #64748b;
}

.drag-note {
  margin-top: 18px;
  font-size: 13px;
  line-height: 1.45;
  color: #475569;
}
`

var laneOrder = []string{"todo", "doing", "done"}

var laneTitles = map[string]string{
	"todo":  "To do",
	"doing": "Doing",
	"done":  "Done",
}

type boardCard struct {
	ID       string
	Title    string
	Body     string
	Tag      string
	TagClass string
	Assignee string
}

type dragState struct {
	CardID      string
	SourceLane  string
	SourceIndex int
	TargetLane  string
	TargetIndex int
	StartX      int
	StartY      int
	PointerX    int
	PointerY    int
	PointerDown bool
	IsDragging  bool
}

type boardSnapshot struct {
	Lanes    map[string][]boardCard
	Drag     dragState
	Viewport image.Point
	Geometry boardGeometry
}

type boardGeometry struct {
	ShellX   int
	ShellY   int
	ShellW   int
	LaneTop  int
	LaneW    int
	LaneGap  int
	CardsTop int
	CardH    int
	CardGap  int
}

type appState struct {
	mu       sync.Mutex
	lanes    map[string][]boardCard
	drag     dragState
	viewport image.Point
	geoCache boardGeometry
	geoSize  image.Point
	geoValid bool
}

func seedBoard() map[string][]boardCard {
	return map[string][]boardCard{
		"todo": {
			{ID: "card-1", Title: "Floating shell polish", Body: "Keep the glass shell transparent around the edges while cards stay readable and draggable.", Tag: "Visual", TagClass: "tag-blue", Assignee: "Elena"},
			{ID: "card-2", Title: "日本語 copy check", Body: "Confirm short Japanese labels fit the current shaper without clipping in compact cards.", Tag: "i18n", TagClass: "tag-amber", Assignee: "Mika"},
			{ID: "card-3", Title: "Pointer payload", Body: "Send x/y and delta data through onpointer handlers so drag state can live in app code.", Tag: "Runtime", TagClass: "tag-emerald", Assignee: "Luis"},
		},
		"doing": {
			{ID: "card-4", Title: "Ghost card tracking", Body: "Render an absolute ghost near the cursor and keep the source card semi-transparent during drag.", Tag: "Interaction", TagClass: "tag-blue", Assignee: "Nora"},
			{ID: "card-5", Title: "Entrega visual", Body: "Spanish copy should wrap cleanly while the lane target stays obvious on drop.", Tag: "UX", TagClass: "tag-rose", Assignee: "Pablo"},
		},
		"done": {
			{ID: "card-6", Title: "Windows-first window hook", Body: "Capture Win32ViewEvent and apply the native transparency style without breaking other platforms.", Tag: "Platform", TagClass: "tag-emerald", Assignee: "Ari"},
		},
	}
}

func newAppState() *appState {
	return &appState{lanes: seedBoard()}
}

func (s *appState) snapshot(size image.Point) boardSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	if size.X > 0 && size.Y > 0 {
		if s.viewport != size {
			s.viewport = size
			s.geoValid = false
		}
	}
	lanes := make(map[string][]boardCard, len(laneOrder))
	for _, lane := range laneOrder {
		items := s.lanes[lane]
		copyItems := make([]boardCard, len(items))
		copy(copyItems, items)
		lanes[lane] = copyItems
	}
	return boardSnapshot{Lanes: lanes, Drag: s.drag, Viewport: s.viewport, Geometry: s.getBoardGeometryLocked()}
}

func (s *appState) getBoardGeometryLocked() boardGeometry {
	if s.geoValid && s.geoSize == s.viewport {
		return s.geoCache
	}
	s.geoCache = computeBoardGeometry(s.viewport)
	s.geoSize = s.viewport
	s.geoValid = true
	return s.geoCache
}

func (s *appState) onEvent(eventName string, payload map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch {
	case eventName == "board.reset":
		s.lanes = seedBoard()
		s.drag = dragState{}
	case eventName == "board.pointermove":
		s.moveDragFromAnyLocked(eventName, payload)
	case eventName == "board.pointerup":
		s.endDragFromAnyLocked(payload)
	case strings.HasPrefix(eventName, "card.pointerdown."):
		s.startDragLocked(strings.TrimPrefix(eventName, "card.pointerdown."), payload)
	case strings.HasPrefix(eventName, "card.pointermove."):
		s.moveDragLocked(strings.TrimPrefix(eventName, "card.pointermove."), payload)
	case strings.HasPrefix(eventName, "card.pointerup."):
		s.endDragLocked(strings.TrimPrefix(eventName, "card.pointerup."), payload)
	}
	return nil
}

func (s *appState) startDragLocked(cardID string, payload map[string]any) {
	lane, index, ok := s.findCardLocked(cardID)
	if !ok {
		return
	}
	x := payloadInt(payload["x"])
	y := payloadInt(payload["y"])
	s.drag = dragState{
		CardID:      cardID,
		SourceLane:  lane,
		SourceIndex: index,
		TargetLane:  lane,
		TargetIndex: index,
		StartX:      x,
		StartY:      y,
		PointerX:    x,
		PointerY:    y,
		PointerDown: true,
		IsDragging:  false,
	}
}

func (s *appState) moveDragLocked(cardID string, payload map[string]any) {
	if s.drag.CardID == "" || s.drag.CardID != cardID {
		return
	}
	s.moveDragFromAnyLocked("card.pointermove."+cardID, payload)
}

func (s *appState) moveDragFromAnyLocked(eventName string, payload map[string]any) {
	if s.drag.CardID == "" || !s.drag.PointerDown {
		return
	}
	x := payloadInt(payload["x"])
	y := payloadInt(payload["y"])
	pressed := payloadBool(payload["pressed"])
	if !pressed {
		s.drag.PointerX = x
		s.drag.PointerY = y
		if s.drag.IsDragging {
			s.finalizeDragToCurrentTargetLocked()
		} else {
			s.drag = dragState{}
		}
		return
	}
	if !s.drag.IsDragging {
		_ = eventName
		dx0 := x - s.drag.StartX
		dy0 := y - s.drag.StartY
		if dx0*dx0+dy0*dy0 < 25 {
			return
		}
		s.drag.IsDragging = true
	}
	s.drag.PointerX = x
	s.drag.PointerY = y
	targetLane, targetIndex := s.resolveDropTargetLocked(x, y)
	if targetLane != "" {
		s.drag.TargetLane = targetLane
		s.drag.TargetIndex = targetIndex
	}
}

func (s *appState) endDragLocked(cardID string, payload map[string]any) {
	if s.drag.CardID == "" || s.drag.CardID != cardID {
		return
	}
	s.endDragFromAnyLocked(payload)
}

func (s *appState) endDragFromAnyLocked(payload map[string]any) {
	if s.drag.CardID == "" || !s.drag.PointerDown {
		return
	}
	if !s.drag.IsDragging {
		s.drag = dragState{}
		return
	}
	s.drag.PointerX = payloadInt(payload["x"])
	s.drag.PointerY = payloadInt(payload["y"])
	s.finalizeDragToCurrentTargetLocked()
}

func (s *appState) finalizeDragToCurrentTargetLocked() {
	targetLane := s.drag.TargetLane
	targetIndex := s.drag.TargetIndex
	if targetLane == "" {
		targetLane = s.drag.SourceLane
		targetIndex = s.drag.SourceIndex
	}
	s.moveCardLocked(s.drag.CardID, targetLane, targetIndex)
	s.drag = dragState{}
}

func (s *appState) findCardLocked(cardID string) (string, int, bool) {
	for _, lane := range laneOrder {
		for index, card := range s.lanes[lane] {
			if card.ID == cardID {
				return lane, index, true
			}
		}
	}
	return "", -1, false
}

func (s *appState) moveCardLocked(cardID, targetLane string, targetIndex int) {
	sourceLane, sourceIndex, ok := s.findCardLocked(cardID)
	if !ok {
		return
	}
	card := s.lanes[sourceLane][sourceIndex]
	updatedSource := append([]boardCard{}, s.lanes[sourceLane][:sourceIndex]...)
	updatedSource = append(updatedSource, s.lanes[sourceLane][sourceIndex+1:]...)
	s.lanes[sourceLane] = updatedSource

	if targetLane == "" {
		targetLane = sourceLane
	}
	targetCards := append([]boardCard{}, s.lanes[targetLane]...)
	if sourceLane == targetLane && targetIndex > sourceIndex {
		targetIndex--
	}
	if targetIndex < 0 {
		targetIndex = 0
	}
	if targetIndex > len(targetCards) {
		targetIndex = len(targetCards)
	}
	targetCards = append(targetCards, boardCard{})
	copy(targetCards[targetIndex+1:], targetCards[targetIndex:])
	targetCards[targetIndex] = card
	s.lanes[targetLane] = targetCards
}

func (s *appState) resolveDropTargetLocked(x, y int) (string, int) {
	if s.drag.CardID == "" {
		return "", 0
	}
	geo := s.getBoardGeometryLocked()
	if x < geo.ShellX || x > geo.ShellX+geo.ShellW {
		return s.drag.TargetLane, s.drag.TargetIndex
	}
	innerX := x - (geo.ShellX + 24)
	if innerX < 0 {
		innerX = 0
	}
	laneIndex := resolveLaneIndexWithHysteresis(innerX, geo, s.drag.TargetLane)
	targetLane := laneOrder[laneIndex]
	cards := cardsWithoutDragging(s.lanes[targetLane], s.drag.CardID)
	index := len(cards)
	relY := y - geo.CardsTop
	if relY < 0 {
		return targetLane, 0
	}
	for i := range cards {
		midpoint := i*(geo.CardH+geo.CardGap) + geo.CardH/2
		if relY < midpoint {
			index = i
			break
		}
	}
	if index < 0 {
		index = 0
	}
	if index > len(cards) {
		index = len(cards)
	}
	return targetLane, index
}

func resolveLaneIndexWithHysteresis(innerX int, geo boardGeometry, currentLane string) int {
	if len(laneOrder) == 0 {
		return 0
	}
	if innerX < 0 {
		innerX = 0
	}
	maxInner := len(laneOrder)*geo.LaneW + (len(laneOrder)-1)*geo.LaneGap - 1
	if maxInner > 0 && innerX > maxInner {
		innerX = maxInner
	}
	if currentLane != "" {
		if currentIndex := laneIndexOf(currentLane); currentIndex >= 0 {
			left := currentIndex*(geo.LaneW+geo.LaneGap) - geo.LaneGap/2
			right := left + geo.LaneW + geo.LaneGap
			if innerX >= left && innerX <= right {
				return currentIndex
			}
		}
	}
	bestIndex := 0
	bestDist := 1 << 30
	for i := range laneOrder {
		center := i*(geo.LaneW+geo.LaneGap) + geo.LaneW/2
		dist := absInt(innerX - center)
		if dist < bestDist {
			bestDist = dist
			bestIndex = i
		}
	}
	return bestIndex
}

func laneIndexOf(lane string) int {
	for i, id := range laneOrder {
		if id == lane {
			return i
		}
	}
	return -1
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func cardsWithoutDragging(cards []boardCard, draggingID string) []boardCard {
	out := make([]boardCard, 0, len(cards))
	for _, card := range cards {
		if card.ID == draggingID {
			continue
		}
		out = append(out, card)
	}
	return out
}

func computeBoardGeometry(viewport image.Point) boardGeometry {
	width := viewport.X - 56
	if width > 1120 {
		width = 1120
	}
	if width < 760 {
		width = max(640, viewport.X-32)
	}
	if width <= 0 {
		width = 960
	}
	shellX := max(16, (viewport.X-width)/2)
	laneGap := 18
	laneW := (width - 48 - laneGap*2) / 3
	return boardGeometry{
		ShellX:   shellX,
		ShellY:   28,
		ShellW:   width,
		LaneTop:  144,
		LaneW:    laneW,
		LaneGap:  laneGap,
		CardsTop: 192,
		CardH:    124,
		CardGap:  14,
	}
}

func payloadInt(raw any) int {
	switch typed := raw.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(typed))
		return n
	default:
		return 0
	}
}

func payloadBool(raw any) bool {
	b, ok := raw.(bool)
	if ok {
		return b
	}
	return false
}

func laneChipText(snapshot boardSnapshot) string {
	if snapshot.Drag.CardID == "" {
		return "Drag a card between columns"
	}
	if !snapshot.Drag.IsDragging {
		return "Press and move to start dragging"
	}
	if snapshot.Drag.TargetLane == "" {
		return "Dragging"
	}
	return "Drop target: " + laneTitles[snapshot.Drag.TargetLane]
}

func buildTag(text string, className string) *giocss.Node {
	tag := components.Badge(text, "tag")
	if className != "" {
		tag.AddClass(className)
	}
	return tag
}

func buildCardContent(card boardCard) []*giocss.Node {
	return []*giocss.Node{
		components.Row(
			components.Text(card.Title, "card-title"),
			buildTag(card.Tag, card.TagClass),
		),
		components.Text(card.Body, "card-body"),
		components.Row(
			components.Text("Owner", "assignee"),
			components.Text(card.Assignee, "assignee"),
		),
	}
}

func buildCardNode(card boardCard, dragging bool) *giocss.Node {
	cardNode := components.Div([]string{"card"}, buildCardContent(card)...)
	cardNode.Children[0].AddClass("card-top")
	cardNode.Children[2].AddClass("card-footer")
	cardNode.SetProp("onpointerdown", "card.pointerdown."+card.ID)
	cardNode.SetProp("onpointermove", "card.pointermove."+card.ID)
	cardNode.SetProp("onpointerup", "card.pointerup."+card.ID)
	if dragging {
		cardNode.AddClass("card-dragging")
	}
	return cardNode
}

func buildGhostNode(card boardCard, snapshot boardSnapshot) *giocss.Node {
	ghost := components.Div([]string{"drag-ghost"}, buildCardContent(card)...)
	ghost.Children[0].AddClass("card-top")
	ghost.Children[2].AddClass("card-footer")
	ghost.SetProp("onpointermove", "board.pointermove")
	ghost.SetProp("onpointerup", "board.pointerup")
	geo := snapshot.Geometry
	if geo.LaneW <= 0 {
		geo = computeBoardGeometry(snapshot.Viewport)
	}
	ghostW := max(220, geo.LaneW-8)
	ghost.SetProp("style.position", "absolute")
	ghost.SetProp("style.left", fmt.Sprintf("%dpx", max(18, snapshot.Drag.PointerX-ghostW/2)))
	ghost.SetProp("style.top", fmt.Sprintf("%dpx", max(18, snapshot.Drag.PointerY-geo.CardH/2)))
	ghost.SetProp("style.width", fmt.Sprintf("%dpx", ghostW))
	ghost.SetProp("style.z-index", "1200")
	return ghost
}

func buildLaneNode(snapshot boardSnapshot, lane string) *giocss.Node {
	visibleCards := snapshot.Lanes[lane]
	if snapshot.Drag.IsDragging && snapshot.Drag.CardID != "" {
		visibleCards = cardsWithoutDragging(snapshot.Lanes[lane], snapshot.Drag.CardID)
	}
	cardsNoDrag := visibleCards
	title := components.Text(laneTitles[lane], "lane-title")
	count := components.Text(strconv.Itoa(len(snapshot.Lanes[lane])), "lane-count")
	head := components.Row(title, count)
	head.AddClass("lane-head")

	stack := components.Column()
	stack.AddClass("lane-stack")
	insertedPlaceholder := false
	for index, card := range visibleCards {
		if snapshot.Drag.IsDragging && snapshot.Drag.CardID != "" && snapshot.Drag.TargetLane == lane && snapshot.Drag.TargetIndex == index {
			stack.AddChild(components.Div([]string{"drop-slot"}))
			insertedPlaceholder = true
		}
		stack.AddChild(buildCardNode(card, false))
	}
	if snapshot.Drag.IsDragging && snapshot.Drag.CardID != "" && snapshot.Drag.TargetLane == lane && snapshot.Drag.TargetIndex >= len(cardsNoDrag) {
		stack.AddChild(components.Div([]string{"drop-slot"}))
		insertedPlaceholder = true
	}
	if len(cardsNoDrag) == 0 && !insertedPlaceholder {
		stack.AddChild(components.Text("Drop a card here to start this lane.", "empty-slot"))
	}

	laneNode := components.Column(head, stack)
	laneNode.AddClass("lane")
	if snapshot.Drag.IsDragging && snapshot.Drag.CardID != "" && snapshot.Drag.TargetLane == lane {
		laneNode.AddClass("lane-target")
	}
	return laneNode
}

func buildUI(snapshot boardSnapshot) *giocss.Node {
	root := giocss.NewNode("body")
	root.SetProp("onpointermove", "board.pointermove")
	root.SetProp("onpointerup", "board.pointerup")

	titleWrap := components.Column(
		components.Text("Transparent Window", "eyebrow"),
		components.Heading(1, "Todo board with drag cards", "title"),
		components.Text("Windows-first transparent shell, explicit pointer events, and draggable cards that fade while moving. The copy mixes English, español, and 日本語 to keep the sample honest.", "subtitle"),
	)
	titleWrap.AddClass("title-wrap")

	actions := components.Row(
		components.Button("Reset board", "pill-btn"),
		components.Button("Close", "pill-btn"),
	)
	actions.AddClass("actions")
	actions.Children[0].SetProp("event", "board.reset")
	actions.Children[1].SetProp("event", "board.close")

	header := components.Row(titleWrap, actions)
	header.AddClass("header")

	status := components.Row(
		components.Text("Window mode: transparent", "status-chip"),
		components.Text(laneChipText(snapshot), "status-chip"),
	)
	status.AddClass("status-strip")

	board := components.Row()
	board.AddClass("board")
	board.SetProp("onpointermove", "board.pointermove")
	board.SetProp("onpointerup", "board.pointerup")
	for _, lane := range laneOrder {
		board.AddChild(buildLaneNode(snapshot, lane))
	}

	shell := components.Column(header, status, board, components.Text("This sample keeps the background transparent and renders a translucent shell on top. Drag a card, watch the source fade, and drop into another column.", "drag-note"))
	shell.AddClass("shell")
	root.AddChild(shell)

	if snapshot.Drag.IsDragging && snapshot.Drag.CardID != "" {
		if lane, index, ok := findCardInSnapshot(snapshot, snapshot.Drag.CardID); ok {
			_ = lane
			_ = index
			card := snapshot.Lanes[lane][index]
			root.AddChild(buildGhostNode(card, snapshot))
		}
	}

	return root
}

func findCardInSnapshot(snapshot boardSnapshot, cardID string) (string, int, bool) {
	for _, lane := range laneOrder {
		for index, card := range snapshot.Lanes[lane] {
			if card.ID == cardID {
				return lane, index, true
			}
		}
	}
	return "", -1, false
}

func main() {
	debugSample := os.Getenv("GIOCSS_SAMPLE21_DEBUG") == "1"
	transparentWindow := os.Getenv("GIOCSS_SAMPLE21_TRANSPARENT") != "0"
	if debugSample {
		log.Printf("[sample21] starting, transparent=%v", transparentWindow)
	}

	ss := giocss.NewStyleSheet()
	ss.ParseCSSText(css)
	state := newAppState()

	var rt *giocss.WindowRuntime
	rt = giocss.NewWindowRuntime(
		giocss.WindowOptions{
			Title:       "Sample 21 - Transparent Todo Board",
			Width:       1280,
			Height:      820,
			Transparent: transparentWindow,
		},
		giocss.WindowRuntimeHooks{
			DispatchEvent: func(eventName string, payload map[string]any) error {
				if debugSample {
					log.Printf("[sample21] event=%s payload=%v", eventName, payload)
				}
				if eventName == "board.close" {
					if rt != nil {
						rt.Close()
					}
					return nil
				}
				return state.onEvent(eventName, payload)
			},
			Snapshot: func(size image.Point) giocss.WindowRuntimeSnapshot {
				if debugSample {
					log.Printf("[sample21] snapshot size=%dx%d", size.X, size.Y)
				}
				snapshot := state.snapshot(size)
				root := buildUI(snapshot)
				return giocss.WindowRuntimeSnapshot{
					RootLayout:   giocss.LayoutNodeToNative(root, size.X, size.Y, ss),
					RootCSS:      giocss.ResolveNodeStyle(root, ss, size.X),
					StyleSheet:   ss,
					ScreenWidth:  size.X,
					ScreenHeight: size.Y,
				}
			},
			EmitRuntimeError: func(err error) {
				log.Printf("[sample21] runtime error: %v", err)
			},
			OnClose: func() {
				if debugSample {
					log.Printf("[sample21] closed")
				}
			},
		},
	)
	rt.Run()
}
