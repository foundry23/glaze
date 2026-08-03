package main

import (
	"os"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "html characters are left alone",
			value: map[string]string{"path": "a<b>&c"},
			want:  "{\"path\":\"a<b>&c\"}\n",
		},
		{
			name:  "output is one line",
			value: []map[string]string{{"a": "1"}, {"b": "2"}},
			want:  "[{\"a\":\"1\"},{\"b\":\"2\"}]\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out strings.Builder
			if err := encode(&out, tc.value); err != nil {
				t.Fatalf("encode: %v", err)
			}
			if out.String() != tc.want {
				t.Errorf("encode = %q, want %q", out.String(), tc.want)
			}
		})
	}
}

func TestOriginsRegistry(t *testing.T) {
	if len(origins) == 0 {
		t.Fatal("no origins registered")
	}

	seen := map[string]bool{}
	for _, o := range origins {
		if o.name == "" {
			t.Errorf("origin %+v is missing a name", o)
		}
		if o.parse == nil {
			t.Errorf("origin %q has no parser", o.name)
		}
		if seen[o.name] {
			t.Errorf("duplicate origin name %q", o.name)
		}
		seen[o.name] = true
	}

	if !seen["git-worktree"] {
		t.Error("no origin registered for git-worktree")
	}
}

func TestLookup(t *testing.T) {
	t.Run("known origin", func(t *testing.T) {
		got, err := lookup("git-worktree")
		if err != nil {
			t.Fatalf("lookup: %v", err)
		}
		if got.name != "git-worktree" {
			t.Errorf("name = %q, want git-worktree", got.name)
		}
	})

	for _, tc := range []struct{ name, arg string }{
		{name: "missing origin", arg: ""},
		{name: "unknown origin", arg: "git-nonsense"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lookup(tc.arg)
			if err == nil {
				t.Fatalf("lookup(%q) succeeded, want error", tc.arg)
			}
			if !strings.Contains(err.Error(), "git-worktree") {
				t.Errorf("error %q does not list the valid origins", err)
			}
		})
	}
}

func TestReadFromStdin(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	go func() {
		_, _ = writer.WriteString("worktree /repo\x00HEAD abc\x00detached\x00\x00")
		_ = writer.Close()
	}()

	restore := swapStdin(t, reader)
	defer restore()

	got, err := read()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if want := "worktree /repo\x00HEAD abc\x00detached\x00\x00"; string(got) != want {
		t.Errorf("read = %q, want %q", got, want)
	}
}

func TestReadWithNoStdinErrors(t *testing.T) {
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("opening %s: %v", os.DevNull, err)
	}
	defer func() { _ = devNull.Close() }()

	restore := swapStdin(t, devNull)
	defer restore()

	_, err = read()
	if err == nil {
		t.Fatal("read succeeded with no input on stdin, want an error")
	}
	if !strings.Contains(err.Error(), "stdin") {
		t.Errorf("error %q does not mention stdin", err)
	}
}

func swapStdin(t *testing.T, file *os.File) func() {
	t.Helper()
	previous := os.Stdin
	os.Stdin = file
	return func() { os.Stdin = previous }
}
