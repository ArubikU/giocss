package render

import "testing"

func TestResolveInvalidStateRequiredSelectWithDisabledSelection(t *testing.T) {
	path := "/form/role"
	state := &GioWindowState{inputValues: map[string]string{}}
	props := map[string]any{
		"tag":      "select",
		"required": true,
		"options": []any{
			map[string]any{"label": "Locked", "value": "locked", "selected": true, "disabled": true},
		},
	}

	if !resolveInvalidState(path, props, state, false, false) {
		t.Fatalf("resolveInvalidState() = false, want true")
	}
}
