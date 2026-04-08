package engine

import (
	"image"
	"log"
	"os"
	"time"

	gioapp "gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

type Window struct {
	appWin                     *gioapp.Window
	opts                       WindowOptions
	platformTransparencyTried  bool
	platformTransparencyActive bool
}

func (w *Window) Invalidate() {
	if w != nil && w.appWin != nil {
		w.appWin.Invalidate()
	}
}

func (w *Window) Close() {
	if w != nil && w.appWin != nil {
		w.appWin.Perform(system.ActionClose)
	}
}

func (w *Window) Inner() *gioapp.Window {
	if w == nil {
		return nil
	}
	return w.appWin
}

func (w *Window) SetTitle(title string) {
	if w != nil && w.appWin != nil {
		w.appWin.Option(gioapp.Title(title))
	}
}

type WindowOptions struct {
	Title       string
	Width       int
	Height      int
	DisableGPU  bool
	Transparent bool
}

func NewWindow(opts WindowOptions) *Window {
	w := &gioapp.Window{}
	windowOpts := []gioapp.Option{
		gioapp.Title(opts.Title),
		gioapp.Size(unit.Dp(float32(opts.Width)), unit.Dp(float32(opts.Height))),
		gioapp.CustomRenderer(opts.DisableGPU),
	}
	if opts.Transparent {
		windowOpts = append(windowOpts, gioapp.Decorated(false))
	}
	w.Option(windowOpts...)
	return &Window{appWin: w, opts: opts}
}

func RunApp() {
	gioapp.Main()
}

type RenderContext struct {
	Gtx layout.Context
}

type GioWindowLoopHooks struct {
	OnDestroy     func()
	OnBeforeFrame func(size image.Point, frameTime time.Time)
	ProcessEvents func(rctx *RenderContext, frameTime time.Time) bool
	RenderFrame   func(rctx *RenderContext, frameTime time.Time)
	OnAfterFrame  func(frameTime time.Time)
}

func RunGioWindowLoop(gw *Window, hooks GioWindowLoopHooks) {
	if gw == nil || gw.appWin == nil {
		return
	}
	debugLoop := os.Getenv("GIOCSS_DEBUG_LOOP") == "1"
	var ops op.Ops
	for {
		e := gw.appWin.Event()
		switch e := e.(type) {
		case gioapp.DestroyEvent:
			if debugLoop {
				log.Printf("[window_loop] destroy event: %v", e.Err)
			}
			if hooks.OnDestroy != nil {
				hooks.OnDestroy()
			}
			return
		case gioapp.FrameEvent:
			if debugLoop {
				log.Printf("[window_loop] frame: %dx%d", e.Size.X, e.Size.Y)
			}
			gtx := gioapp.NewContext(&ops, e)
			frameTime := time.Now()
			rctx := &RenderContext{Gtx: gtx}
			if hooks.OnBeforeFrame != nil {
				hooks.OnBeforeFrame(e.Size, frameTime)
			}
			layoutDirty := false
			if hooks.ProcessEvents != nil {
				layoutDirty = hooks.ProcessEvents(rctx, frameTime)
			}
			if layoutDirty {
				gw.Invalidate()
			}
			if hooks.RenderFrame != nil {
				hooks.RenderFrame(rctx, frameTime)
			}
			if hooks.OnAfterFrame != nil {
				hooks.OnAfterFrame(frameTime)
			}
			e.Frame(gtx.Ops)
		default:
			if debugLoop {
				log.Printf("[window_loop] platform event: %T", e)
			}
			handlePlatformWindowEvent(gw, e)
		}
	}
}
