//go:build integration

package porcelain

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func repo(t *testing.T) (func(args ...string) []byte, string) {
	t.Helper()
	dir := t.TempDir()

	git := func(args ...string) []byte {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		// Isolate from the developer's own git configuration.
		cmd.Env = append(cmd.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
			"GIT_AUTHOR_DATE=2026-01-02T03:04:05+00:00",
			"GIT_COMMITTER_DATE=2026-01-02T03:04:05+00:00",
		)
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("git %s: %v", strings.Join(args, " "), err)
		}
		return out
	}

	git("init", "-q", "-b", "main", ".")
	git("config", "user.name", "Test Dev")
	git("config", "user.email", "dev@example.com")
	write(t, filepath.Join(dir, "a.txt"), "line one\nline two\n")
	git("add", "-A")
	git("commit", "-q", "-m", "initial commit")
	return git, dir
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

func TestIntegrationWorktrees(t *testing.T) {
	git, dir := repo(t)

	git("worktree", "add", "-q", filepath.Join(dir, "..", "wt-feature"), "-b", "feature")
	git("worktree", "add", "-q", "--detach", filepath.Join(dir, "..", "wt-detached"), "HEAD")
	git("worktree", "lock", filepath.Join(dir, "..", "wt-feature"), "--reason", "held for testing")

	// Both delimiter forms must produce the same result.
	for _, args := range [][]string{
		{"worktree", "list", "--porcelain"},
		{"worktree", "list", "--porcelain", "-z"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			got, err := ParseWorktrees(git(args...))
			if err != nil {
				t.Fatalf("ParseWorktrees: %v", err)
			}
			if len(got) != 3 {
				t.Fatalf("got %d worktrees, want 3: %+v", len(got), got)
			}

			var main, detached, locked *Worktree
			for i := range got {
				switch {
				case got[i].BranchName == "main":
					main = &got[i]
				case got[i].Detached:
					detached = &got[i]
				case got[i].BranchName == "feature":
					locked = &got[i]
				}
			}
			if main == nil || main.Branch != "refs/heads/main" || main.Head == "" {
				t.Errorf("main worktree = %+v", main)
			}
			if detached == nil || detached.Branch != "" || detached.BranchName != "" {
				t.Errorf("detached worktree = %+v", detached)
			}
			if locked == nil || !locked.Locked || locked.LockReason != "held for testing" {
				t.Errorf("locked worktree = %+v", locked)
			}
			for _, worktree := range got {
				if worktree.Bare {
					t.Errorf("worktree %s reported bare", worktree.Path)
				}
			}
		})
	}
}

func TestIntegrationWorktreesBare(t *testing.T) {
	// A bare repository reports no HEAD and no branch, only the bare marker.
	dir := t.TempDir()
	bare := filepath.Join(dir, "bare.git")
	if out, err := exec.Command("git", "init", "-q", "--bare", bare).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v: %s", err, out)
	}
	out, err := exec.Command("git", "-C", bare, "worktree", "list", "--porcelain").Output()
	if err != nil {
		t.Fatalf("git worktree list: %v", err)
	}

	got, err := ParseWorktrees(out)
	if err != nil {
		t.Fatalf("ParseWorktrees: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d worktrees, want 1: %+v", len(got), got)
	}
	if !got[0].Bare || got[0].Head != "" || got[0].Branch != "" {
		t.Errorf("bare worktree = %+v, want bare with no head or branch", got[0])
	}
}

func TestIntegrationWorktreesPrunable(t *testing.T) {
	git, dir := repo(t)
	gone := filepath.Join(dir, "..", "wt-gone")
	git("worktree", "add", "-q", "--detach", gone, "HEAD")

	// Deleting a worktree's directory behind git's back makes it prunable.
	if err := os.RemoveAll(gone); err != nil {
		t.Fatalf("removing worktree directory: %v", err)
	}

	got, err := ParseWorktrees(git("worktree", "list", "--porcelain", "-z"))
	if err != nil {
		t.Fatalf("ParseWorktrees: %v", err)
	}
	var prunable *Worktree
	for i := range got {
		if got[i].Prunable {
			prunable = &got[i]
		}
	}
	if prunable == nil {
		t.Fatalf("no prunable worktree found: %+v", got)
	}
	if prunable.PruneReason == "" {
		t.Errorf("prunable worktree %+v has no reason", prunable)
	}
}

func TestIntegrationWorktreePathWithNewline(t *testing.T) {
	git, dir := repo(t)
	odd := filepath.Join(dir, "..", "two\nlines")
	if out, err := exec.Command("git", "-C", dir, "worktree", "add", "-q", "--detach", odd, "HEAD").CombinedOutput(); err != nil {
		t.Skipf("this filesystem will not take a newline in a path: %v: %s", err, out)
	}

	got, err := ParseWorktrees(git("worktree", "list", "--porcelain", "-z"))
	if err != nil {
		t.Fatalf("ParseWorktrees: %v", err)
	}
	var found bool
	for _, worktree := range got {
		if strings.Contains(worktree.Path, "two\nlines") {
			found = true
		}
	}
	if !found {
		t.Errorf("no worktree with a newline in its path: %+v", got)
	}
}
