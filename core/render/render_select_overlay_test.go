package render

import (
	"image"
	"testing"
)

func TestResolveSelectOverlayMenuRectPrefersOpeningUpWithinConstraint(t *testing.T) {
	selectBounds := image.Rect(20, 150, 220, 182)
	constraintRect := image.Rect(0, 0, 260, 190)

	rect := resolveSelectOverlayMenuRect(selectBounds, constraintRect, 320, 240, 32, 4)
	if rect.Min.Y >= selectBounds.Max.Y {
		t.Fatalf("menu opened downward at y=%d, want upward placement", rect.Min.Y)
	}
	if rect.Min.Y < constraintRect.Min.Y || rect.Max.Y > constraintRect.Max.Y {
		t.Fatalf("menu rect %v exceeds constraint %v", rect, constraintRect)
	}
}

func TestResolveSelectOverlayMenuRectClampsHorizontally(t *testing.T) {
	selectBounds := image.Rect(180, 20, 320, 52)
	constraintRect := image.Rect(40, 0, 240, 200)

	rect := resolveSelectOverlayMenuRect(selectBounds, constraintRect, 240, 200, 32, 3)
	if rect.Min.X < constraintRect.Min.X || rect.Max.X > constraintRect.Max.X {
		t.Fatalf("menu rect %v exceeds horizontal constraint %v", rect, constraintRect)
	}
}

func TestSelectOverlayOptionHoveredUsesFullRowRect(t *testing.T) {
	rowRect := image.Rect(50, 80, 250, 112)
	if !selectOverlayOptionHovered(true, image.Pt(240, 90), rowRect) {
		t.Fatal("expected point inside row to be treated as hovered")
	}
	if selectOverlayOptionHovered(true, image.Pt(251, 90), rowRect) {
		t.Fatal("expected point outside row to not be treated as hovered")
	}
}

func TestSelectOverlayBackgroundColorPrefersBackgroundColor(t *testing.T) {
	css := map[string]string{
		"background":       "#ffffff",
		"background-color": "#e8f3fb",
	}
	if got := selectOverlayBackgroundColor(css, "transparent"); got != "#e8f3fb" {
		t.Fatalf("expected background-color to win, got %q", got)
	}
}

func TestSelectOverlayBackgroundColorFallsBackToBackground(t *testing.T) {
	css := map[string]string{
		"background": "#e8f3fb",
	}
	if got := selectOverlayBackgroundColor(css, "transparent"); got != "#e8f3fb" {
		t.Fatalf("expected background fallback, got %q", got)
	}
}

func TestSelectOverlayMenuPaintCSSAddsDefaultsWhenMissing(t *testing.T) {
	css := selectOverlayMenuPaintCSS(nil)
	if css["background-color"] != "#ffffff" {
		t.Fatalf("expected default menu background-color, got %q", css["background-color"])
	}
	if css["border-width"] != "1px" || css["border-style"] != "solid" {
		t.Fatalf("expected default menu border, got width=%q style=%q", css["border-width"], css["border-style"])
	}
}

func TestSelectOverlayOptionPaintCSSKeepsExplicitTopBorder(t *testing.T) {
	base := map[string]string{"border-top": "none"}
	css := selectOverlayOptionPaintCSS(base, true, false)
	if css["border-top-width"] != "" {
		t.Fatalf("expected no injected divider when top border is explicit, got %q", css["border-top-width"])
	}
}

func TestSelectOverlayOptionHeightRespectsMinMax(t *testing.T) {
	css := map[string]string{
		"height":     "20px",
		"min-height": "28px",
		"max-height": "26px",
	}
	got := selectOverlayOptionHeight(css, 32, 320, 240)
	if got != 26 {
		t.Fatalf("expected clamped option height 26, got %d", got)
	}
}

func TestSelectOverlayMenuInsetsAndContentRect(t *testing.T) {
	css := map[string]string{
		"padding-left":   "8px",
		"padding-right":  "6px",
		"padding-top":    "4px",
		"padding-bottom": "2px",
		"border-width":   "1px",
		"border-style":   "solid",
	}
	insets := selectOverlayMenuInsetsFromCSS(css, 200, 100, 320, 240)
	if insets.InsetX() != 16 {
		t.Fatalf("expected insetX=16, got %d", insets.InsetX())
	}
	if insets.InsetY() != 8 {
		t.Fatalf("expected insetY=8, got %d", insets.InsetY())
	}
	rect := insets.ContentRect(image.Rect(10, 20, 210, 120))
	if rect.Min.X != 19 || rect.Min.Y != 25 || rect.Max.X != 203 || rect.Max.Y != 117 {
		t.Fatalf("unexpected content rect: %v", rect)
	}
}

func TestSelectOverlayMenuHeightIncludesContainerInsets(t *testing.T) {
	css := map[string]string{
		"padding-top":    "5px",
		"padding-bottom": "7px",
		"border-top-width":  "2px",
		"border-bottom-width": "2px",
		"border-style":      "solid",
	}
	insets := selectOverlayMenuInsetsFromCSS(css, 180, 90, 320, 240)
	h := selectOverlayMenuHeight(css, 90, insets, 320, 240)
	if h != 106 {
		t.Fatalf("expected menu outer height 106, got %d", h)
	}
}
