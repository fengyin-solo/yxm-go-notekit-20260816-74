package validator

import "testing"

func TestTrimSpace(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"  hello  ", "hello"},
		{"\tfoo\n", "foo"},
		{"  ", ""},
		{"no-space", "no-space"},
		{"a", "a"},
	}
	for _, c := range cases {
		got := trimSpace(c.input)
		if got != c.want {
			t.Errorf("trimSpace(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestCreateNotebookValidate(t *testing.T) {
	ve := CreateNotebookRequest{Name: ""}.Validate()
	if !ve.HasErrors() {
		t.Error("empty name should fail validation")
	}
	ve = CreateNotebookRequest{Name: "  My Notebook  "}.Validate()
	if ve.HasErrors() {
		t.Errorf("unexpected errors: %v", ve)
	}
}

func TestCreateNoteValidate(t *testing.T) {
	ve := CreateNoteRequest{NotebookID: "", Title: ""}.Validate()
	if len(ve) < 2 {
		t.Errorf("expected at least 2 errors, got %d", len(ve))
	}
	ve = CreateNoteRequest{NotebookID: "nb1", Title: "Note title"}.Validate()
	if ve.HasErrors() {
		t.Errorf("unexpected errors: %v", ve)
	}
}
