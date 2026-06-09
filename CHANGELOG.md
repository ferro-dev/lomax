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

[Unreleased]: https://github.com/ferro-dev/lomax/commits/main
