# Security Policy

FidgetDesk stores information about people's computers and networks, so security reports are taken seriously.

## Reporting a vulnerability

**Please don't open a public issue.** Report privately through GitHub instead:
**[Report a vulnerability](https://github.com/codejenganova/fidgetdesk/security/advisories/new)**

Include what you found, how to reproduce it, and the version affected (`fidgetdesk -version`). You'll get an acknowledgement as soon as possible. This is a hobby project, so please allow reasonable time for a fix before disclosing publicly.

## Supported versions

Until version 1.0, only the latest release receives security fixes.

## Current security model

FidgetDesk is in early development and **has no authentication yet**. By default it listens only on `127.0.0.1`, so only the same computer can reach it. Don't expose it to a network you don't fully trust until login support lands.
