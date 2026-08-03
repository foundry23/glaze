set quiet

binary := "glaze"

[private]
default:
    @just --list

[doc("Build the binary")]
build:
    go build -o {{binary}} .

[doc("Format code (goimports + gofumpt)")]
fmt:
    go tool goimports -w .
    go tool gofumpt -w .

[doc("Run go vet")]
vet:
    go vet ./...

[doc("Run the test suite")]
test:
    go test ./...

[doc("Vet + test")]
check: vet test

[doc("Install to GOBIN / $GOPATH/bin")]
install:
    go install .

[doc("Run without building (e.g. just run --origin=git-worktree)")]
run *args:
    go run . {{args}}

[doc("Run the integration tests (drives the real git binary)")]
test-integration:
    go test -tags=integration ./porcelain
