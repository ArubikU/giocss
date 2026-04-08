package preview

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	giocss "github.com/ArubikU/giocss"
	core "github.com/ArubikU/giocss/core"
)

const OutputPathEnv = "GIOCSS_PREVIEW_PATH"

func RequestedOutputPath() string {
	return os.Getenv(OutputPathEnv)
}

func CaptureSnapshotPNG(opts giocss.WindowOptions, hooks giocss.WindowRuntimeHooks, outputPath string) error {
	outputPath = filepath.Clean(outputPath)
	if outputPath == "." || outputPath == "" {
		return fmt.Errorf("preview output path is required")
	}

	size := image.Pt(maxInt(1, opts.Width), maxInt(1, opts.Height))
	snapshot := giocss.WindowRuntimeSnapshot{}
	if hooks.Snapshot != nil {
		snapshot = hooks.Snapshot(size)
	}
	screenW := maxInt(1, snapshot.ScreenWidth)
	screenH := maxInt(1, snapshot.ScreenHeight)
	if screenW <= 1 {
		screenW = size.X
	}
	if screenH <= 1 {
		screenH = size.Y
	}
	size = image.Pt(screenW, screenH)

	win, err := headless.NewWindow(size.X, size.Y)
	if err != nil {
		return err
	}
	defer win.Release()

	frame := buildFrameSnapshot(size, snapshot)
	ops := new(op.Ops)
	rctx := &core.RenderContext{Gtx: layout.Context{
		Ops:         ops,
		Now:         time.Now(),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Exact(size),
	}}
	state := core.NewGioWindowState()
	host := &core.GioRenderHost{
		HoverState:         frame.HoverState,
		ActiveState:        frame.ActiveState,
		StyleSheet:         frame.StyleSheet,
		RenderStore:        frame.RenderStore,
		CanUseWindowEffects: true,
	}
	state.BeginFrame(size, frame.FrameNumber, host.HoverState, host.ActiveState, frame.Debug, host.StyleSheet, frame.ProfileComponents, frame.ProfileFull)
	state.PrepareRenderFrame()
	if frame.RootLayout != nil {
		core.DrawGioBackground(rctx, 0, 0, frame.ScreenW, frame.ScreenH, frame.RootCSS)
		core.DrawGioTree(rctx, frame.RootLayout, host, "root", state, "", image.Rect(0, 0, frame.ScreenW, frame.ScreenH), nil, 0, 0, time.Now())
		core.DrawSelectDropdownOverlay(rctx, host, state, frame.ScreenW, frame.ScreenH)
		core.DrawPickerModal(rctx, state, frame.ScreenW, frame.ScreenH)
		core.ApplyFrameCursorOverride(rctx, state)
	}
	if host.RenderStore != nil {
		host.RenderStore.Finalize()
	}
	if err := win.Frame(ops); err != nil {
		return err
	}

	img := image.NewRGBA(image.Rectangle{Max: size})
	if err := win.Screenshot(img); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

func buildFrameSnapshot(size image.Point, snapshot giocss.WindowRuntimeSnapshot) core.WindowFrameSnapshot {
	screenW := snapshot.ScreenWidth
	screenH := snapshot.ScreenHeight
	if screenW <= 0 {
		screenW = size.X
	}
	if screenH <= 0 {
		screenH = size.Y
	}
	return core.WindowFrameSnapshot{
		HoverState:          map[string]bool{},
		ActiveState:         map[string]bool{},
		StyleSheet:          snapshot.StyleSheet,
		RenderStore:         core.NewRenderStore(),
		CanUseWindowEffects: true,
		FrameNumber:         1,
		RootLayout:          snapshot.RootLayout,
		RootCSS:             snapshot.RootCSS,
		ScreenW:             screenW,
		ScreenH:             screenH,
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}