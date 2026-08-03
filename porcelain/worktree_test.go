package porcelain

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseWorktrees(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Worktree
	}{
		{
			name: "newline form, branch detached and locked",
			input: "worktree /repo\x0aHEAD 1a568ef629f75052cb03848eca98c3f1abd19d23\x0abranch refs/heads/main\x0a" +
				"\x0aworktree /wt-detached\x0aHEAD 1a568ef629f75052cb03848eca98c3f1abd19d23\x0adetached\x0a" +
				"\x0aworktree /wt-feature\x0aHEAD 1a568ef629f75052cb03848eca98c3f1abd19d23\x0abranch refs/heads/feature\x0alocked held for testing\x0a\x0a",
			want: []Worktree{
				{Path: "/repo", Head: "1a568ef629f75052cb03848eca98c3f1abd19d23", Branch: "refs/heads/main", BranchName: "main"},
				{Path: "/wt-detached", Head: "1a568ef629f75052cb03848eca98c3f1abd19d23", Detached: true},
				{
					Path: "/wt-feature", Head: "1a568ef629f75052cb03848eca98c3f1abd19d23",
					Branch: "refs/heads/feature", BranchName: "feature",
					Locked: true, LockReason: "held for testing",
				},
			},
		},
		{
			name: "nul form",
			input: "worktree /repo\x00HEAD abc\x00branch refs/heads/main\x00" +
				"\x00worktree /wt-detached\x00HEAD abc\x00detached\x00\x00",
			want: []Worktree{
				{Path: "/repo", Head: "abc", Branch: "refs/heads/main", BranchName: "main"},
				{Path: "/wt-detached", Head: "abc", Detached: true},
			},
		},
		{
			name:  "bare repository",
			input: "worktree /bare.git\x0abare\x0a\x0a",
			want:  []Worktree{{Path: "/bare.git", Bare: true}},
		},
		{
			name:  "locked with no reason",
			input: "worktree /wt\x0aHEAD abc\x0adetached\x0alocked\x0a\x0a",
			want:  []Worktree{{Path: "/wt", Head: "abc", Detached: true, Locked: true}},
		},
		{
			name:  "prunable with reason",
			input: "worktree /gone\x0aHEAD abc\x0adetached\x0aprunable gitdir file points to non-existent location\x0a\x0a",
			want: []Worktree{{
				Path: "/gone", Head: "abc", Detached: true,
				Prunable: true, PruneReason: "gitdir file points to non-existent location",
			}},
		},
		{
			name:  "unterminated final record",
			input: "worktree /repo\x0aHEAD abc\x0abranch refs/heads/main\x0a",
			want:  []Worktree{{Path: "/repo", Head: "abc", Branch: "refs/heads/main", BranchName: "main"}},
		},
		{
			name:  "quoted path",
			input: "worktree \"/tmp/h\\303\\251llo\"\x0aHEAD abc\x0adetached\x0a\x0a",
			want:  []Worktree{{Path: "/tmp/héllo", Head: "abc", Detached: true}},
		},
		{
			name:  "path containing a newline, nul form",
			input: "worktree /tmp/two\nlines\x00HEAD abc\x00detached\x00\x00",
			want:  []Worktree{{Path: "/tmp/two\nlines", Head: "abc", Detached: true}},
		},
		{
			name:  "unknown annotation is ignored",
			input: "worktree /repo\x0aHEAD abc\x0abranch refs/heads/main\x0asomething-new yes\x0a\x0a",
			want:  []Worktree{{Path: "/repo", Head: "abc", Branch: "refs/heads/main", BranchName: "main"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseWorktrees([]byte(tc.input))
			if err != nil {
				t.Fatalf("ParseWorktrees: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseWorktrees =\n%+v\nwant\n%+v", got, tc.want)
			}
		})
	}
}

func TestParseWorktreesRejectsNonPorcelain(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantContains string
	}{
		{
			name: "plain git worktree list output",
			input: "/repo         1a568ef [main]\x0a" +
				"/wt-detached  1a568ef (detached HEAD)\x0a" +
				"/wt-feature   1a568ef [feature] locked\x0a",
			wantContains: "input format not as expected",
		},
		{
			name:         "a different porcelain format entirely",
			input:        "A  .gitignore\x0a M a.txt\x0a",
			wantContains: "input format not as expected",
		},
		{
			name:         "arbitrary text",
			input:        "hello world\x0a",
			wantContains: "input format not as expected",
		},
		{
			name:         "json",
			input:        "{\"a\":1}\x0a",
			wantContains: "input format not as expected",
		},
		{
			name:         "empty input",
			input:        "",
			wantContains: "no worktrees found",
		},
		{
			name:         "whitespace only",
			input:        "\x0a\x0a",
			wantContains: "no worktrees found",
		},
		{
			name:         "key after a complete record",
			input:        "worktree /repo\x0aHEAD abc\x0adetached\x0a\x0aHEAD def\x0a",
			wantContains: "input format not as expected",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseWorktrees([]byte(tc.input))
			if err == nil {
				t.Fatalf("ParseWorktrees(%q) = %+v, want error", tc.input, got)
			}
			if !strings.Contains(err.Error(), tc.wantContains) {
				t.Errorf("error %q does not contain %q", err, tc.wantContains)
			}
			if !strings.HasPrefix(err.Error(), "worktree: ") {
				t.Errorf("error %q is not prefixed \"worktree: \"", err)
			}
		})
	}
}

func TestParseWorktreesErrorTruncatesLongLines(t *testing.T) {
	line := "/" + strings.Repeat("very-long-path-segment/", 40) + "repo  abc [main]"
	_, err := ParseWorktrees([]byte(line + "\x0a"))
	if err == nil {
		t.Fatal("expected an error")
	}
	if len(err.Error()) > 200 {
		t.Errorf("error is %d characters, want it truncated:\n%s", len(err.Error()), err)
	}
	if !strings.Contains(err.Error(), "…") {
		t.Errorf("error %q does not show truncation", err)
	}
}
