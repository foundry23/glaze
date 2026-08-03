package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

// version is overridable at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	app := &cli.Command{
		Name:  "glaze",
		Usage: "Convert git porcelain to JSON",
		Description: strings.TrimSpace(`
Reads git porcelain on stdin and writes JSON

EXAMPLES:
   git worktree list --porcelain | glaze --origin=git-worktree
   git worktree list --porcelain -z | glaze --origin=git-worktree | jq -r '.[].path'
`),
		Version:               version,
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "origin",
				Usage: "format on stdin: " + strings.Join(originNames(), ", "),
			},
		},
		ShellComplete: func(_ context.Context, cmd *cli.Command) {
			for _, name := range originNames() {
				fmt.Fprintln(cmd.Root().Writer, name)
			}
		},
		Action: run,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "error: "+err.Error())
		os.Exit(1)
	}
}

func run(_ context.Context, cmd *cli.Command) error {
	src, err := lookup(cmd.String("origin"))
	if err != nil {
		return err
	}
	if args := cmd.Args().Slice(); len(args) > 0 {
		return fmt.Errorf("unexpected argument %q: glaze reads only from stdin", strings.Join(args, " "))
	}
	data, err := read()
	if err != nil {
		return err
	}
	value, err := src.parse(data)
	if err != nil {
		return err
	}
	return encode(os.Stdout, value)
}

func read() ([]byte, error) {
	if !piped(os.Stdin) {
		return nil, errors.New("no input on stdin: pipe git output in")
	}
	return io.ReadAll(os.Stdin)
}

func encode(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

func piped(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice == 0
}
