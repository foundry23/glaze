package main

import (
	"fmt"
	"strings"

	"github.com/foundry23/glaze/porcelain"
)

type origin struct {
	name  string
	parse func([]byte) (any, error)
}

var origins = []origin{
	{
		name:  "git-worktree",
		parse: func(data []byte) (any, error) { return porcelain.ParseWorktrees(data) },
	},
}

func lookup(name string) (origin, error) {
	if name == "" {
		return origin{}, fmt.Errorf("--origin is required; valid origins: %s", strings.Join(originNames(), ", "))
	}
	for _, o := range origins {
		if o.name == name {
			return o, nil
		}
	}
	return origin{}, fmt.Errorf("unknown origin %q; valid origins: %s", name, strings.Join(originNames(), ", "))
}

func originNames() []string {
	names := make([]string, 0, len(origins))
	for _, o := range origins {
		names = append(names, o.name)
	}
	return names
}
