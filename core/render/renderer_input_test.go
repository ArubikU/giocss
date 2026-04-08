package render

import "testing"

func TestShouldSyncExternalInputRequiresExplicitExternalValue(t *testing.T) {
	inputExternal := map[string]string{"form/name": "hello"}

	if ShouldSyncExternalInput(inputExternal, "form/name", "", false) {
		t.Fatal("ShouldSyncExternalInput() returned true for missing external value")
	}
}

func TestShouldSyncExternalInputAllowsExplicitClear(t *testing.T) {
	inputExternal := map[string]string{"form/name": "hello"}

	if !ShouldSyncExternalInput(inputExternal, "form/name", "", true) {
		t.Fatal("ShouldSyncExternalInput() returned false for explicit empty external value")
	}
}

func TestShouldSyncExternalInputIgnoresStableExternalValue(t *testing.T) {
	inputExternal := map[string]string{"form/name": "hello"}

	if ShouldSyncExternalInput(inputExternal, "form/name", "hello", true) {
		t.Fatal("ShouldSyncExternalInput() returned true for unchanged external value")
	}
}

func TestShouldSyncExternalInputAllowsFirstControlledValue(t *testing.T) {
	if !ShouldSyncExternalInput(map[string]string{}, "form/name", "hello", true) {
		t.Fatal("ShouldSyncExternalInput() returned false for first explicit external value")
	}
}
