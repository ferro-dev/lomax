# Contributing to lomax

Thanks for your interest in `lomax`. This document covers how to report bugs,
request features, and submit code.

> **Status: pre-development.** The project is working through
> [Milestone 0](docs/music-cli-plan.md#15-development-roadmap) (repository
> bootstrap). Until M0 is complete there is no buildable code yet — the most
> useful contribution right now is engaging with the
> [planning document](docs/music-cli-plan.md).

## Project scope (read this first)

`lomax` has a deliberately narrow scope. Before opening an issue or PR, please
make sure your idea fits:

- **Linux only.** macOS and Windows are explicit non-goals. "Make this work on
  $OS" requests for either will be closed. Community forks are welcome to take
  that on.
- **CLI only.** No GUI, no music player, no streaming client.
- **Fully open source, no paid tier.** Features are never gated.

See [Goals & Non-Goals](docs/music-cli-plan.md#2-goals--non-goals) for the full
statement.

## Reporting bugs

Open a [bug report](https://github.com/ferro-dev/lomax/issues/new?template=bug_report.md)
and fill in the template. A good report includes:

- `lomax --version` output
- Your distro and version (e.g. Ubuntu 24.04, Arch rolling)
- Versions of relevant system tools (`ffmpeg -version`, `fpcalc -version`)
- Exact command run and the full output (use a code block)
- What you expected versus what happened

## Requesting features

Open a [feature request](https://github.com/ferro-dev/lomax/issues/new?template=feature_request.md).
Describe the problem you're trying to solve, not just the solution you have in
mind. Check the [roadmap](docs/music-cli-plan.md#15-development-roadmap) first —
many things are already planned and scoped to a milestone.

For open-ended ideas and questions, use
[Discussions](https://github.com/ferro-dev/lomax/discussions) rather than Issues.

## Development setup

`lomax` is written in **Go 1.22+**. The repository is a
[multi-module monorepo](docs/music-cli-plan.md#16-open-decisions): the core CLI
lives at the repo root, and each first-party plugin has its own `go.mod` under
`plugins/<name>/`.

```bash
git clone https://github.com/ferro-dev/lomax.git
cd lomax
go build ./...
go test ./...
```

Some functionality shells out to system binaries (`ffmpeg`, `fpcalc`,
`metaflac`). See [docs/distros.md](docs/distros.md) for install commands per
distro.

### Code style

- `gofmt` (enforced; CI fails on unformatted code)
- `golangci-lint` (configuration in `.golangci.yml`) — install it with
  `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`
  or your distro's package.
- Enable the pre-commit hook with `make hooks` (equivalently
  `git config core.hooksPath .githooks`). It runs gofmt, `go vet`,
  `golangci-lint`, and the tests before each commit — the same gates CI
  enforces. `make check` runs them on demand.

## Submitting changes

1. Fork the repo and create a branch off `main`.
2. Make your change. **Tests are required** — code changes ship with
   corresponding test updates (add, update, or remove as appropriate).
3. Run `gofmt`, `golangci-lint`, and `go test ./...` locally; all must pass.
4. Open a PR and fill in the
   [pull request template](.github/pull_request_template.md).
5. `main` is protected — PRs require review before merge.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/) style:
`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`. Keep the subject in
imperative mood and under 50 characters.

## License

By contributing, you agree that your contributions are licensed under the
[Apache License 2.0](LICENSE), the same license that covers the project.
