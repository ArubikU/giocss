package input

import "testing"

func TestNormalizeTypedInputValuePreservesWhitespaceForTextInputs(t *testing.T) {
	raw := "Test "
	gotLive := NormalizeTypedInputValue("text", raw, map[string]any{}, false)
	if gotLive != raw {
		t.Fatalf("NormalizeTypedInputValue(text, live) = %q, want %q", gotLive, raw)
	}

	gotFinal := NormalizeTypedInputValue("text", raw, map[string]any{}, true)
	if gotFinal != raw {
		t.Fatalf("NormalizeTypedInputValue(text, finalize) = %q, want %q", gotFinal, raw)
	}
}

func TestNormalizeTypedInputValueStillSanitizesUTF8(t *testing.T) {
	raw := string([]byte{'A', 0xff, 'B'})
	got := NormalizeTypedInputValue("text", raw, map[string]any{}, false)
	if got != "AB" {
		t.Fatalf("NormalizeTypedInputValue(text, utf8) = %q, want %q", got, "AB")
	}
}
