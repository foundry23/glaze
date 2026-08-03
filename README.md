# glaze

[![CI](https://github.com/foundry23/glaze/actions/workflows/ci.yml/badge.svg)](https://github.com/foundry23/glaze/actions/workflows/ci.yml)
[![Go version](https://img.shields.io/github/go-mod/go-version/foundry23/glaze)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

glaze converts git porcelain to JSON.

## Install

With [Homebrew](https://brew.sh):

```bash
brew install foundry23/tap/glaze
```

Or with Go:

```bash
go install github.com/foundry23/glaze@latest
```

## Usage

Pipe git's output in and name the format with `--origin`:

```bash
git worktree list --porcelain | glaze --origin=git-worktree
git worktree list --porcelain -z | glaze --origin=git-worktree
```

### Invalid input

glaze rejects anything that is not the origin format it was told to expect.

### Origins

| Origin | Converts | Emits |
|---|---|---|
| `git-worktree` | `git worktree list --porcelain [-z]` | JSON array of worktrees |

More are straightforward to add — see [Adding an origin](#adding-an-origin).

### Flags

| Flag | Purpose | Required |
|---|---|---|
| `--origin` | format on stdin | Yes |


## Examples

```bash

git worktree list --porcelain -z | glaze --origin=git-worktree | jq -r '.[].path'                                      # every path
git worktree list --porcelain -z | glaze --origin=git-worktree | jq -r '.[] | select(.prunable) | .path'               # directory has gone missing
git worktree list --porcelain -z | glaze --origin=git-worktree | jq -r '.[] | select(.branch_name) | .branch_name'     # skip detached and bare
```

## Adding an origin

1. **Add a parser** to `porcelain/`, exported as `Parse<Thing>([]byte)`. Accept
   both delimiter forms by building on the shared `scanner`, and return errors
   prefixed once with the git command's name.
2. **Register it** in `origins.go`:

   ```go
   {
       name:  "git-status",
       parse: func(data []byte) (any, error) { return porcelain.ParseStatus(data) },
   },
   ```

Each parser needs a unit test over captured bytes and an integration test that
produces those bytes from git.

## Development

```bash
just build             # go build -o glaze .
just fmt               # goimports + gofumpt
just check             # go vet + go test
just test-integration  # drives the real git binary
just run               # go run . (e.g. just run --origin=git-worktree)
```

## Contributing

Contributions are welcome:

1. **Open an issue first** for anything non-trivial so the approach can be agreed before you build it.
2. Fork and branch, and add tests where it makes sense.
3. Run `just check` before opening a PR.
4. Keep commits focused and the CI green.

AI-assisted contributions are welcome, but code **must** be reviewed by a human before being submitted.

Any contributions that have not obviously been reviewed by a human may be closed without comment or justification.

## License

[MIT](LICENSE).
