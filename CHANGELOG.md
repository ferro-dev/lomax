# Changelog

All notable changes to `lomax` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Architecture and planning document (`docs/music-cli-plan.md`).
- `LICENSE` (Apache 2.0) and `NOTICE` (with Alan Lomax attribution).
- `README.md`.
- Community infrastructure: `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`,
  `SECURITY.md`, `CHANGELOG.md`, `docs/attribution.md`, `docs/distros.md`,
  GitHub issue and pull request templates.
- Go module (`go.mod`, Go 1.22+) and a minimal CLI entrypoint built on Cobra:
  `lomax --version` and `lomax about` print the version and the mandatory Alan
  Lomax attribution line.
- Continuous integration (GitHub Actions): gofmt + `go vet`, and build + test
  across Ubuntu 22.04/24.04 (native, with the race detector) and Debian 12 /
  Fedora / Arch (containers), all on the Go 1.22 floor.
- Linting and dev tooling: `golangci-lint` config (`.golangci.yml`) plus a CI
  lint job, a `.githooks/pre-commit` hook mirroring the CI gates, and a
  `Makefile` (`build`/`test`/`vet`/`fmt`/`lint`/`hooks`/`check`).

### Changed
- Reconciled `docs/music-cli-plan.md` to the decided Go stack: rewrote the
  dependency-license table, architecture diagram, plugin-discovery model,
  transcode tag handling, distribution plan, and roadmap (M1/M4/M5/M6/M7) from
  the placeholder Python examples to Cobra / Bubble Tea / `dhowden/tag` /
  `modernc.org/sqlite` / `go-plugin` / Goreleaser. Moved GitHub Sponsors from a
  Day-1 item to Milestone 6. The §6 language comparison is retained as history.

[Unreleased]: https://github.com/ferro-dev/lomax/commits/main
