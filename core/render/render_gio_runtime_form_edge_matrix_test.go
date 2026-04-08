package render

import (
	"reflect"
	"testing"
)

func testFormState(props map[string]map[string]any, input map[string]string, bools map[string]bool) *GioWindowState {
	if props == nil {
		props = map[string]map[string]any{}
	}
	if input == nil {
		input = map[string]string{}
	}
	if bools == nil {
		bools = map[string]bool{}
	}
	return &GioWindowState{
		propsForPath: props,
		inputValues:  input,
		boolValues:   bools,
	}
}

func TestIsSubmittableControlPropsMatrix(t *testing.T) {
	cases := []struct {
		name  string
		props map[string]any
		want  bool
	}{
		{name: "nil props", props: nil, want: false},
		{name: "input text", props: map[string]any{"tag": "input", "type": "text"}, want: true},
		{name: "input email", props: map[string]any{"tag": "input", "type": "email"}, want: true},
		{name: "input number", props: map[string]any{"tag": "input", "type": "number"}, want: true},
		{name: "input hidden", props: map[string]any{"tag": "input", "type": "hidden"}, want: true},
		{name: "input checkbox", props: map[string]any{"tag": "input", "type": "checkbox"}, want: true},
		{name: "input check alias", props: map[string]any{"tag": "input", "type": "check"}, want: true},
		{name: "input radio", props: map[string]any{"tag": "input", "type": "radio"}, want: true},
		{name: "input submit", props: map[string]any{"tag": "input", "type": "submit"}, want: false},
		{name: "input reset", props: map[string]any{"tag": "input", "type": "reset"}, want: false},
		{name: "input button", props: map[string]any{"tag": "input", "type": "button"}, want: false},
		{name: "input image", props: map[string]any{"tag": "input", "type": "image"}, want: false},
		{name: "input file", props: map[string]any{"tag": "input", "type": "file"}, want: false},
		{name: "select tag", props: map[string]any{"tag": "select"}, want: true},
		{name: "textarea tag", props: map[string]any{"tag": "textarea"}, want: true},
		{name: "component select", props: map[string]any{"component": "select"}, want: true},
		{name: "component dropdown", props: map[string]any{"component": "dropdown"}, want: true},
		{name: "component textarea", props: map[string]any{"component": "textarea"}, want: true},
		{name: "disabled bool", props: map[string]any{"tag": "input", "type": "text", "disabled": true}, want: false},
		{name: "disabled attr string", props: map[string]any{"tag": "input", "type": "text", "disabled": "disabled"}, want: false},
		{name: "disabled zero", props: map[string]any{"tag": "input", "type": "text", "disabled": 0}, want: true},
		{name: "button tag submit", props: map[string]any{"tag": "button", "type": "submit"}, want: false},
		{name: "div not control", props: map[string]any{"tag": "div"}, want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := isSubmittableControlProps(tc.props)
			if got != tc.want {
				t.Fatalf("isSubmittableControlProps() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestResolveSubmittableControlValueMatrix(t *testing.T) {
	cases := []struct {
		name        string
		path        string
		props       map[string]any
		state       *GioWindowState
		wantValue   any
		wantInclude bool
	}{
		{
			name:      "text from state",
			path:      "/f/name",
			props:     map[string]any{"tag": "input", "type": "text"},
			state:     testFormState(nil, map[string]string{"/f/name": "alice"}, nil),
			wantValue: "alice", wantInclude: true,
		},
		{
			name:      "text from props",
			path:      "/f/name",
			props:     map[string]any{"tag": "input", "type": "text", "value": "bob"},
			state:     testFormState(nil, nil, nil),
			wantValue: "bob", wantInclude: true,
		},
		{
			name:      "text trimmed",
			path:      "/f/name",
			props:     map[string]any{"tag": "input", "type": "text"},
			state:     testFormState(nil, map[string]string{"/f/name": "  jane  "}, nil),
			wantValue: "jane", wantInclude: true,
		},
		{
			name:      "textarea from state",
			path:      "/f/msg",
			props:     map[string]any{"tag": "textarea"},
			state:     testFormState(nil, map[string]string{"/f/msg": "hello"}, nil),
			wantValue: "hello", wantInclude: true,
		},
		{
			name:      "textarea component",
			path:      "/f/msg",
			props:     map[string]any{"component": "textarea", "value": "fallback"},
			state:     testFormState(nil, nil, nil),
			wantValue: "fallback", wantInclude: true,
		},
		{
			name:      "checkbox checked bool map",
			path:      "/f/cb",
			props:     map[string]any{"tag": "input", "type": "checkbox", "value": "go"},
			state:     testFormState(nil, nil, map[string]bool{"/f/cb": true}),
			wantValue: "go", wantInclude: true,
		},
		{
			name:      "checkbox checked attr",
			path:      "/f/cb",
			props:     map[string]any{"tag": "input", "type": "checkbox", "checked": true},
			state:     testFormState(nil, nil, nil),
			wantValue: "on", wantInclude: true,
		},
		{
			name:      "checkbox unchecked bool map",
			path:      "/f/cb",
			props:     map[string]any{"tag": "input", "type": "checkbox", "value": "go"},
			state:     testFormState(nil, nil, map[string]bool{"/f/cb": false}),
			wantValue: nil, wantInclude: false,
		},
		{
			name:      "checkbox unchecked by default",
			path:      "/f/cb",
			props:     map[string]any{"tag": "input", "type": "checkbox"},
			state:     testFormState(nil, nil, nil),
			wantValue: nil, wantInclude: false,
		},
		{
			name:      "checkbox empty value yields on",
			path:      "/f/cb",
			props:     map[string]any{"tag": "input", "type": "checkbox", "value": ""},
			state:     testFormState(nil, nil, map[string]bool{"/f/cb": true}),
			wantValue: "on", wantInclude: true,
		},
		{
			name:      "radio selected by group",
			path:      "/f/r1",
			props:     map[string]any{"tag": "input", "type": "radio", "name": "role", "value": "admin"},
			state:     testFormState(nil, map[string]string{"radio:role": "admin"}, nil),
			wantValue: "admin", wantInclude: true,
		},
		{
			name:      "radio not selected",
			path:      "/f/r1",
			props:     map[string]any{"tag": "input", "type": "radio", "name": "role", "value": "admin"},
			state:     testFormState(nil, map[string]string{"radio:role": "user"}, nil),
			wantValue: nil, wantInclude: false,
		},
		{
			name:      "radio checked attr",
			path:      "/f/r1",
			props:     map[string]any{"tag": "input", "type": "radio", "name": "role", "value": "admin", "checked": true},
			state:     testFormState(nil, nil, nil),
			wantValue: "admin", wantInclude: true,
		},
		{
			name:      "radio empty value yields on",
			path:      "/f/r1",
			props:     map[string]any{"tag": "input", "type": "radio", "name": "role", "value": "", "checked": true},
			state:     testFormState(nil, nil, nil),
			wantValue: "on", wantInclude: true,
		},
		{
			name:      "select with nil state uses prop value",
			path:      "/f/sel",
			props:     map[string]any{"tag": "select", "value": "mx"},
			state:     nil,
			wantValue: "mx", wantInclude: true,
		},
		{
			name: "select uses selected option",
			path: "/f/sel",
			props: map[string]any{
				"tag":     "select",
				"options": []any{map[string]any{"value": "ar"}, map[string]any{"value": "mx", "selected": true}},
			},
			state:     testFormState(nil, nil, nil),
			wantValue: "mx", wantInclude: true,
		},
		{
			name: "select selected disabled falls back to enabled option",
			path: "/f/sel",
			props: map[string]any{
				"tag":     "select",
				"options": []any{map[string]any{"value": "ar"}, map[string]any{"value": "mx", "selected": true, "disabled": true}},
			},
			state:     testFormState(nil, nil, nil),
			wantValue: "ar", wantInclude: true,
		},
		{
			name:      "dropdown component treated as select",
			path:      "/f/dd",
			props:     map[string]any{"component": "dropdown", "value": "v1"},
			state:     nil,
			wantValue: "v1", wantInclude: true,
		},
		{
			name:      "unknown tag excluded",
			path:      "/f/x",
			props:     map[string]any{"tag": "div", "value": "x"},
			state:     testFormState(nil, nil, nil),
			wantValue: nil, wantInclude: false,
		},
		{
			name:      "input empty string is still included",
			path:      "/f/name",
			props:     map[string]any{"tag": "input", "type": "text"},
			state:     testFormState(nil, map[string]string{"/f/name": ""}, nil),
			wantValue: "", wantInclude: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gotValue, gotInclude := resolveSubmittableControlValue(tc.path, tc.props, tc.state)
			if gotInclude != tc.wantInclude {
				t.Fatalf("include = %v, want %v", gotInclude, tc.wantInclude)
			}
			if !reflect.DeepEqual(gotValue, tc.wantValue) {
				t.Fatalf("value = %#v, want %#v", gotValue, tc.wantValue)
			}
		})
	}
}

func TestCollectFormValuesEdgeMatrix(t *testing.T) {
	cases := []struct {
		name     string
		formPath string
		state    *GioWindowState
		want     map[string]any
	}{
		{
			name:     "nil state",
			formPath: "/f",
			state:    nil,
			want:     map[string]any{},
		},
		{
			name:     "empty form path",
			formPath: "",
			state:    testFormState(map[string]map[string]any{"/f": {"tag": "form"}}, nil, nil),
			want:     map[string]any{},
		},
		{
			name:     "nil props map",
			formPath: "/f",
			state:    &GioWindowState{propsForPath: nil, inputValues: map[string]string{}, boolValues: map[string]bool{}},
			want:     map[string]any{},
		},
		{
			name:     "basic text descendant",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{"/f": {"tag": "form"}, "/f/name": {"tag": "input", "type": "text", "name": "username"}},
				map[string]string{"/f/name": "alice"},
				nil,
			),
			want: map[string]any{"username": "alice"},
		},
		{
			name:     "disabled descendant skipped",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{"/f": {"tag": "form"}, "/f/name": {"tag": "input", "type": "text", "name": "username", "disabled": true}},
				map[string]string{"/f/name": "alice"},
				nil,
			),
			want: map[string]any{},
		},
		{
			name:     "unchecked checkbox skipped",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{"/f": {"tag": "form"}, "/f/cb": {"tag": "input", "type": "checkbox", "name": "agree"}},
				nil,
				map[string]bool{"/f/cb": false},
			),
			want: map[string]any{},
		},
		{
			name:     "checked checkbox default on",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{"/f": {"tag": "form"}, "/f/cb": {"tag": "input", "type": "checkbox", "name": "agree"}},
				nil,
				map[string]bool{"/f/cb": true},
			),
			want: map[string]any{"agree": "on"},
		},
		{
			name:     "checked checkbox custom value",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{"/f": {"tag": "form"}, "/f/cb": {"tag": "input", "type": "checkbox", "name": "topic", "value": "go"}},
				nil,
				map[string]bool{"/f/cb": true},
			),
			want: map[string]any{"topic": "go"},
		},
		{
			name:     "radio selected by state",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":    {"tag": "form"},
					"/f/r1": {"tag": "input", "type": "radio", "name": "role", "value": "admin"},
					"/f/r2": {"tag": "input", "type": "radio", "name": "role", "value": "user"},
				},
				map[string]string{"radio:role": "user"},
				nil,
			),
			want: map[string]any{"role": "user"},
		},
		{
			name:     "radio none selected",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{"/f": {"tag": "form"}, "/f/r1": {"tag": "input", "type": "radio", "name": "role", "value": "admin"}},
				map[string]string{"radio:role": "user"},
				nil,
			),
			want: map[string]any{},
		},
		{
			name:     "repeated name aggregates",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":   {"tag": "form"},
					"/f/a": {"tag": "input", "type": "checkbox", "name": "tags", "value": "a"},
					"/f/b": {"tag": "input", "type": "checkbox", "name": "tags", "value": "b"},
				},
				nil,
				map[string]bool{"/f/a": true, "/f/b": true},
			),
			want: map[string]any{"tags": []any{"a", "b"}},
		},
		{
			name:     "outside descendant skipped",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":      {"tag": "form"},
					"/g/name": {"tag": "input", "type": "text", "name": "outside"},
				},
				map[string]string{"/g/name": "x"},
				nil,
			),
			want: map[string]any{},
		},
		{
			name:     "external control via form attribute",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":   {"tag": "form", "id": "main"},
					"/ext": {"tag": "input", "type": "text", "name": "external", "form": "main"},
				},
				map[string]string{"/ext": "value"},
				nil,
			),
			want: map[string]any{"external": "value"},
		},
		{
			name:     "external wrong form id skipped",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":   {"tag": "form", "id": "main"},
					"/ext": {"tag": "input", "type": "text", "name": "external", "form": "other"},
				},
				map[string]string{"/ext": "value"},
				nil,
			),
			want: map[string]any{},
		},
		{
			name:     "external via form-id attribute",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":   {"tag": "form", "id": "main"},
					"/ext": {"tag": "textarea", "name": "external", "form-id": "main"},
				},
				map[string]string{"/ext": "hello"},
				nil,
			),
			want: map[string]any{"external": "hello"},
		},
		{
			name:     "file input excluded",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{"/f": {"tag": "form"}, "/f/file": {"tag": "input", "type": "file", "name": "upload"}},
				nil,
				nil,
			),
			want: map[string]any{},
		},
		{
			name:     "fieldset disabled excludes descendants",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":         {"tag": "form"},
					"/f/fs":      {"tag": "fieldset", "disabled": true},
					"/f/fs/name": {"tag": "input", "type": "text", "name": "username"},
				},
				map[string]string{"/f/fs/name": "alice"},
				nil,
			),
			want: map[string]any{},
		},
		{
			name:     "nested disabled fieldset excludes descendants",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":              {"tag": "form"},
					"/f/fs1":          {"tag": "fieldset", "disabled": true},
					"/f/fs1/fs2":      {"tag": "fieldset"},
					"/f/fs1/fs2/name": {"tag": "input", "type": "text", "name": "username"},
				},
				map[string]string{"/f/fs1/fs2/name": "alice"},
				nil,
			),
			want: map[string]any{},
		},
		{
			name:     "external in disabled fieldset excluded",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":        {"tag": "form", "id": "main"},
					"/wrap":     {"tag": "fieldset", "disabled": true},
					"/wrap/ext": {"tag": "input", "type": "text", "name": "external", "form": "main"},
				},
				map[string]string{"/wrap/ext": "value"},
				nil,
			),
			want: map[string]any{},
		},
		{
			name:     "enabled fieldset allows descendants",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":         {"tag": "form"},
					"/f/fs":      {"tag": "fieldset", "disabled": false},
					"/f/fs/name": {"tag": "input", "type": "text", "name": "username"},
				},
				map[string]string{"/f/fs/name": "alice"},
				nil,
			),
			want: map[string]any{"username": "alice"},
		},
		{
			name:     "disabled fieldset keeps first legend descendants enabled",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":                 {"tag": "form"},
					"/f/fs":              {"tag": "fieldset", "disabled": true},
					"/f/fs/legendA":      {"tag": "legend"},
					"/f/fs/legendA/name": {"tag": "input", "type": "text", "name": "username"},
				},
				map[string]string{"/f/fs/legendA/name": "alice"},
				nil,
			),
			want: map[string]any{"username": "alice"},
		},
		{
			name:     "disabled fieldset excludes second legend descendants",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":                 {"tag": "form"},
					"/f/fs":              {"tag": "fieldset", "disabled": true},
					"/f/fs/legendA":      {"tag": "legend"},
					"/f/fs/legendA/ok":   {"tag": "input", "type": "text", "name": "okName"},
					"/f/fs/legendB":      {"tag": "legend"},
					"/f/fs/legendB/nope": {"tag": "input", "type": "text", "name": "nopeName"},
				},
				map[string]string{"/f/fs/legendA/ok": "ok", "/f/fs/legendB/nope": "nope"},
				nil,
			),
			want: map[string]any{"okName": "ok"},
		},
		{
			name:     "disabled fieldset excludes non legend descendants when first legend exists",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":            {"tag": "form"},
					"/f/fs":         {"tag": "fieldset", "disabled": true},
					"/f/fs/legendA": {"tag": "legend"},
					"/f/fs/direct":  {"tag": "input", "type": "text", "name": "blocked"},
				},
				map[string]string{"/f/fs/direct": "x"},
				nil,
			),
			want: map[string]any{},
		},
		{
			name:     "mixed descendant and external",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":      {"tag": "form", "id": "main"},
					"/f/name": {"tag": "input", "type": "text", "name": "username"},
					"/ext":    {"tag": "textarea", "name": "bio", "form": "main"},
				},
				map[string]string{"/f/name": "alice", "/ext": "dev"},
				nil,
			),
			want: map[string]any{"username": "alice", "bio": "dev"},
		},
		{
			name:     "selected disabled option falls back to enabled",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{
					"/f":     {"tag": "form"},
					"/f/sel": {"tag": "select", "name": "country", "options": []any{map[string]any{"value": "ar"}, map[string]any{"value": "mx", "selected": true, "disabled": true}}},
				},
				nil,
				nil,
			),
			want: map[string]any{"country": "ar"},
		},
		{
			name:     "id fallback for missing name",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{"/f": {"tag": "form"}, "/f/name": {"tag": "input", "type": "text", "id": "userId"}},
				map[string]string{"/f/name": "alice"},
				nil,
			),
			want: map[string]any{"userId": "alice"},
		},
		{
			name:     "empty name is skipped",
			formPath: "/f",
			state: testFormState(
				map[string]map[string]any{"/f": {"tag": "form"}, "/f/name": {"tag": "input", "type": "text", "name": ""}},
				map[string]string{"/f/name": "alice"},
				nil,
			),
			want: map[string]any{},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := collectFormValues(tc.formPath, tc.state)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("collectFormValues() = %#v, want %#v", got, tc.want)
			}
		})
	}
}
