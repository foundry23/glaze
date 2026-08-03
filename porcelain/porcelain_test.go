package porcelain

import (
	"slices"
	"testing"
)

func TestNewScannerDetectsTerminator(t *testing.T) {
	tests := []struct {
		name  string
		input string
		nul   bool
		lines []string
	}{
		{
			name:  "newline terminated",
			input: "one\ntwo\n",
			lines: []string{"one", "two"},
		},
		{
			name:  "nul terminated",
			input: "one\x00two\x00",
			nul:   true,
			lines: []string{"one", "two"},
		},
		{
			name:  "trailing blank line survives, newline form",
			input: "one\ntwo\n\n",
			lines: []string{"one", "two", ""},
		},
		{
			name:  "trailing blank line survives, nul form",
			input: "one\x00two\x00\x00",
			nul:   true,
			lines: []string{"one", "two", ""},
		},
		{
			name:  "unterminated final line",
			input: "one\ntwo",
			lines: []string{"one", "two"},
		},
		{
			name:  "empty input",
			input: "",
			lines: nil,
		},
		{
			name:  "newline inside a nul terminated record",
			input: "key\nvalue\x00",
			nul:   true,
			lines: []string{"key\nvalue"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newScanner([]byte(tc.input))
			if s.nul != tc.nul {
				t.Errorf("nul = %v, want %v", s.nul, tc.nul)
			}
			if !slices.Equal(s.lines, tc.lines) {
				t.Errorf("lines = %q, want %q", s.lines, tc.lines)
			}
		})
	}
}

func TestUnquote(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "octal escapes for utf-8",
			input: `"h\303\251llo w\303\266rld.txt"`,
			want:  "héllo wörld.txt",
		},
		{name: "quote and backslash", input: `"say \"hi\"\\done"`, want: `say "hi"\done`},
		{name: "control escapes", input: `"a\tb\nc\rd"`, want: "a\tb\nc\rd"},
		{name: "bell backspace formfeed vtab", input: `"\a\b\f\v"`, want: "\a\b\f\v"},
		{name: "no escapes", input: `"plain.txt"`, want: "plain.txt"},
		{name: "empty", input: `""`, want: ""},
		{name: "unterminated escape", input: `"trailing\`, wantErr: true},
		{name: "unknown escape", input: `"bad\q"`, wantErr: true},
		{name: "truncated octal", input: `"bad\30"`, wantErr: true},
		{name: "not quoted", input: `plain.txt`, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := unquote(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("unquote(%q) = %q, want error", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unquote(%q): %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("unquote(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
