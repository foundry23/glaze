package porcelain

import (
	"errors"
	"fmt"
	"strings"
)

type Worktree struct {
	Path        string `json:"path"`
	Head        string `json:"head,omitempty"`
	Branch      string `json:"branch,omitempty"`
	BranchName  string `json:"branch_name,omitempty"`
	Bare        bool   `json:"bare"`
	Detached    bool   `json:"detached"`
	Locked      bool   `json:"locked"`
	LockReason  string `json:"lock_reason,omitempty"`
	Prunable    bool   `json:"prunable"`
	PruneReason string `json:"prune_reason,omitempty"`
}

func ParseWorktrees(data []byte) ([]Worktree, error) {
	worktrees, err := parseWorktrees(data)
	if err != nil {
		return nil, fmt.Errorf("worktree: %w", err)
	}
	return worktrees, nil
}

func parseWorktrees(data []byte) ([]Worktree, error) {
	s := newScanner(data)
	worktrees := []Worktree{}

	var current *Worktree
	flush := func() {
		if current != nil {
			worktrees = append(worktrees, *current)
			current = nil
		}
	}

	for s.more() {
		line := s.next()
		if line == "" {
			flush()
			continue
		}

		key, value := keyword(line)
		if key == "worktree" {
			flush()
			path, err := s.path(value)
			if err != nil {
				return nil, err
			}
			current = &Worktree{Path: path}
			continue
		}
		if current == nil {
			if len(worktrees) == 0 {
				return nil, fmt.Errorf(
					"input format not as expected, expected a line beginning \"worktree \", got %q",
					ellipsis(line, 60))
			}
			return nil, fmt.Errorf("input format not as expected, expected a line beginning \"worktree \", got %q", key)
		}

		switch key {
		case "HEAD":
			current.Head = value
		case "branch":
			current.Branch = value
			current.BranchName = strings.TrimPrefix(value, "refs/heads/")
		case "bare":
			current.Bare = true
		case "detached":
			current.Detached = true
		case "locked":
			current.Locked = true
			reason, err := s.reason(value)
			if err != nil {
				return nil, fmt.Errorf("%s: locked: %w", current.Path, err)
			}
			current.LockReason = reason
		case "prunable":
			current.Prunable = true
			reason, err := s.reason(value)
			if err != nil {
				return nil, fmt.Errorf("%s: prunable: %w", current.Path, err)
			}
			current.PruneReason = reason
		default:
			// Unknown fields are ignored
		}
	}
	flush()

	if len(worktrees) == 0 {
		return nil, errors.New("no worktrees found")
	}

	return worktrees, nil
}

func (s *scanner) reason(value string) (string, error) {
	if strings.HasPrefix(value, `"`) {
		return unquote(value)
	}
	return value, nil
}
