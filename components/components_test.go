package components

import (
	"testing"

	giocss "github.com/ArubikU/giocss"
)

func TestSelectPreservesOptionValueMetadata(t *testing.T) {
	node := Select("role", []*giocss.Node{
		Option("Choose a role", "", true, false),
		Option("Reader", "reader", false, false),
		Option("Owner", "owner", false, true),
	})

	options, ok := node.GetProp("options").([]any)
	if !ok || len(options) != 3 {
		t.Fatalf("options = %#v, want 3 structured entries", node.GetProp("options"))
	}
	reader, ok := options[1].(map[string]any)
	if !ok {
		t.Fatalf("options[1] = %#v, want map[string]any", options[1])
	}
	if reader["label"] != "Reader" || reader["value"] != "reader" {
		t.Fatalf("reader option = %#v, want label/value preserved", reader)
	}
	if reader["tag"] != "option" {
		t.Fatalf("reader option tag = %#v, want option", reader["tag"])
	}
	owner, ok := options[2].(map[string]any)
	if !ok || owner["disabled"] != true {
		t.Fatalf("owner option = %#v, want disabled metadata", options[2])
	}
}
