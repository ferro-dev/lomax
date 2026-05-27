# lomax

A Linux-native CLI for managing a music library — opinionated, plugin-extensible, transcode- and device-sync-aware.

> **Status: pre-development.** This repository currently contains architecture and planning documents only. No code yet. See [`docs/music-cli-plan.md`](docs/music-cli-plan.md) for the full design.

## What it is

`lomax` is a command-line music library manager in the spirit of [beets](https://beets.io/), built fresh for Linux in 2026. It reads, writes, and reconciles audio file tags against MusicBrainz, AcoustID, Discogs, and Last.fm; organises files by configurable naming templates; and — via first-party plugins — maintains transcoded mirror libraries and syncs them to USB drives, SD cards, Rockbox players, MTP devices, and Subsonic-compatible servers.

## What it is not

- Not a GUI application.
- Not a music player.
- Not a streaming-service client.
- Not cross-platform. **Linux is the only supported operating system.** macOS and Windows are explicit non-goals.

See the [Goals & Non-Goals](docs/music-cli-plan.md#2-goals--non-goals) section of the plan for the full scope statement.

## Why "lomax"?

The project is named after **Alan Lomax** (1915–2002), the American ethnomusicologist who spent his life travelling, recording, and cataloguing folk music traditions — including the field recordings that built the Library of Congress's Archive of American Folk Song. His work is, in spirit, what this tool tries to be in software: a careful, durable system for finding, organising, and preserving a music collection.

`lomax` is an independent, unaffiliated homage. It is **not** a product of the Alan Lomax estate, the [Association for Cultural Equity](https://www.culturalequity.org/) (the nonprofit Lomax founded in 1983 to repatriate his recordings to their communities of origin), or the Library of Congress American Folklife Center. Those organisations are the canonical sources for his actual archive and legacy; this project just borrows the name.

If you don't already know Lomax's work, [Cultural Equity's archive site](https://research.culturalequity.org/) is the best place to start.

## Planning document

The full architecture and decision log is at [`docs/music-cli-plan.md`](docs/music-cli-plan.md). It covers:

- Language and stack choice (Go 1.22+)
- Linux platform strategy and supported distros
- Plugin system design (HashiCorp `go-plugin` over gRPC)
- Mirror libraries and transcode-on-sync
- Distribution plan (language-native → Flatpak → AUR/COPR/PPA)
- Development roadmap (M0–M8)
- All decisions resolved on 2026-05-27

## License

Apache 2.0 — see [`LICENSE`](LICENSE) for the full text and [`NOTICE`](NOTICE) for attribution. Rationale for the licensing choice (including the patent-grant clause and the decision to keep plugin licensing permissive) is in the [planning document](docs/music-cli-plan.md#5-open-source-strategy).

## Contributing

Pre-development. Issues and discussion are welcome once the repository is past Milestone 0. In the meantime, the planning document is the artifact to engage with.
