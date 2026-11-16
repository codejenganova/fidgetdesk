# Notes for AI coding assistants

FidgetDesk is a single-binary, self-hosted helpdesk written in Go. Read README.md and CONTRIBUTING.md first; they are the source of truth.

## Rules

- Never push to `main`. Work on a branch (`feat/…`, `fix/…`, `chore/…`, `docs/…`) and open a PR.
- Commit messages follow Conventional Commits.
- Pure Go only: no cgo. For SQLite, use `modernc.org/sqlite`, not `mattn/go-sqlite3`.
- Prefer the standard library. Justify any new dependency in the PR description.
- All HTTP handlers go through `web.Server.ServeHTTP`, which applies security headers and the cross-origin check. Don't bypass it.
- Handlers depend on the `store.Store` interface, never on a concrete store.
- Never add real client data, logs, or database files to the repo or test fixtures.
- Update CHANGELOG.md under **Unreleased** for user-facing changes.

## Before finishing a change

```sh
gofmt -l .
go vet ./...
go test -race ./...
golangci-lint run
```
