# Changelog

All notable changes to this project are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Initial project scaffold: single-binary Go web app with clients and tickets.
- In-memory store behind a `Store` interface, ready for a SQLite implementation.
- Security headers, HTML escaping, and protection against cross-site form posts.
- CI on Windows, macOS, and Linux; Dependabot; linting; contributor docs.
