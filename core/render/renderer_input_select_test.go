package render

import "testing"

func TestResolveSelectModelUsesOptionValueAndSelectedFlag(t *testing.T) {
	path := "/form/role"
	inputValues := map[string]string{}
	props := map[string]any{
		"options": []any{
			map[string]any{"label": "Choose role", "value": "", "selected": true},
			map[string]any{"label": "Reader", "value": "reader"},
			map[string]any{"label": "Owner", "value": "owner", "disabled": true},
		},
	}

	model := ResolveSelectModel(path, props, inputValues)
	if model.SelectedLabel != "Choose role" {
		t.Fatalf("SelectedLabel = %q, want %q", model.SelectedLabel, "Choose role")
	}
	if model.SelectedValue != "" {
		t.Fatalf("SelectedValue = %q, want empty", model.SelectedValue)
	}
	if got := inputValues[path]; got != "" {
		t.Fatalf("inputValues[path] = %q, want empty", got)
	}
	if len(model.EnabledIndexes) != 2 || model.EnabledIndexes[0] != 0 || model.EnabledIndexes[1] != 1 {
		t.Fatalf("EnabledIndexes = %#v, want [0 1]", model.EnabledIndexes)
	}
}

func TestResolveSelectModelSkipsDisabledSelectedOption(t *testing.T) {
	path := "/form/role"
	inputValues := map[string]string{}
	props := map[string]any{
		"options": []any{
			map[string]any{"label": "Owner", "value": "owner", "selected": true, "disabled": true},
			map[string]any{"label": "Reader", "value": "reader"},
		},
	}

	model := ResolveSelectModel(path, props, inputValues)
	if model.Index != 1 {
		t.Fatalf("Index = %d, want %d", model.Index, 1)
	}
	if model.SelectedValue != "reader" {
		t.Fatalf("SelectedValue = %q, want %q", model.SelectedValue, "reader")
	}
	if got := inputValues[path]; got != "reader" {
		t.Fatalf("inputValues[path] = %q, want %q", got, "reader")
	}
}

func TestResolveSelectModelInitialValueMatchesOptionValue(t *testing.T) {
	path := "/form/role"
	inputValues := map[string]string{}
	props := map[string]any{
		"value": "owner",
		"options": []any{
			map[string]any{"label": "Choose role", "value": ""},
			map[string]any{"label": "Reader role", "value": "reader"},
			map[string]any{"label": "Owner role", "value": "owner"},
		},
	}

	model := ResolveSelectModel(path, props, inputValues)
	if model.SelectedLabel != "Owner role" {
		t.Fatalf("SelectedLabel = %q, want %q", model.SelectedLabel, "Owner role")
	}
	if model.SelectedValue != "owner" {
		t.Fatalf("SelectedValue = %q, want %q", model.SelectedValue, "owner")
	}
	if got := inputValues[path]; got != "owner" {
		t.Fatalf("inputValues[path] = %q, want %q", got, "owner")
	}
}

func TestResolveSelectModelMarksSelectedDisabledWhenNoEnabledOptions(t *testing.T) {
	path := "/form/role"
	inputValues := map[string]string{}
	props := map[string]any{
		"options": []any{
			map[string]any{"label": "Locked", "value": "locked", "selected": true, "disabled": true},
		},
	}

	model := ResolveSelectModel(path, props, inputValues)
	if !model.SelectedDisabled {
		t.Fatalf("SelectedDisabled = %v, want true", model.SelectedDisabled)
	}
	if len(model.EnabledIndexes) != 0 {
		t.Fatalf("EnabledIndexes = %#v, want empty", model.EnabledIndexes)
	}
}

func TestResolveSelectModelRetainsOptionPropsForStyling(t *testing.T) {
	path := "/form/role"
	inputValues := map[string]string{}
	props := map[string]any{
		"options": []any{
			map[string]any{"label": "Reader", "value": "reader", "class": "role-reader", "style.color": "#22577a"},
		},
	}

	model := ResolveSelectModel(path, props, inputValues)
	if len(model.Entries) != 1 {
		t.Fatalf("Entries len = %d, want 1", len(model.Entries))
	}
	if got := model.Entries[0].Props["class"]; got != "role-reader" {
		t.Fatalf("option class = %#v, want role-reader", got)
	}
	if got := model.Entries[0].Props["style.color"]; got != "#22577a" {
		t.Fatalf("option style.color = %#v, want #22577a", got)
	}
}
