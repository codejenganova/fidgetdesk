# Contributing to FidgetDesk

Thanks for helping out! This guide covers how to get set up and how changes get merged.

## Ground rules

- Be kind. See the [Code of Conduct](CODE_OF_CONDUCT.md).
- **Never commit real data.** That means no client names, passwords, email exports, logs, or `.db` files. Use made-up examples ("Smile Dental", "Apt 4B").
- Report security issues privately. See [SECURITY.md](SECURITY.md).

## Development setup

1. Install [Go 1.24+](https://go.dev/dl/) and, optionally, [golangci-lint v2](https://golangci-lint.run/welcome/install/).
2. Clone the repo and run the app:

   ```sh
   go run ./cmd/fidgetdesk
   ```

3. Before opening a pull request, run:

   ```sh
   gofmt -l .           # should print nothing
   go vet ./...
   go test -race ./...
   golangci-lint run
   ```

   On Windows, `-race` requires a C compiler. If you don't have one, run `go test ./...` and CI will run the race detector for you.

## Making a change

1. **Open or find an issue first** for anything bigger than a small fix, so we can agree on the approach.
2. **Create a branch** from `main`, named for the kind of change:
   `feat/email-intake`, `fix/ticket-status-404`, `docs/readme-install`, `chore/update-deps`.
3. **Write tests** for new behavior. Table-driven tests are preferred.
4. **Commit using [Conventional Commits](https://www.conventionalcommits.org/):**

   ```
   feat: add notes field to clients
   fix: return 404 for unknown ticket ids
   docs: explain the -addr flag
   ```

5. **Open a pull request** and fill in the template. CI must pass before merging.
6. PRs are **squash-merged**, so the PR title becomes the commit message on `main`. It should also follow Conventional Commits.

## Design principles

- **Single binary.** Everything (templates, CSS, migrations) is embedded with `go:embed`. No runtime downloads.
- **Pure Go, no cgo.** This keeps cross-compiling to Windows, macOS, and Linux a one-line build.
- **Few dependencies.** Prefer the standard library. Every new dependency needs a reason in the PR description.
- **Safe by default.** Listen on localhost only, escape all output, and reject cross-site form posts.
- **Optional extras stay optional.** Features like AI or invoicing must never be required for the core helpdesk to work.

## Releases

Versions follow [Semantic Versioning](https://semver.org/). User-facing changes go in [CHANGELOG.md](CHANGELOG.md) under **Unreleased** as part of each PR.
