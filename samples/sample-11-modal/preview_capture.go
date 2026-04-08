package main

import (
	"image"
	"log"
	"os"

	giocss "github.com/ArubikU/giocss"
	"github.com/ArubikU/giocss/samples/internal/preview"
)

func init() {
	outputPath := preview.RequestedOutputPath()
	if outputPath == "" {
		return
	}

	ss := giocss.NewStyleSheet()
	ss.ParseCSSText(css)
	app := newAppState()
	err := preview.CaptureSnapshotPNG(
		giocss.WindowOptions{Title: "Sample 11 - Modals via Absolute/Z-Index", Width: 800, Height: 600},
		giocss.WindowRuntimeHooks{
			Snapshot: func(size image.Point) giocss.WindowRuntimeSnapshot {
				isOpen, title, body := app.snapshot()
				root := buildUI(isOpen, title, body)
				return giocss.WindowRuntimeSnapshot{RootLayout: giocss.LayoutNodeToNative(root, size.X, size.Y, ss), RootCSS: giocss.ResolveNodeStyle(root, ss, size.X), StyleSheet: ss, ScreenWidth: size.X, ScreenHeight: size.Y}
			},
		},
		outputPath,
	)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}