package render

import "testing"

func captureSubmitDispatch(state *GioWindowState, buttonPath string, buttonProps map[string]any) (bool, string, map[string]any) {
	called := false
	eventName := ""
	payload := map[string]any{}
	host := &GioRenderHost{
		DispatchEvent: func(name string, p map[string]any) error {
			called = true
			eventName = name
			payload = p
			return nil
		},
	}
	dispatchFormSubmitFromButton(state, host, buttonPath, buttonProps)
	return called, eventName, payload
}

func TestDispatchFormSubmitFromButtonIncludesNamedSubmitterInput(t *testing.T) {
	formPath := "/f"
	submitPath := "/f/submit"
	state := &GioWindowState{
		propsForPath: map[string]map[string]any{
			formPath:   {"tag": "form", "id": "main"},
			"/f/name":  {"tag": "input", "type": "text", "name": "username"},
			submitPath: {"tag": "input", "type": "submit", "name": "action", "value": "save"},
		},
		inputValues: map[string]string{"/f/name": "alice"},
		boolValues:  map[string]bool{},
	}

	called, eventName, payload := captureSubmitDispatch(state, submitPath, state.propsForPath[submitPath])
	if !called {
		t.Fatalf("submit event was not dispatched")
	}
	if eventName != "submit" {
		t.Fatalf("eventName = %q, want %q", eventName, "submit")
	}
	valuesRaw, ok := payload["values"]
	if !ok {
		t.Fatalf("payload missing values: %#v", payload)
	}
	values, ok := valuesRaw.(map[string]any)
	if !ok {
		t.Fatalf("payload values type = %T, want map[string]any", valuesRaw)
	}
	if got := values["username"]; got != "alice" {
		t.Fatalf("values[username] = %#v, want %q", got, "alice")
	}
	if got := values["action"]; got != "save" {
		t.Fatalf("values[action] = %#v, want %q", got, "save")
	}
}

func TestDispatchFormSubmitFromButtonIncludesDefaultSubmitButtonTag(t *testing.T) {
	formPath := "/f"
	submitPath := "/f/submit"
	state := &GioWindowState{
		propsForPath: map[string]map[string]any{
			formPath:   {"tag": "form", "id": "main"},
			submitPath: {"tag": "button", "name": "intent", "value": "publish"},
		},
		inputValues: map[string]string{},
		boolValues:  map[string]bool{},
	}

	called, _, payload := captureSubmitDispatch(state, submitPath, state.propsForPath[submitPath])
	if !called {
		t.Fatalf("submit event was not dispatched")
	}
	values, _ := payload["values"].(map[string]any)
	if got := values["intent"]; got != "publish" {
		t.Fatalf("values[intent] = %#v, want %q", got, "publish")
	}
}

func TestDispatchFormSubmitFromButtonSkipsUnnamedOrDisabledSubmitter(t *testing.T) {
	formPath := "/f"
	state := &GioWindowState{
		propsForPath: map[string]map[string]any{
			formPath:     {"tag": "form", "id": "main"},
			"/f/submit1": {"tag": "input", "type": "submit", "value": "go"},
			"/f/submit2": {"tag": "input", "type": "submit", "name": "intent", "value": "delete", "disabled": true},
		},
		inputValues: map[string]string{},
		boolValues:  map[string]bool{},
	}

	called1, _, payload1 := captureSubmitDispatch(state, "/f/submit1", state.propsForPath["/f/submit1"])
	if !called1 {
		t.Fatalf("submit event for unnamed submitter was not dispatched")
	}
	values1, _ := payload1["values"].(map[string]any)
	if _, ok := values1["intent"]; ok {
		t.Fatalf("unexpected submitter value for unnamed button: %#v", values1["intent"])
	}

	called2, _, payload2 := captureSubmitDispatch(state, "/f/submit2", state.propsForPath["/f/submit2"])
	if !called2 {
		t.Fatalf("submit event for disabled submitter was not dispatched")
	}
	values2, _ := payload2["values"].(map[string]any)
	if _, ok := values2["intent"]; ok {
		t.Fatalf("disabled submitter should not be included, got %#v", values2["intent"])
	}
}

func TestDispatchFormSubmitFromButtonRejectsTypeButton(t *testing.T) {
	formPath := "/f"
	buttonPath := "/f/btn"
	state := &GioWindowState{
		propsForPath: map[string]map[string]any{
			formPath:   {"tag": "form", "id": "main"},
			buttonPath: {"tag": "button", "type": "button", "name": "intent", "value": "noop"},
		},
		inputValues: map[string]string{},
		boolValues:  map[string]bool{},
	}

	called, _, _ := captureSubmitDispatch(state, buttonPath, state.propsForPath[buttonPath])
	if called {
		t.Fatalf("type=button should not dispatch submit")
	}
}

func TestIsSubmitButtonPropsButtonDefaultAndExplicitTypes(t *testing.T) {
	if !isSubmitButtonProps(map[string]any{"tag": "button"}) {
		t.Fatalf("button without type should be submit by default")
	}
	if !isSubmitButtonProps(map[string]any{"tag": "button", "type": "submit"}) {
		t.Fatalf("button type=submit should be submit")
	}
	if isSubmitButtonProps(map[string]any{"tag": "button", "type": "button"}) {
		t.Fatalf("button type=button should not be submit")
	}
}
