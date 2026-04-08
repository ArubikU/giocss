package engine

import "testing"

func TestCycleSelectStateSkipsDisabledAndUsesOptionValue(t *testing.T) {
	path := "/form/role"
	key := "sel:" + path
	inputValues := map[string]string{}
	labels := []string{"Choose role", "Reader", "Owner"}
	values := []string{"", "reader", "owner"}
	enabled := []int{0, 1}

	payload, ok := CycleSelectState(path, key, inputValues, labels, values, enabled, 0, 123)
	if !ok {
		t.Fatalf("CycleSelectState() = not ok")
	}
	if got := payload["value"]; got != "reader" {
		t.Fatalf("payload[value] = %#v, want %q", got, "reader")
	}
	if got := payload["index"]; got != 1 {
		t.Fatalf("payload[index] = %#v, want %d", got, 1)
	}

	payload, ok = CycleSelectState(path, key, inputValues, labels, values, enabled, 1, 124)
	if !ok {
		t.Fatalf("CycleSelectState() second call = not ok")
	}
	if got := payload["index"]; got != 0 {
		t.Fatalf("payload[index] second = %#v, want %d", got, 0)
	}
	if got := payload["value"]; got != "Choose role" {
		t.Fatalf("payload[value] second = %#v, want %q", got, "Choose role")
	}
}

func TestCycleSelectStateRepairsStoredDisabledIndex(t *testing.T) {
	path := "/form/role"
	key := "sel:" + path
	inputValues := map[string]string{key: "2"}
	labels := []string{"Choose role", "Reader", "Owner"}
	values := []string{"", "reader", "owner"}
	enabled := []int{0, 1}

	payload, ok := CycleSelectState(path, key, inputValues, labels, values, enabled, 0, 125)
	if !ok {
		t.Fatalf("CycleSelectState() = not ok")
	}
	if got := payload["index"]; got != 0 {
		t.Fatalf("payload[index] = %#v, want %d", got, 0)
	}
}

func TestCycleSelectStateNoEnabledOptionsReturnsFalse(t *testing.T) {
	path := "/form/role"
	key := "sel:" + path
	inputValues := map[string]string{}
	labels := []string{"Locked"}
	values := []string{"locked"}
	enabled := []int{}

	payload, ok := CycleSelectState(path, key, inputValues, labels, values, enabled, 0, 126)
	if ok {
		t.Fatalf("CycleSelectState() ok = true, want false with payload %#v", payload)
	}
}
