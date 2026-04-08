//go:build windows

package engine

import (
	"log"
	"os"
	"unsafe"

	gioapp "gioui.org/app"
	win "golang.org/x/sys/windows"
)

const (
	gwlExStyle     = int32(-20)
	wsExLayered    = 0x00080000
	lwaAlpha       = 0x00000002
	swpNoMove      = 0x0002
	swpNoSize      = 0x0001
	swpNoZOrder    = 0x0004
	swpFrameChange = 0x0020
)

var (
	user32ProcGetWindowLongPtrW          = win.NewLazySystemDLL("user32.dll").NewProc("GetWindowLongPtrW")
	user32ProcSetWindowLongPtrW          = win.NewLazySystemDLL("user32.dll").NewProc("SetWindowLongPtrW")
	user32ProcSetLayeredWindowAttributes = win.NewLazySystemDLL("user32.dll").NewProc("SetLayeredWindowAttributes")
	user32ProcSetWindowPos               = win.NewLazySystemDLL("user32.dll").NewProc("SetWindowPos")
	dwmapiProcExtendFrame                = win.NewLazySystemDLL("dwmapi.dll").NewProc("DwmExtendFrameIntoClientArea")
)

type dwmMargins struct {
	Left   int32
	Right  int32
	Top    int32
	Bottom int32
}

func handlePlatformWindowEvent(w *Window, e any) {
	if w == nil || !w.opts.Transparent || w.platformTransparencyTried {
		return
	}
	debugTransparency := os.Getenv("GIOCSS_DEBUG_TRANSPARENCY") == "1"
	if os.Getenv("GIOCSS_DISABLE_NATIVE_TRANSPARENCY") == "1" {
		w.platformTransparencyTried = true
		if debugTransparency {
			log.Printf("[transparency] disabled by env")
		}
		return
	}
	view, ok := e.(gioapp.Win32ViewEvent)
	if !ok || !view.Valid() || view.HWND == 0 {
		return
	}
	w.platformTransparencyTried = true
	if debugTransparency {
		log.Printf("[transparency] trying hwnd=%d", view.HWND)
	}
	hwnd := win.Handle(view.HWND)
	go func() {
		err := applyTransparentWindowStyle(hwnd)
		if err == nil {
			w.platformTransparencyActive = true
			if w.appWin != nil {
				w.appWin.Invalidate()
			}
			if debugTransparency {
				log.Printf("[transparency] active=true")
			}
			return
		}
		if debugTransparency {
			log.Printf("[transparency] apply failed: %v", err)
		}
	}()
}

func applyTransparentWindowStyle(hwnd win.Handle) error {
	useDWM := os.Getenv("GIOCSS_TRANSPARENT_USE_DWM") == "1"
	debugTransparency := os.Getenv("GIOCSS_DEBUG_TRANSPARENCY") == "1"
	if hwnd == 0 {
		return win.ERROR_INVALID_WINDOW_HANDLE
	}
	if debugTransparency {
		log.Printf("[transparency] step=getWindowLongPtr")
	}
	exStyle, err := getWindowLongPtr(hwnd, gwlExStyle)
	if err != nil {
		return err
	}
	if exStyle&wsExLayered == 0 {
		if debugTransparency {
			log.Printf("[transparency] step=setWindowLongPtr")
		}
		if err := setWindowLongPtr(hwnd, gwlExStyle, exStyle|wsExLayered); err != nil {
			return err
		}
	}
	if debugTransparency {
		log.Printf("[transparency] step=setLayeredWindowAttributes")
	}
	if _, _, callErr := user32ProcSetLayeredWindowAttributes.Call(uintptr(hwnd), 0, 255, lwaAlpha); callErr != nil && callErr != win.Errno(0) {
		return callErr
	}
	if useDWM {
		if debugTransparency {
			log.Printf("[transparency] step=dwmExtendFrameIntoClientArea")
		}
		if err := dwmExtendFrameIntoClientArea(hwnd, dwmMargins{Left: -1, Right: -1, Top: -1, Bottom: -1}); err != nil {
			return err
		}
	}
	if debugTransparency {
		log.Printf("[transparency] step=setWindowPos")
	}
	if _, _, callErr := user32ProcSetWindowPos.Call(uintptr(hwnd), 0, 0, 0, 0, 0, swpNoMove|swpNoSize|swpNoZOrder|swpFrameChange); callErr != nil && callErr != win.Errno(0) {
		return callErr
	}
	if debugTransparency {
		log.Printf("[transparency] step=done")
	}
	return nil
}

func getWindowLongPtr(hwnd win.Handle, index int32) (uintptr, error) {
	r0, _, callErr := user32ProcGetWindowLongPtrW.Call(uintptr(hwnd), uintptr(int64(index)))
	if r0 == 0 && callErr != nil && callErr != win.Errno(0) {
		return 0, callErr
	}
	return r0, nil
}

func setWindowLongPtr(hwnd win.Handle, index int32, value uintptr) error {
	r0, _, callErr := user32ProcSetWindowLongPtrW.Call(uintptr(hwnd), uintptr(int64(index)), value)
	if r0 == 0 && callErr != nil && callErr != win.Errno(0) {
		return callErr
	}
	return nil
}

func dwmExtendFrameIntoClientArea(hwnd win.Handle, margins dwmMargins) error {
	r0, _, callErr := dwmapiProcExtendFrame.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&margins)))
	if r0 != 0 {
		if callErr != nil && callErr != win.Errno(0) {
			return callErr
		}
		return win.Errno(r0)
	}
	return nil
}
