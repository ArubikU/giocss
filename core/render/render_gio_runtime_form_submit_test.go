package render

import "testing"

func TestCollectFormValuesSkipsDisabledAndUncheckedControls(t *testing.T) {
	formPath := "/root/form"
	state := &GioWindowState{
		propsForPath: map[string]map[string]any{
			formPath:                 {"tag": "form", "id": "main-form"},
			"/root/form/name":        {"tag": "input", "type": "text", "name": "username"},
			"/root/form/disabled":    {"tag": "input", "type": "text", "name": "skip", "disabled": true},
			"/root/form/cb1":         {"tag": "input", "type": "checkbox", "name": "topics", "value": "go"},
			"/root/form/cb2":         {"tag": "input", "type": "checkbox", "name": "topics", "value": "rust"},
			"/root/form/r1":          {"tag": "input", "type": "radio", "name": "role", "value": "admin"},
			"/root/form/r2":          {"tag": "input", "type": "radio", "name": "role", "value": "user"},
			"/root/form/country":     {"tag": "select", "name": "country", "options": []any{map[string]any{"label": "Argentina", "value": "ar"}, map[string]any{"label": "Mexico", "value": "mx", "selected": true}}},
			"/root/other/irrelevant": {"tag": "input", "type": "text", "name": "outside"},
		},
		inputValues: map[string]string{
			"/root/form/name":        "alice",
			"/root/form/disabled":    "nope",
			"radio:role":             "user",
			"/root/other/irrelevant": "outside",
		},
		boolValues: map[string]bool{
			"/root/form/cb1": true,
			"/root/form/cb2": false,
		},
	}

	values := collectFormValues(formPath, state)
	if got, ok := values["username"]; !ok || got != "alice" {
		t.Fatalf("values[username] = %#v, want %q", got, "alice")
	}
	if _, ok := values["skip"]; ok {
		t.Fatalf("disabled control should not be submitted, got %#v", values["skip"])
	}
	if got, ok := values["topics"]; !ok || got != "go" {
		t.Fatalf("values[topics] = %#v, want %q", got, "go")
	}
	if got, ok := values["role"]; !ok || got != "user" {
		t.Fatalf("values[role] = %#v, want %q", got, "user")
	}
	if got, ok := values["country"]; !ok || got != "mx" {
		t.Fatalf("values[country] = %#v, want %q", got, "mx")
	}
	if _, ok := values["outside"]; ok {
		t.Fatalf("outside control should not be submitted, got %#v", values["outside"])
	}
}

func TestCollectFormValuesAggregatesRepeatedNames(t *testing.T) {
	formPath := "/f"
	state := &GioWindowState{
		propsForPath: map[string]map[string]any{
			"/f/a": {"tag": "input", "type": "checkbox", "name": "tags", "value": "a"},
			"/f/b": {"tag": "input", "type": "checkbox", "name": "tags", "value": "b"},
		},
		inputValues: map[string]string{},
		boolValues: map[string]bool{
			"/f/a": true,
			"/f/b": true,
		},
	}

	values := collectFormValues(formPath, state)
	tagsRaw, ok := values["tags"]
	if !ok {
		t.Fatalf("values[tags] missing")
	}
	tags, ok := tagsRaw.([]any)
	if !ok {
		t.Fatalf("values[tags] type = %T, want []any", tagsRaw)
	}
	if len(tags) != 2 || tags[0] != "a" || tags[1] != "b" {
		t.Fatalf("values[tags] = %#v, want [a b]", tags)
	}
}
