# FidgetDesk

[![CI](https://github.com/codejenganova/fidgetdesk/actions/workflows/ci.yml/badge.svg)](https://github.com/codejenganova/fidgetdesk/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A lightweight, self-hosted helpdesk for the person everyone calls when their computer breaks.

FidgetDesk is one executable. Download it, double-click it, and open your browser. There's no database server to install, no Docker, and no per-seat subscription.

**Who it's for**

- **Solo IT techs** juggling a handful of small-business clients (the dentist's office, the law firm, the shop down the street)
- **Homelabbers** who want a record of what they changed and why
- **The go-to tech person** for a family, a group of friends, or an apartment building

> **Status: early development.** The app runs, but data is currently kept **in memory only** and is lost when it stops. Persistent SQLite storage is the next milestone (see the [Roadmap](#roadmap)).

## Quick start

### Run from source

You need [Go 1.24+](https://go.dev/dl/).

```sh
git clone https://github.com/codejenganova/fidgetdesk.git
cd fidgetdesk
go run ./cmd/fidgetdesk
```

Then open <http://127.0.0.1:8080>.

### Options

| Flag | Default | Description |
| --- | --- | --- |
| `-addr` | `127.0.0.1:8080` | Address to listen on. The default only accepts connections from the same computer. |
| `-version` | | Print the version and exit. |

> ⚠️ FidgetDesk has **no login yet**. Keep the default `127.0.0.1` address unless you're on a network you fully trust.

## Roadmap

- [x] Clients and tickets in a simple web UI, shipped as a single binary
- [ ] SQLite storage: all data in one file, so a backup is a single copy
- [ ] Login for the technician
- [ ] Notes per client (network info, the printer that always jams)
- [ ] Email-to-ticket: clients email you, and tickets appear automatically
- [ ] Time tracking per ticket, with an export for invoicing
- [ ] Release builds for Windows, macOS, and Linux
- [ ] Optional AI assistant using any OpenAI-compatible endpoint (Claude, OpenAI, or a local model through Ollama). Off by default.

## Project layout

```
cmd/fidgetdesk/     entry point: flags, HTTP server, graceful shutdown
internal/store/     data model and the Store interface (+ in-memory implementation)
internal/web/       HTTP handlers, embedded HTML templates and CSS
```

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for how to get set up and how pull requests work.

## License

[MIT](LICENSE)
