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

	transparentWindow := os.Getenv("GIOCSS_SAMPLE21_TRANSPARENT") != "0"
	ss := giocss.NewStyleSheet()
	ss.ParseCSSText(css)
	state := newAppState()
	err := preview.CaptureSnapshotPNG(
		giocss.WindowOptions{Title: "Sample 21 - Transparent Todo Board", Width: 1280, Height: 820, Transparent: transparentWindow},
		giocss.WindowRuntimeHooks{
			Snapshot: func(size image.Point) giocss.WindowRuntimeSnapshot {
				snapshot := state.snapshot(size)
				root := buildUI(snapshot)
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