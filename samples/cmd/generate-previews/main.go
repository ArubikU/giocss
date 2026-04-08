package main

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"

	giocss "github.com/ArubikU/giocss"
	"github.com/ArubikU/giocss/components"
	"github.com/ArubikU/giocss/samples/internal/preview"
)

type sampleTarget struct {
	Dir         string
	PreviewPath string
	ExtraEnv    map[string]string
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fatal(err)
	}

	targets := []sampleTarget{
		{Dir: "sample-01-hello-world", PreviewPath: filepath.Join("samples", "sample-01-hello-world", "preview.png")},
		{Dir: "sample-02-buttons", PreviewPath: filepath.Join("samples", "sample-02-buttons", "preview.png")},
		{Dir: "sample-03-login-form", PreviewPath: filepath.Join("samples", "sample-03-login-form", "preview.png")},
		{Dir: "sample-04-cards", PreviewPath: filepath.Join("samples", "sample-04-cards", "preview.png")},
		{Dir: "sample-05-navigation", PreviewPath: filepath.Join("samples", "sample-05-navigation", "preview.png")},
		{Dir: "sample-06-typography", PreviewPath: filepath.Join("samples", "sample-06-typography", "preview.png")},
		{Dir: "sample-07-color-swatches", PreviewPath: filepath.Join("samples", "sample-07-color-swatches", "preview.png")},
		{Dir: "sample-08-todo-list", PreviewPath: filepath.Join("samples", "sample-08-todo-list", "preview.png")},
		{Dir: "sample-09-dashboard", PreviewPath: filepath.Join("samples", "sample-09-dashboard", "preview.png")},
		{Dir: "sample-10-dark-theme", PreviewPath: filepath.Join("samples", "sample-10-dark-theme", "preview.png")},
		{Dir: "sample-11-modal", PreviewPath: filepath.Join("samples", "sample-11-modal", "preview.png")},
		{Dir: "sample-12-data-table", PreviewPath: filepath.Join("samples", "sample-12-data-table", "preview.png")},
		{Dir: "sample-13-tabs", PreviewPath: filepath.Join("samples", "sample-13-tabs", "preview.png")},
		{Dir: "sample-14-accordion", PreviewPath: filepath.Join("samples", "sample-14-accordion", "preview.png")},
		{Dir: "sample-15-notification-center", PreviewPath: filepath.Join("samples", "sample-15-notification-center", "preview.png")},
		{Dir: "sample-16-side-drawer", PreviewPath: filepath.Join("samples", "sample-16-side-drawer", "preview.png")},
		{Dir: "sample-17-search-autocomplete", PreviewPath: filepath.Join("samples", "sample-17-search-autocomplete", "preview.png")},
		{Dir: "sample-18-docs-viewer", PreviewPath: filepath.Join("samples", "sample-18-docs-viewer", "preview.png")},
		{Dir: "sample-19-advanced-selectors", PreviewPath: filepath.Join("samples", "sample-19-advanced-selectors", "preview.png")},
		{Dir: "sample-20-form-rerender", PreviewPath: filepath.Join("samples", "sample-20-form-rerender", "preview.png")},
		{Dir: "sample-21-transparent-todo-board", PreviewPath: filepath.Join("samples", "sample-21-transparent-todo-board", "preview.png"), ExtraEnv: map[string]string{"GIOCSS_SAMPLE21_TRANSPARENT": "0", "POLYLOFT_GIO_LOW_POWER": "1"}},
	}

	for _, target := range targets {
		fmt.Printf("Generating %s\n", target.PreviewPath)
		if err := renderSample(root, target); err != nil {
			fatal(err)
		}
	}

	if err := renderRootREADMEPreview(filepath.Join(root, "readme-preview.png")); err != nil {
		fatal(err)
	}
	fmt.Println("Generated sample previews and root README preview")
}

func renderSample(root string, target sampleTarget) error {
	cmd := exec.Command("go", "run", "./samples/"+target.Dir)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	env := append([]string{}, os.Environ()...)
	env = append(env, preview.OutputPathEnv+"="+filepath.Join(root, target.PreviewPath))
	for key, value := range target.ExtraEnv {
		env = append(env, key+"="+value)
	}
	cmd.Env = env
	return cmd.Run()
}

func renderRootREADMEPreview(outputPath string) error {
	ss := giocss.NewStyleSheet()
	ss.ParseCSSText(`
body {
  background: linear-gradient(180deg, #0b1020 0%, #131b33 100%);
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 480px;
  padding: 24px;
  background: #101820;
  color: white;
  border-radius: 20px;
  box-shadow: 0 18px 40px rgba(0,0,0,0.28);
}

.title {
  font-size: 26px;
  font-weight: bold;
}

.copy {
  color: #b9c7f5;
  line-height: 1.45;
}
`)
	return preview.CaptureSnapshotPNG(
		giocss.WindowOptions{Title: "giocss README preview", Width: 900, Height: 640},
		giocss.WindowRuntimeHooks{
			Snapshot: func(size image.Point) giocss.WindowRuntimeSnapshot {
				root := giocss.NewNode("body")
				card := components.Column(
					components.Text("Hello giocss", "title"),
					components.Text("A lightweight CSS-driven UI layer for Gio runtimes, layouts, and desktop experiments.", "copy"),
				)
				card.AddClass("card")
				root.AddChild(card)
				return giocss.WindowRuntimeSnapshot{RootLayout: giocss.LayoutNodeToNative(root, size.X, size.Y, ss), RootCSS: giocss.ResolveNodeStyle(root, ss, size.X), StyleSheet: ss, ScreenWidth: size.X, ScreenHeight: size.Y}
			},
		},
		outputPath,
	)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}