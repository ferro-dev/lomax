# lomax — Music Library CLI

> **Project name:** `lomax` (always lowercase, including binary name, repo, package, AUR slug, and prose references)
> **Named after:** Alan Lomax (1915–2002), American ethnomusicologist; cataloger and field-recorder of folk music traditions for the Library of Congress. The tool's purpose — discovering, organizing, and preserving a music library — mirrors his life's work. Attribution is mandatory in all primary documentation surfaces (see [Section 13](#13-community-infrastructure)).
>
> **Status:** Pre-development / Architecture phase  
> **Date:** 2026-04-13  
> **Author:** TBD  
> **License:** Apache 2.0 (core and plugins)  
> **Monetization:** Donation-only (GitHub Sponsors; project site TBD)

---

## Table of Contents

1. [Problem Statement](#1-problem-statement)
2. [Goals & Non-Goals](#2-goals--non-goals)
3. [Existing Alternatives](#3-existing-alternatives)
4. [Why Build Fresh](#4-why-build-fresh)
5. [Open Source Strategy](#5-open-source-strategy)
6. [Language & Framework Choice](#6-language--framework-choice)
7. [Linux Platform Strategy](#7-linux-platform-strategy)
8. [Architecture Overview](#8-architecture-overview)
9. [Plugin System Design](#9-plugin-system-design)
10. [Device Sync Reality Check](#10-device-sync-reality-check)
11. [Mirror Libraries & Transcode-on-Sync](#11-mirror-libraries--transcode-on-sync)
12. [Distribution Plan](#12-distribution-plan)
13. [Community Infrastructure](#13-community-infrastructure)
14. [Differentiation From Beets](#14-differentiation-from-beets)
15. [Development Roadmap](#15-development-roadmap)
16. [Open Decisions](#16-open-decisions)

---

## 1. Problem Statement

[beets](https://beets.io/) is the de facto CLI music library manager, but it carries ~15 years of architectural debt and has real frustrations in practice:

- Slow on large libraries (sequential metadata resolution, no async)
- Opaque matching — difficult to understand *why* it chose a particular match
- Non-standard configuration system (`confuse` library, bespoke)
- Plugin API is powerful but inconsistently documented; hooks are stringly-typed
- Distribution is pip-only; no first-class presence in apt, Flatpak, or other Linux package managers
- Limited scope: no device sync beyond a few community plugins

The goal is a modern, opinionated-but-configurable replacement that learns from beets' strengths and corrects its weaknesses.

---

## 2. Goals & Non-Goals

### Goals

- **CLI-first.** The terminal is the primary and designed interface. No GUI is planned.
- **Opinionated defaults.** Sensible out-of-box behavior for tagging, naming, and organization. No required configuration to get started.
- **Configurable.** Every opinionated default can be overridden via config file or flags.
- **Extensible via plugins.** Core handles local library management. Plugins handle sync, cloud services, and integrations.
- **Linux-native.** This is a Linux tool, designed for Linux from day 1. macOS and Windows are explicit non-goals. The project does not pretend to be cross-platform; design choices favor Linux idioms (XDG, systemd-friendly, distro packaging) without compromise. See [Section 7](#7-linux-platform-strategy).
- **Widely distributable on Linux.** Flatpak, apt, Fedora COPR/RPM, Arch (AUR), and language-native channels (pip/pipx, `cargo install`, `go install`) — whichever fit the chosen language.
- **Fully open source.** All code, all features, day 1.

### Non-Goals

- A GUI application.
- A streaming service client.
- A music player.
- A paid or "open-core" tier of any kind.
- **Cross-platform support.** macOS and Windows are out of scope. The project does not test on, ship binaries for, or accept "make this work on $OS" feature requests for either. Community ports may exist as forks; that is the upstream project's complete relationship to non-Linux platforms.

---

## 3. Existing Alternatives

Research as of April 2026. Focus is on CLI-capable, actively maintained tools.

### beets
- **Language:** Python
- **GitHub:** [beetbox/beets](https://github.com/beetbox/beets) (~14,500 stars)
- **Status:** Actively maintained
- **Strengths:** Feature-complete, large plugin ecosystem, MusicBrainz integration, AcoustID fingerprinting
- **Weaknesses:** Age of architecture, performance, distribution, config system, opaque matching UX

### wrtag
- **Language:** Go
- **GitHub:** [sentriz/wrtag](https://github.com/sentriz/wrtag)
- **Status:** Actively developed (commits as of March 2025)
- **Strengths:** Fast, composable, MusicBrainz-based, ships as single binary, optional web interface
- **Weaknesses:** Younger project, smaller ecosystem, less extensible plugin model
- **Notable:** The only credible emerging CLI competitor to beets. Architecture and goals overlap significantly with this project — worth monitoring closely.

### MusicBrainz Picard
- **Language:** Python/Qt
- **GitHub:** [metabrainz/picard](https://github.com/metabrainz/picard)
- **Status:** Actively maintained (official MusicBrainz project)
- **Strengths:** Authoritative MusicBrainz integration, excellent Unicode/internationalization, AcoustID support
- **Weaknesses:** GUI-first; `picard` has a CLI mode but it is not designed for scripted/pipeline use

### Kid3
- **Language:** C++/Qt
- **Status:** Active
- **Strengths:** Has a CLI mode (`kid3-cli`) as a secondary interface
- **Weaknesses:** GUI-first; CLI is an afterthought, not composable

### OneTagger
- **Language:** Rust
- **GitHub:** [Marekkon5/onetagger](https://github.com/Marekkon5/onetagger)
- **Status:** Active
- **Strengths:** Fast, interesting Rust implementation, multiple metadata sources
- **Weaknesses:** GUI-first; no CLI interface

### Summary

The CLI music library management space is genuinely thin. beets dominates, and no other tool offers a full combination of: CLI-first interface, library management (not just tagging), plugin extensibility, and modern Linux distribution (Flatpak, distro-native packages). This is a real gap.

---

## 4. Why Build Fresh

Contributing to beets versus building a new tool is a legitimate question. The case for building fresh:

| Issue | Why It Prevents Contribution |
|-------|------------------------------|
| 15-year-old architecture | Fundamental refactors (async, type hints throughout) require breaking changes beets cannot easily ship |
| `confuse` config system | Bespoke, non-standard; replacement would break all existing configs |
| Internal DB layer | Tightly coupled custom SQLite wrapper; replacing it is a full rewrite |
| Hook system | Stringly-typed event names; no formal hook specs; typing it correctly breaks plugin API |
| Distribution | No maintainer bandwidth for apt/Flatpak/distro packaging infrastructure |

A full modernization of beets' internals is effectively a rewrite. Building a new project allows clean design, a new name/brand, and no obligation to maintain backward compatibility with a decade of beets configs and plugins.

---

## 5. Open Source Strategy

### License: Apache 2.0

**Recommended license: Apache 2.0** (OSI-approved, [SPDX: Apache-2.0](https://spdx.org/licenses/Apache-2.0.html))

Reasoning over the alternatives:

- **vs. MIT:** Apache 2.0 includes an explicit patent grant clause. This matters in the audio space, where MP3 and AAC have historically had codec IP issues. Apache 2.0 protects contributors and users more completely.
- **vs. GPL v2/v3:** GPL is copyleft — it would require all plugin authors to also release their plugins under GPL. This is a valid choice if forcing an open plugin ecosystem is a goal. However, it reduces adoption by organizations that want internal/private plugins, and is a barrier for some Linux distribution packaging. The permissive choice (Apache 2.0) is recommended to maximize community growth.
- **vs. LGPL:** More complex to reason about; unnecessary for a CLI application.

> **Plugin Licensing — DECIDED 2026-05-27: Apache 2.0** for core and plugins both. Permissive across the board. Plugin authors may ship under any license they choose (including closed-source). Note that with Go subprocess plugins via gRPC, even a GPL-v3 core would not have forced plugin copyleft, so the practical difference was symbolic; Apache 2.0 wins on corporate adoption and Flathub compatibility.

### Dependency License Compatibility

All planned core dependencies are compatible with Apache 2.0:

| Dependency | License |
|------------|---------|
| Mutagen | LGPL-2.0 |
| musicbrainzngs | BSD-2-Clause |
| pyacoustid | MIT |
| Chromaprint (fpcalc binary) | LGPL-2.1 |
| Typer | MIT |
| Rich | MIT |
| Pluggy | MIT |
| SQLAlchemy | MIT |
| Dynaconf | MIT |
| libmtp | LGPL-2.1 |
| libgpod | LGPL-2.0 |

LGPL libraries are compatible with Apache 2.0 distribution. No GPL dependencies in the planned core stack.

### Monetization

- **Donation-only.** No paid tier, no "pro" features, no open-core split.
- **GitHub Sponsors** — set up at launch alongside the repository.
- **Project site** — documentation site with optional donation links. Platform TBD.
- No features are gated. The plugin system is not designed with monetization hooks.

---

## 6. Language & Framework Choice

> **Status: DECIDED — Go 1.22+.** Rationale recorded 2026-05-27. Workload is I/O-bound (MusicBrainz/AcoustID/Discogs HTTP, SQLite, FFmpeg subprocess), so Rust's peak-CPU advantage was smaller than headline. Go wins on: (1) HashiCorp `go-plugin` is the most mature subprocess-plugin IPC in any language — directly aligned with the "language-agnostic plugins" goal; (2) fastest iteration loop for solo-maintainer roadmap M1–M6; (3) goroutines + channels = simpler async metadata resolver than `tokio` lifetimes; (4) `wrtag` confirms Go is workable in this domain. Accepted costs: thinner audio-tagging library landscape (`dhowden/tag` reads, write libraries scattered across `bogem/id3v2` + `go-flac` + `abema/go-mp4`, possible upstream contributions); no equivalent to Symphonia, must shell to FFmpeg for decode.

The original three-way comparison below is retained for reference.

### Why this is a real question

Defaulting to Python — beets' language and the language with the deepest audio-metadata ecosystem — is the obvious move, but obvious moves should still be justified. With Linux as the only target platform, the language choice now turns on three real questions: how easy is it for someone else to write a plugin, how performant is metadata resolution on a 100,000-track library, and how cleanly does the project ship through Linux distribution channels (Flatpak, distro packages, language-native installers)?

The three serious candidates: **Python**, **Go**, **Rust**. Other languages (Node.js, C#, Java) are technically viable but each loses on at least one major axis (audio metadata library quality, ecosystem maturity, distribution friction, runtime weight) without winning on any. They are not considered further here.

### Comparison Across Project Goals

| Criterion | Python 3.11+ | Go 1.22+ | Rust (stable) |
|---|---|---|---|
| **Audio tag read/write** | Mutagen — gold standard, every format, mature | `dhowden/tag` is read-only; tag *writing* is fragmented across smaller libs (`go-flac`, `id3-go`) | `lofty-rs` — full read+write, modern, comprehensive |
| **Audio decoding** | Punt to `ffmpeg` subprocess | Punt to `ffmpeg` subprocess | Symphonia (decode-only, pure Rust, very mature); `ffmpeg` for encode |
| **AcoustID/Chromaprint** | `pyacoustid` + `fpcalc` binary, mature | HTTP API + `fpcalc` subprocess (no native client) | HTTP API + `fpcalc` subprocess (no native client) |
| **MusicBrainz client** | `musicbrainzngs` — mature, used by Picard | `gomusicbrainz` exists, less complete | `musicbrainz_rs` exists, niche |
| **Discogs / Last.fm clients** | `discogs_client`, `pylast` — mature | Hand-rolled HTTP + JSON | Hand-rolled HTTP + JSON |
| **MTP (Android)** | `python-mtp` over libmtp; works | `go-mtp` exists, sparser | Bindings to libmtp possible; sparser still |
| **libgpod (stock iPod)** | cffi binding workable; libgpod itself is archived | Effectively unavailable | Effectively unavailable |
| **Plugin discovery** | `pip install foo-plugin-x` and it just appears via setuptools entry points | Go's `plugin` package works on Linux, but is fragile across builds; subprocess plugins are the more robust choice | `libloading` for in-process, or subprocess; subprocess is the saner default |
| **Plugin authorship friction** | Lowest — anyone who can write Python can write a plugin, no toolchain needed | Higher — plugin authors need a Go toolchain and a build step | Highest — Rust toolchain plus matching ABI/version |
| **Distribution to end users** | pipx for language-aware install; PyPI; needs Python interpreter (every Linux distro ships one) | Single static binary; trivial to drop into apt/Flatpak/AUR packaging | Single static binary; same packaging story as Go |
| **Flatpak packaging** | Mature path via `flatpak-pip-generator`; some manifest verbosity | Trivial — drop a binary into the manifest | Trivial — drop a binary into the manifest |
| **Distro-package friendliness (apt, RPM, AUR)** | Easy via `python3-` packages on apt/dnf; native PEP 517 wheels | Easy — single binary; one package per arch | Easy — single binary; one package per arch |
| **Async / parallel metadata fetching** | `asyncio` works but is the rough part of Python | Goroutines + channels — simple and effective | `tokio` + `async/await` — powerful, more verbose |
| **Performance on 100k-track libraries** | The slow option; this is beets' main complaint | Fast | Fastest |
| **Dev velocity / time-to-first-feature** | Highest | High | Lowest (steep curve, especially for plugin/IPC layer) |
| **Solo-maintainer sustainability** | Best — least friction for casual contributions, fewer compile-step gotchas | Excellent — fast builds, simple toolchain | Hardest — long compile times, ecosystem churn, more upfront design needed |
| **TUI / Rich terminal output** | Rich (Textual ecosystem) — best-in-class | `lipgloss` + `bubbletea` (Charm) — excellent, very polished | `ratatui` — excellent, slightly lower-level |
| **CLI framework** | Typer — type-hint-native, completion generation | Cobra — de facto standard, completion generation | clap — de facto standard, derive-macro-driven |
| **Configuration story** | Dynaconf + `tomllib` (stdlib) | Viper, or stdlib + `BurntSushi/toml` | `serde` + `toml` crate; clean and typed |
| **Database / migrations** | SQLAlchemy Core + Alembic | `database/sql` + `sqlc` or `goose` for migrations | `sqlx` or `rusqlite` + `refinery` for migrations |
| **Native dep handling (libmtp, libchromaprint, etc.)** | cffi/ctypes binding mature; pip wheels can bundle | cgo works but complicates static-binary story; pure-Go alternatives where they exist | FFI is first-class but requires more wrapper code |

### How each option shifts the project

**If Python:** The project leans on the deepest existing audio-metadata ecosystem and inherits the easiest plugin authorship story in any language. Plugins ship as `pip install`-able packages and Just Work via setuptools entry points. Distribution on Linux is a non-issue: every distro ships a recent-enough Python 3, pipx is widely available, and Flatpak/apt/RPM packaging of Python applications is well-trodden. The cost is performance — for libraries up to roughly tens of thousands of tracks, Python is fine; past that, the sequential metadata-resolver path becomes a bottleneck (this is beets' main complaint). With Linux-only as the target, **Python's case is much stronger than it was when the doc considered cross-platform** — its biggest historical weakness was distribution to non-Linux platforms, which no longer matters.

**If Go:** The project takes the same shape as `wrtag` — single static binary, fast, simple concurrency model via goroutines, trivial to drop into a Flatpak manifest or package as a `.deb`/`.rpm`. The two real costs: (1) the audio-metadata library landscape is materially thinner — `dhowden/tag` reads but does not write, and tag-writing libraries are scattered, less battle-tested, and may need contributions upstream or forks; (2) the plugin authorship model is harder — plugin authors need a Go toolchain, and the ergonomic path is subprocess plugins (a defensible choice — that's how `git`-style and many CLIs work, and it's how Terraform handles providers — but it changes plugin authoring significantly from "drop-in Python module" to "ship a binary that speaks a defined protocol"). On the upside, subprocess plugins are language-agnostic — third-party plugins could even be written in Python.

**If Rust:** The project gets the strongest performance ceiling, the strongest type system, and an audio-tagging library (Lofty) that is genuinely competitive with Mutagen now (this was not true a few years ago). Symphonia adds a pure-Rust audio decoder if needed. Distribution is the same single-binary story as Go. The costs: (1) significantly higher upfront design cost — the plugin system, async resolution, and error model all require more decisions before code can be written; (2) longer compile times will affect day-to-day iteration; (3) Rust toolchain barrier discourages casual plugin contributions; (4) several useful libraries (MusicBrainz, AcoustID, Discogs, Last.fm clients) are thinner in Rust and may need direct HTTP wrappers. This is the highest-ceiling but highest-cost option for a solo maintainer.

### Closer competitors as data points

- **`wrtag`** (Go) ships as a single binary, has a clean codebase, and confirms Go is workable for this domain. It does not have the ambitious plugin system this project plans, however.
- **`OneTagger`** (Rust) confirms Rust is workable for audio tagging at scale. It is a GUI app, not a CLI, but its tagging engine is a useful reference.
- **`beets`** (Python) is the existence proof that Python's ecosystem is sufficient — and also the existence proof of where Python's limits show up (slow imports, awkward async).

### What changes for the rest of this document

Most of the rest of the document is **language-agnostic**: the architecture, the plugin hook spec, the import workflow, the mirror-library design from [Section 11](#11-mirror-libraries--transcode-on-sync). What changes by language:

| Section | Python form | Go form | Rust form |
|---|---|---|---|
| Plugin discovery | setuptools entry points + Pluggy | Subprocess plugins via JSON-RPC or HashiCorp `go-plugin` | Subprocess plugins, or `libloading` for in-process |
| Tagging library | Mutagen | `dhowden/tag` (read) + a writer (TBD) | Lofty |
| CLI framework | Typer + Rich | Cobra + lipgloss/bubbletea | clap + ratatui |
| Config | Dynaconf | Viper or stdlib toml | serde + toml |
| Database | SQLAlchemy Core | `database/sql` + `sqlc` | `sqlx` |
| Distribution Phase 1 | PyPI / pipx | GitHub Releases binary | GitHub Releases binary |
| Distribution Phase 2 | Add Flatpak | Add Flatpak | Add Flatpak |
| Distribution Phase 3 | apt PPA, COPR, AUR | apt PPA, COPR, AUR | apt PPA, COPR, AUR |

The dependency-license table in Section 5 currently assumes Python and will need to be revised if Go or Rust is chosen.

### Recommendation framing

There is no single right answer; this is a values trade-off. A short version of how to choose:

- **Pick Python** if the priority is: lowest risk, fastest path to first useful release, deepest plugin ecosystem, and lowest barrier to outside contributions. Performance ceiling is the accepted cost.
- **Pick Go** if the priority is: a single static binary that drops cleanly into any Linux packaging system, fast builds, and excellent concurrency. Accept the thinner audio-metadata library landscape and a subprocess-based plugin model.
- **Pick Rust** if the priority is: maximum performance, strongest correctness guarantees, and a long-term project where the upfront cost amortizes. Accept slower iteration and a steeper barrier to plugin contributors.

For a solo maintainer who values shipping and contribution velocity over peak performance, the bias is toward **Python** — particularly given Linux-only scope, where Python's distribution disadvantages disappear. For a project where a clean single-binary deliverable is paramount, the bias is toward **Go**. For a project willing to spend the design budget for a long-term performance ceiling, **Rust**.

> **Open Decision — Language.** Decide before first commit. Listed in [Section 16](#16-open-decisions).

The remainder of this document continues to use the original Python-leaning concrete examples (Mutagen, Typer/Rich, Pluggy, Dynaconf, SQLAlchemy) for illustration purposes, since they are the most fleshed out and the project's prior art is mostly Python. Replace mentally with the equivalent stack from the table above if Go or Rust is chosen.

### If Python: Concrete Stack

**Python 3.11 minimum** for: `tomllib` in stdlib (config parsing), `ExceptionGroup` (async error handling), performance improvements, and `Self` type. Python 3.11 is available on every currently-supported Linux distro: Debian 12, Ubuntu 22.04 LTS (via deadsnakes if not using system Python) and 24.04 LTS, Fedora 38+, Arch (rolling), openSUSE Tumbleweed.

**Audio Tagging — Mutagen.** Most battle-tested Python tagging library; covers every format needed: MP3 (ID3v2.3/v2.4), FLAC (Vorbis Comments), OGG Vorbis, Opus, AAC/ALAC (.m4a), AIFF, WavPack, Monkey's Audio (APEv2), Musepack. Write a thin, project-owned abstraction layer over Mutagen rather than inheriting beets' `mediafile` layer.

**Metadata Sources.**

| Source | Python Library | Role |
|--------|----------------|------|
| MusicBrainz | `musicbrainzngs` | Primary: accurate release/recording/artist data |
| AcoustID + Chromaprint | `pyacoustid` + `fpcalc` binary | Audio fingerprinting for untagged/mistagged files |
| Discogs | `discogs_client` | Fallback for releases not in MusicBrainz; good for vinyl rips |
| Last.fm | `pylast` | Genre enrichment, supplemental metadata, similar artists |

Resolution order: **MusicBrainz → AcoustID → Discogs → Last.fm**. Each source is optional and can be disabled per-user in config.

**CLI Framework — Typer + Rich.** Typer for command/subcommand definition; Rich for terminal output (tables of proposed tag changes, progress bars, colored diffs). Both authored by the same developer and composing cleanly.

**Plugin System — Entry Points + Pluggy.** Two-layer architecture: setuptools entry points for discovery (a plugin installed via `pip` is automatically available); Pluggy for the internal hook system (the hook system used by pytest — typed, documented hook specs, call ordering, `firstresult` semantics).

**Configuration — Dynaconf + TOML.** Layered configuration (built-in defaults → system config → user config → environment variables → CLI flags). TOML via Python 3.11's stdlib `tomllib`. JSON Schema published for IDE autocompletion.

**Database — SQLAlchemy Core + SQLite.** Core (not ORM) for the library database; Alembic for migrations; SQLite as the backing store (serverless, single-file).

### If Go: Concrete Stack Sketch

**Go 1.22+** for generics maturity, `slices` package, structured logging.

**Audio Tagging.** `github.com/dhowden/tag` for reads; for writes, the most credible options are `github.com/go-flac/go-flac` (FLAC), `github.com/bogem/id3v2` (MP3), and `github.com/abema/go-mp4` (MP4). Some forking and contribution upstream is likely. Alternative: shell out to `metaflac` and `mid3v2` for writes — uglier but battle-tested; both ship in every Linux distro.

**CLI Framework.** Cobra + Charm's `bubbletea`/`lipgloss` for interactive UI. `pterm` is a simpler alternative if no full-screen TUI is needed.

**Plugin System.** Out-of-process subprocess plugins communicating over stdin/stdout via JSON-RPC, or HashiCorp's `go-plugin` (which uses gRPC over a Unix-domain socket). This is the same model HashiCorp uses for Terraform providers, so the design is well-trodden. Plugins can be written in any language.

**Config.** Stdlib + `github.com/BurntSushi/toml` is leanest. Viper if env-var/flag layering is wanted out of the box.

**Database.** `database/sql` with `mattn/go-sqlite3` (cgo) or `modernc.org/sqlite` (pure Go — preferable for static binary builds). Migrations via `pressly/goose` or `golang-migrate`.

### If Rust: Concrete Stack Sketch

**Rust stable**, MSRV pinned conservatively (~6 months behind latest).

**Audio Tagging.** `lofty` for read+write across all relevant formats. `symphonia` for decoding (used for AcoustID fingerprinting prep); encoding goes through `ffmpeg` subprocess regardless.

**CLI Framework.** `clap` (derive macros) + `ratatui` for any TUI needs. `indicatif` for progress bars, `dialoguer` for interactive prompts.

**Plugin System.** Subprocess plugins (recommended — language-agnostic, ABI-stable across compiler versions, less prone to soundness issues than dlopen) or `libloading` for in-process dynamic libraries.

**Config.** `serde` + `toml` crate. `figment` if Dynaconf-style layering is wanted.

**Database.** `sqlx` (compile-time SQL checks) or `rusqlite` (lower-level). Migrations via `refinery`.

---

## 7. Linux Platform Strategy

The project targets Linux exclusively. macOS and Windows are explicit non-goals ([Section 2](#2-goals--non-goals)). This section covers what "Linux" means in practice — which distros, which kernel/glibc baselines, and which Linux-specific facilities the project relies on.

### Distro Support Matrix

| Distro | Status | Rationale |
|--------|--------|-----------|
| Ubuntu 22.04 LTS, 24.04 LTS | Tier 1 | Most common LTS pair in the wild; CI runs both |
| Debian 12 (Bookworm) | Tier 1 | Stable baseline; CI runs |
| Fedora 39+ | Tier 1 | Modern reference distro |
| Arch Linux (rolling) | Tier 1 | Tracks latest libraries; AUR distribution path |
| openSUSE Tumbleweed | Tier 2 | Best-effort; no dedicated CI |
| RHEL/Rocky/Alma 9 | Tier 2 | Older glibc; older Python; known-good but not CI-gated |
| Alpine Linux | Tier 2 | musl libc differences may surface; community-supported |
| Older releases (Ubuntu 20.04, RHEL 8, Debian 11) | Unsupported | Pre-Python 3.11; pre-glibc 2.34; not worth the back-port cost |

CI runs the Tier 1 set on every PR. Tier 2 distros get a nightly job and accept bug reports.

### System Baseline

- **Kernel 5.15+** (Ubuntu 22.04's LTS kernel). No newer kernel features depended on.
- **glibc 2.35+** (Ubuntu 22.04). Ubuntu 24.04 ships 2.39, Debian 12 ships 2.36, Fedora 39 ships 2.38 — all comfortably above the floor. If the chosen language compiles to a binary (Go, Rust), build against glibc 2.35 to maximize compatibility, or use musl for a maximally-portable build.
- **systemd presence is not assumed.** The tool is a CLI, not a daemon. No `.service` units, no D-Bus activation requirements. Users on systemd-free distros (Devuan, Artix) should not have a worse experience.
- **Wayland vs X11: irrelevant.** This is a terminal application.

### Filesystem & Path Behavior

All target Linux filesystems (ext4, btrfs, XFS, ZFS, F2FS) are case-sensitive, support arbitrary path lengths, and provide nanosecond mtime. The mirror state DB ([Section 11](#11-mirror-libraries--transcode-on-sync)) and library DB use case-sensitive path keys with no path-length workarounds needed.

**Exception — sync target filesystems.** SD cards on devices like the Hifi Walker H2 are typically formatted as exFAT or FAT32. These have 2-second mtime granularity and case-insensitive (but case-preserving) filenames. The sync plugins must:

- Downcast mtime precision when comparing against a FAT-family destination.
- Treat the device-side filename as case-insensitive for orphan detection.
- Sanitize filenames for FAT-illegal characters (`< > : " | ? *`) at write time. This is the same sanitization that Picard, beets, and every other tagger does for FAT targets; the chosen tagging library will likely already provide it.

The source-side library is always on a case-sensitive filesystem, so this is a one-way filename normalization, not a round-trip problem.

### Unicode Normalization

Tag values and filenames are stored as **UTF-8 NFC** throughout. This is the modern Linux default — every major filesystem and userland tool produces NFC. The project does not need to handle NFD-to-NFC conversion at read time. If the chosen language has a Unicode library that defaults to anything other than NFC (none of Python, Go, or Rust do), normalize explicitly at write time.

### XDG Base Directory Spec

The project follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) verbatim. No exceptions, no fallback to dotfile-in-`$HOME`.

| Purpose | Location | Override |
|---------|----------|----------|
| Config | `${XDG_CONFIG_HOME:-$HOME/.config}/lomax/config.toml` | `--config <path>` flag, or `LOMAX_CONFIG` env var |
| State (DBs, cursors) | `${XDG_STATE_HOME:-$HOME/.local/state}/lomax/` | `LOMAX_STATE_DIR` |
| Data (mirrors, etc.) | `${XDG_DATA_HOME:-$HOME/.local/share}/lomax/` | per-profile config |
| Cache | `${XDG_CACHE_HOME:-$HOME/.cache}/lomax/` | `LOMAX_CACHE_DIR` |
| Logs | `${XDG_STATE_HOME:-$HOME/.local/state}/lomax/logs/` | `--log-file <path>` |

System-wide config (`/etc/lomax/config.toml`) is also read if present, with user config layered on top. Useful for distro packages that want to ship a default config.

### Mounted Device Discovery

For the `sync-fs` plugin, "find the H2 SD card" is solved on Linux by: (1) the user mounts the device (manually, automatically via udisks2/GNOME Files/KDE solid, or via a desktop helper); (2) the plugin reads `/proc/self/mountinfo` to enumerate mount points; (3) it matches by volume label (FAT label, exFAT label, or `*-uuid`) configured in the sync profile.

**libudisks2 / udisksctl:** a future enhancement could let the plugin trigger mount/unmount itself via udisks2 D-Bus, removing the manual-mount step. Out of scope for v1; the manual-mount workflow is fine and matches how rsync, beets, and similar tools work today.

The `mtp` plugin uses libmtp directly via the chosen language's bindings — no mount step required.

### FFmpeg, Chromaprint, and Other Native Dependencies

Native binaries the project shells out to:

| Tool | Required by | Install (Debian/Ubuntu) | Install (Fedora) | Install (Arch) |
|------|-------------|------------------------|------------------|----------------|
| `ffmpeg` | transcode plugin | `apt install ffmpeg` | `dnf install ffmpeg` (RPM Fusion) | `pacman -S ffmpeg` |
| `fpcalc` (Chromaprint) | AcoustID fingerprinting | `apt install libchromaprint-tools` | `dnf install chromaprint-tools` | `pacman -S chromaprint` |
| `metaflac` (optional) | FLAC tag operations (Go path) | `apt install flac` | `dnf install flac` | `pacman -S flac` |

The tool checks for these at startup and prints a clear "install with `apt install …`" hint if missing. Bundling them is rejected: every Linux distro has them packaged, FFmpeg has GPL components depending on build flags (a packaging-licensing minefield), and bundling would multiply the artifact size.

For Flatpak builds, FFmpeg and Chromaprint are pulled in as Flatpak-side dependencies via the manifest.

### Sandbox Awareness (Flatpak)

If the project ships as a Flatpak, the sandbox imposes constraints worth designing for from day 1:

- The sandbox cannot see arbitrary host paths. `--filesystem=home` (or, more conservatively, `--filesystem=xdg-music`) is required for the tool to read the user's library.
- USB device access for the `mtp` plugin requires `--device=all` or `--device=usb`. This is granted by Flatpak permission, not silently bypassed.
- Mounted SD cards under `/run/media/$USER/` are *not* visible inside the sandbox by default; `--filesystem=/run/media` is needed.

These are documented in `docs/flatpak-permissions.md` and reflected in the published manifest. The non-Flatpak install paths (apt, AUR, pipx) have none of these constraints.

### Shell Integration

Tab completion ships for **bash, zsh, and fish** — the three shells in standard Linux distro repositories. Generation is delegated to the chosen CLI framework (Typer, Cobra, and clap all support all three).

---

## 8. Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                      CLI Layer (Typer + Rich)                 │
│    import │ tag │ organize │ query │ sync │ config │ plugin   │
└───────────────────────────┬──────────────────────────────────┘
                            │
┌───────────────────────────▼──────────────────────────────────┐
│                        Core Library                           │
│                                                               │
│  ┌─────────────┐  ┌───────────────┐  ┌─────────────────────┐ │
│  │  File I/O   │  │    Tagger     │  │     Library DB      │ │
│  │  (pathlib)  │  │   (mutagen)   │  │ (SQLite/SQLAlchemy) │ │
│  └─────────────┘  └───────────────┘  └─────────────────────┘ │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐  │
│  │                  Metadata Resolver                     │  │
│  │     MusicBrainz → AcoustID → Discogs → Last.fm        │  │
│  │              (async, parallel queries)                 │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐  │
│  │              Plugin System (Pluggy + Entry Points)     │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
                            │
          ┌─────────────────┼──────────────────┐
          │                 │                  │
┌─────────▼──────┐  ┌───────▼──────┐  ┌───────▼────────┐
│ Plugin:        │  │ Plugin:      │  │ Plugin:        │
│ iPod Sync      │  │ Android/MTP  │  │ Subsonic/      │
│ (libgpod)      │  │ (libmtp)     │  │ Navidrome      │
└────────────────┘  └──────────────┘  └────────────────┘
```

### Key Workflows

**Import workflow** (adding new files to managed library):
1. Scan source path → collect audio files
2. Read existing tags → display summary table (Rich)
3. Resolve metadata from configured sources (async)
4. Present proposed changes as a diff (Rich)
5. User confirms, edits, or skips per-album
6. Write tags → move/copy files to library path per naming template
7. Register files in library database

**Manage-in-place workflow** (fixing tags on already-organized files):
1. Scan library path → diff current tags against metadata source
2. Show what would change (Rich table, colored diff)
3. Batch apply, interactive review, or dry-run

**Naming templates:**
- Configurable via TOML; sensible default provided
- Template variables: `{artist}`, `{album_artist}`, `{album}`, `{year}`, `{disc}`, `{track}`, `{title}`, `{format}`, `{bitrate_class}`
- Default example: `{album_artist}/{year} - {album}/{disc:02d}-{track:02d} {title}.{ext}`

---

## 9. Plugin System Design

### Discovery

Plugins declare themselves via a setuptools entry point in their `pyproject.toml`:

```toml
[project.entry-points."<tool-name>.plugins"]
my_plugin = "my_package.plugin:MyPlugin"
```

Installing the plugin package (`pip install tool-my-plugin`) makes it available. The tool discovers all registered entry points at startup.

### Hook Specifications

Plugins hook into defined lifecycle events. Hook specs are versioned; the core guarantees backward compatibility within a major version.

Planned hook categories:

| Hook | When Called | Plugin Use Case |
|------|-------------|-----------------|
| `pre_import` | Before any files are processed | Custom pre-flight validation |
| `post_import` | After files written and DB updated | Trigger sync, notify external system |
| `pre_write_tags` | Before tags written to a file | Modify or add tags programmatically |
| `post_write_tags` | After tags written | Checksumming, logging |
| `resolve_metadata` | During metadata resolution | Add a custom metadata source |
| `format_path` | When generating destination file path | Custom path logic |
| `sync_device` | When `sync` command is invoked | Implement device sync |
| `pre_sync_transform` | Before files are sent to a sync target | Transform source files (e.g., transcode FLAC→MP3) and substitute a mirror path |
| `register_commands` | At CLI startup | Add new subcommands |

### Plugin Configuration

Plugins receive their own namespaced config section:

```toml
[plugins.my_plugin]
some_option = "value"
```

### First-Party Plugins (Bundled, Separate Packages)

These are maintained in the same **multi-module monorepo** — one Git repo, but each plugin has its own `go.mod` under `plugins/<name>/` and ships its own binary. Installed optionally:

- `lomax-plugin-transcode` — Mirror library generation; produces a parallel library with a different format/bitrate (e.g. FLAC source → MP3 320 mirror). Maintains its own state DB to detect changed source files. See [Section 11](#11-mirror-libraries--transcode-on-sync).
- `lomax-plugin-sync-fs` — Filesystem sync to mounted targets (USB drives, SD cards, mounted network shares, Rockbox devices). Composes with `transcode` for transcode-on-sync workflows.
- `lomax-plugin-mtp` — Android/MTP device sync.
- `lomax-plugin-ipod` — iPod Classic sync (libgpod wrapper; legacy/community-supported, only relevant for stock iPod OS — Rockbox iPods use `sync-fs`).
- `lomax-plugin-subsonic` — Sync to Navidrome/Airsonic-Advanced/Subsonic servers.
- `lomax-plugin-lastfm` — Last.fm scrobble integration + tag enrichment.
- `lomax-plugin-discogs` — Discogs metadata source.

**Plugin composition.** The `transcode` and `sync-*` plugins are deliberately separate. `transcode` produces a mirror directory; `sync-*` plugins move files from a source path (the main library OR a mirror) to a target. This composition means a user with an all-FLAC library who only wants to sync as-is uses `sync-fs` alone; a user who needs MP3 320 on the device chains `transcode` then `sync-fs`. The `sync` CLI command coordinates this — see [Section 11](#11-mirror-libraries--transcode-on-sync).

---

## 10. Device Sync Reality Check

| Target | Protocol | Library | Assessment |
|--------|----------|---------|------------|
| USB drive / mounted share | Filesystem | stdlib | Trivial; v1 scope |
| NAS / SMB / NFS share | Filesystem (when mounted) | stdlib | Trivial once mounted |
| SD card (e.g., Hifi Walker H2 / Rockbox) | Filesystem (mounted) | stdlib | Trivial; identical to USB drive case |
| Rockbox iPod (any model) | Filesystem (mounted) | stdlib | Trivial; Rockbox exposes the iPod as a plain mass-storage device |
| Android device | MTP | `libmtp` + `python-mtp` | Works; MTP is occasionally unreliable; v2 plugin scope |
| iPod Classic — stock iPod OS | Custom USB + iTunesDB | `libgpod` | Library is archived; mostly functional for legacy hardware; community-plugin scope |
| iPhone / iPad | AFC/USB | `libimobiledevice` | **Music sync not supported** by libimobiledevice; file-system access only |
| Subsonic-compatible server | HTTP REST API | `py-sonic` | Clean API; good plugin candidate |

**v1 recommendation:** Filesystem targets only (USB drives, SD cards, mounted network shares, Rockbox devices). This covers the majority of enthusiast use cases — including DAPs like the Hifi Walker H2 and any Rockbox-loaded iPod — without native library dependencies.

**v2 recommendation:** MTP (Android) as the first non-filesystem sync plugin. Larger addressable device market than stock iPods.

**Stock iPod note:** Stock-OS iPod support is a legitimate use case, but `libgpod` is archived and the device is discontinued. This should be explicitly scoped as a community-maintained plugin rather than a core-team responsibility. Users in a position to flash Rockbox on their iPods should be encouraged to do so — it sidesteps `libgpod` entirely and routes through the same filesystem sync plugin used for everything else.

**Format mismatch between library and device.** Many of these targets — DAPs, Rockbox players, older iPods — either don't benefit from FLAC (no high-end DAC, no audiophile headphones in use), can't decode FLAC at all, or have constrained storage that makes FLAC impractical. Solving this requires a *mirror library* — a parallel directory tree where every track is transcoded to a target format. This is significant enough in scope that it gets its own section: see [Section 11](#11-mirror-libraries--transcode-on-sync).

---

## 11. Mirror Libraries & Transcode-on-Sync

### Problem

The main library is the source of truth: FLAC where available, MP3 where the source had nothing else (web rips, Bandcamp-only-MP3 releases, old purchases, etc.). This is the right archive format — lossless when possible, original-format otherwise.

External devices have a different problem. A Hifi Walker H2 with 256 GB and a Rockbox iPod Classic with 128 GB are not audiophile playback contexts: the DAC, amp, and headphones used with them aren't resolving the difference between FLAC and a well-encoded MP3. FLAC on these devices buys nothing audible and costs storage. The optimum is **MP3 320 CBR** or **MP3 V0 VBR** for everything — close-to-transparent quality, broadly compatible, and roughly 2–3× smaller than FLAC. With under 100 GB of FLAC today, even MP3 320 (the larger of the two) leaves comfortable headroom on a 128 GB device.

This requires a second representation of the library: a **mirror library** that is a transcoded clone of the main library, structurally identical, but with every track in the device-friendly format.

### Design Choice: Plugin or Separate App?

**Conclusion: two first-party plugins in this tool, not a separate app.**

Reasoning:

- **Single source of metadata.** The transcode step needs the main library's metadata, file paths, naming template, and database. Extracting this into a separate app means duplicating the library reader, the path template engine, and the DB schema, or coupling the apps tightly anyway via an exported manifest. Both options are worse than keeping it in-process.
- **Sync is already plugin-scoped.** The original plan already places filesystem and MTP sync in plugins. Transcoding is upstream of sync in the same pipeline. Splitting it out across an app boundary cuts that pipeline at an awkward seam.
- **Existing plugin hooks fit.** With the addition of one hook (`pre_sync_transform`), the transcode plugin can transparently intercept files on their way to a sync target and substitute mirror paths. No special-case code in core.
- **Composes with non-transcode workflows.** A user who only wants FLAC-to-FLAC sync (e.g., to a phone with a good DAC) installs `sync-fs` alone. A user who wants transcode-on-sync installs both `transcode` and `sync-fs`. The plugins do not depend on each other — they cooperate through the hook system.

The original [Section 10](#10-device-sync-reality-check) reservation that iPod sync should be community-maintained still stands; what's added here is that **transcoding and filesystem sync are core enough to the enthusiast use case to be first-party plugins**, not afterthoughts.

### Existing Solutions Considered

This isn't a new problem and it's worth being honest about what already works:

| Solution | What it does | Why not just use it |
|----------|--------------|---------------------|
| `beets convert` plugin | Transcodes a beets query to a destination dir using FFmpeg; skips files that already exist at dest | Skip-logic is filename-existence only — no detection of source re-tags or re-encodes. No deletion of orphaned mirror files. No sync step. We'd inherit beets' problems we're explicitly leaving behind. |
| [`mtpsync`](https://github.com/barsanuphe/mtpsync) | FLAC→MP3 transcode + rsync to MTP device, source-hash named output | Works, but Linux-only, MTP-only, and built around beets' library. No tag preservation guarantee, no integration with our schema. |
| `FlacSquisher` / `fre:ac` | GUI batch transcoders | One-shot conversion tools, not synchronizers. No state tracking, no incremental updates. |
| `ffmpeg` + `rsync` shell pipeline | DIY: find FLAC files, transcode, rsync | Fully manual; no metadata copy from source's tags, no orphan detection beyond what `rsync --delete` provides, no resumability across devices. |

The capability gap that justifies building this in-tool: **incremental, bidirectional-aware, multi-target mirror maintenance with full tag fidelity and deletion handling, integrated with the same library DB that owns the source files.** None of the existing tools do all of that.

### Architecture: Two Plugins, One Pipeline

```
┌──────────────────────────────────────────────────────────────────┐
│                  Main Library (source of truth)                  │
│                  ~/Music/library/  (FLAC + MP3 mix)              │
│                  Tracked by core library DB                      │
└──────────────────────────────┬───────────────────────────────────┘
                               │
                  user runs: lomax sync <profile>
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│                  lomax-plugin-transcode                           │
│  - Reads sync profile config                                     │
│  - Diffs main library against mirror state DB                    │
│  - For each track: transcode (FFmpeg), copy, or skip             │
│  - Writes/refreshes mirror at configured mirror path             │
│  - Updates mirror state DB                                       │
└──────────────────────────────┬───────────────────────────────────┘
                               │
                Mirror Library (~/Music/mirrors/h2-mp3/)
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│                  lomax-plugin-sync-fs                             │
│  - Diffs mirror against device state DB                          │
│  - Copies new/changed files to mounted device                    │
│  - Removes files on device that no longer exist in mirror        │
│  - Updates device state DB                                       │
└──────────────────────────────────────────────────────────────────┘
                               │
                               ▼
                  Mounted Device (H2 SD card / Rockbox iPod)
```

The two plugins share a tag-copying utility from core (so re-encoded files carry forward all main-library tags) but otherwise have no direct dependency on each other. They communicate through filesystem state and the sync profile config.

### Sync Profiles

A user can have multiple devices with different formats, different subsets of the library, and different paths. Each is a named **sync profile** in config:

```toml
[sync.profiles.h2]
description = "Hifi Walker H2 — Rockbox, SD card"
mirror_path = "~/Music/mirrors/h2/"
target_path = "/run/media/$USER/H2_SD/Music/"
transcode = "mp3-320-cbr"      # named transcode preset
filter = ""                     # empty = entire library
include_album_art = true
art_max_dimension = 1000        # downscale embedded art for device

[sync.profiles.ipod_rockbox]
description = "iPod Classic — Rockbox"
mirror_path = "~/Music/mirrors/ipod/"
target_path = "/run/media/$USER/IPOD/Music/"
transcode = "mp3-v0"
filter = "rating:>=3"           # only synced ≥3-star tracks
include_album_art = true
art_max_dimension = 600

[sync.profiles.phone]
description = "Phone — DAC + IEMs, full quality"
mirror_path = "<unused>"
target_path = "/run/media/$USER/Phone/Music/"
transcode = "passthrough"       # no mirror, sync source as-is
filter = "added:>30d"           # last 30 days
```

Notes on the schema:

- **`transcode = "passthrough"`** skips the transcode plugin entirely; `sync-fs` runs against the main library directly. The mirror plugin is not required for non-transcoding profiles.
- **`filter`** uses the same query language as `lomax query` (Milestone 4), so a "what gets synced to this device" rule is just a saved query.
- **`mirror_path`** is per-profile, not per-format, intentionally — two profiles with the same target format but different filters need separate mirrors so neither gets surprised by orphan files from the other.

### Transcode Presets

Format and encoder settings are named presets in config, referenced by sync profiles. Defaults shipped:

| Preset | Codec | Settings | Notes |
|--------|-------|----------|-------|
| `mp3-320-cbr` | LAME via FFmpeg | `-b:a 320k` constant bitrate | Maximum quality MP3, broadest player compatibility |
| `mp3-v0` | LAME via FFmpeg | `-q:a 0` (V0 VBR, ~245 kbps avg) | Near-transparent, ~25% smaller than 320 CBR; avoid only on devices with broken VBR seek tables |
| `mp3-v2` | LAME via FFmpeg | `-q:a 2` (V2 VBR, ~190 kbps avg) | Compact preset for very storage-constrained devices |
| `aac-256` | FFmpeg native AAC | `-c:a aac -b:a 256k` | For Apple devices that prefer AAC |
| `passthrough` | none | copy-only | Used for "no transcode" profiles |

The preset table is user-extensible — anyone can define a new entry. The shipped LAME settings should be documented as "sane defaults; tune if you have an opinion." The user mentioned MP3 320, which maps to `mp3-320-cbr` exactly.

### Mirror State Database

This is the crux of doing this correctly. Beets' `convert` plugin checks "does the destination file exist?" — which fails the moment a source file is re-tagged or re-encoded, because the dest filename hasn't changed but the content should. The fix is a small **mirror state database** — one SQLite file per mirror — that tracks, for each mirror file:

| Column | Purpose |
|--------|---------|
| `source_path` | Where the source lived in the main library at last transcode |
| `source_size` | Byte size of source at last transcode |
| `source_mtime` | mtime of source at last transcode |
| `source_blake3` | Optional content hash of source (configurable, off by default for speed) |
| `mirror_path` | Where the transcoded file lives in the mirror |
| `mirror_size` | Byte size of the mirror file |
| `transcode_preset` | Which preset was used (e.g., `mp3-320-cbr`) |
| `transcoded_at` | Timestamp |

The transcode pass logic:

1. **Enumerate.** Walk the main library, applying the profile's filter.
2. **Decide per track.** For each candidate source file:
   - If no mirror state row exists → **transcode**.
   - If state row exists but `source_size` or `source_mtime` differs → **transcode** (file changed).
   - If state row exists and `transcode_preset` in row differs from the profile → **transcode** (preset changed, e.g., user switched from V0 to 320 CBR).
   - If hash verification is enabled and `source_blake3` differs → **transcode**.
   - Otherwise → **skip** (mirror is current).
   - Special case: if source format is already the mirror format and `never_re-encode_lossy` is true → **copy**, not transcode (avoid re-encoding existing MP3s into MP3 again, which always loses quality).
3. **Orphan cleanup.** Any mirror file with no corresponding row in the source enumeration (or whose source no longer exists) is deleted from the mirror, and its row removed from the state DB.

This handles the four real-world cases: (a) new tracks added to main library, (b) tracks re-tagged in main library, (c) tracks deleted from main library, (d) tracks renamed in main library (handled as delete + add, since the source path changed).

> **Open Decision — Hashing.** mtime + size catches almost everything. Content hashing catches the edge case of a tag rewrite that doesn't change file size and somehow preserves mtime (uncommon, but possible if the tagger explicitly preserves mtime). Recommend: hash off by default for speed; opt-in via config flag for users who want belt-and-braces.

### Tag Handling on Transcode

FFmpeg can copy metadata from source to destination, but not all tag fields survive across formats and not all of them survive *correctly* — this is the source of the duplicate-artist bug seen in the linked beets discourse thread. The correct approach:

1. Run FFmpeg without metadata copy (`-map_metadata -1`) so the output starts clean.
2. Read source tags through the project's Mutagen abstraction.
3. Map fields to ID3v2.4 (for MP3) using the same mapping the main importer uses.
4. Write tags to the transcoded file via Mutagen.
5. Embed/copy album art separately, optionally resizing it (a 5000×5000 cover doesn't help an iPod's screen and wastes flash).

Doing tags ourselves (rather than relying on FFmpeg's metadata copy) means the mirror inherits the project's normalized tag conventions automatically — including any custom fields and disambiguation logic the main library uses.

### Sync State Database

The `sync-fs` plugin maintains its own per-device state DB in the same fashion. For each device file:

| Column | Purpose |
|--------|---------|
| `mirror_path` | Source-side path within the mirror |
| `device_path` | Path on the device |
| `mirror_size` | Size at last sync |
| `mirror_mtime` | mtime at last sync |
| `synced_at` | Timestamp |

Logic mirrors the transcode pass: copy new, copy changed, delete orphans. Critically, this state is keyed on the **mirror**, not the main library, so a transcode pass that produces an identical mirror file (e.g., re-running the same preset) doesn't trigger unnecessary device writes.

The state DB lives at `<mirror_path>/.lomax-state.db` rather than on the device, so a device that's just been freshly mounted to a different machine doesn't lose continuity. (Devices like the H2 frequently get yanked out without proper unmounting; we should not depend on metadata stored on the device itself.)

> **Open Decision — Where to keep device state on multi-machine setups.** If a user syncs the same device from two different computers, each machine has its own state DB and they will disagree. v1 punts on this: sync from one machine. A future option is to write a small manifest file to the device on each sync containing per-file checksums — slower but recoverable across machines.

### CLI Surface

```
lomax sync                        # sync all enabled profiles, in order
lomax sync <profile>              # sync just one profile
lomax sync <profile> --dry-run    # show what would happen, change nothing
lomax sync <profile> --transcode-only   # rebuild mirror, don't push to device
lomax sync <profile> --push-only        # push current mirror, don't re-transcode
lomax sync status                 # report mirror freshness + device disk usage per profile
lomax sync prune <profile>        # force orphan cleanup pass on mirror or device
```

Rich output during sync: progress bar showing "transcoding 47/312", followed by "syncing 312 files (4.2 GB)". `--dry-run` prints a Rich diff of what would change on disk.

### Workflow Walk-Through (User's Stated Use Case)

User has FLAC + occasional MP3 main library, a Hifi Walker H2 with Rockbox, and an iPod Classic that may or may not be Rockbox.

**One-time setup:**

```bash
pipx install lomax
lomax plugin install transcode sync-fs
# edit config, add the [sync.profiles.h2] block above
```

**Initial sync to H2:**

```bash
# mount H2 SD card to /run/media/user/H2_SD/
lomax sync h2
# → transcodes ~100 GB FLAC + MP3 to ~50 GB MP3 320, populates mirror, copies to SD
```

First run is the slow one (full transcode). Subsequent runs only touch what changed. A typical "added one new album, fixed tags on a couple of tracks" sync takes seconds for the diff and the time to encode the changed files.

**Day-to-day:**

```bash
lomax import ~/Downloads/new-album/   # adds to main library
lomax sync h2                         # mirror catches up, device updated
lomax sync ipod                       # same album also goes to iPod
```

If the iPod is still on stock OS, the user enables `lomax-plugin-ipod` (libgpod) instead of `sync-fs` for that profile, with full awareness that this path is community-maintained. Once the refurbish is done and Rockbox is loaded, the iPod profile flips to `sync-fs` and the libgpod dependency goes away.

### Why Not iTunes / Why Not Existing Tools

Stock iPod sync today means iTunes, which means a parallel system that doesn't know about the main library — track ratings, play counts, and tag fixes don't flow between iTunes and the main library, and tag edits in the main library require a full re-export to iTunes. Folding the iPod into the same sync framework as everything else (whether through libgpod or, after Rockbox, through `sync-fs`) eliminates the dual-bookkeeping problem entirely. iTunes itself runs only on macOS and Windows, so for a Linux-only main library it's already an out-of-band tool requiring a separate machine or VM — another reason a Linux-native sync path is worth the engineering.

A standalone "transcode-and-sync" tool sitting next to the main CLI is technically possible but creates the same parallel-system problem: it would need its own copy of the path template, its own way to read the library DB, and its own filter language. Better to keep one library, one config, one query language, and let plugins extend the verb list.

### Summary of Decisions in This Section

- **Mirror libraries are first-party.** `lomax-plugin-transcode` and `lomax-plugin-sync-fs` ship in the monorepo from the first device-sync milestone.
- **They are separate plugins, not one combined plugin.** `transcode` produces a mirror; `sync-fs` pushes a directory (mirror or main library) to a target. Composing them is the user's call per profile.
- **Sync profiles are the user-facing abstraction.** Each profile maps a filter + transcode preset + mirror path + target path.
- **State databases are mandatory, not optional.** Both mirror and device passes maintain SQLite state for incremental updates and orphan cleanup. This is what differentiates the workflow from `beets convert` and from a hand-rolled `ffmpeg | rsync` pipeline.
- **Tag rewriting goes through Mutagen, not FFmpeg metadata copy.** Avoids the duplicate-artist class of bugs.
- **Default MP3 preset is `mp3-320-cbr`** — matches the user's stated preference and is the safest choice for player compatibility on H2 and Rockbox iPod alike.

---

## 12. Distribution Plan

> **Note:** This section currently describes Python-flavored distribution mechanics; conditional alternatives for Go and Rust are noted inline. The high-level phases (language-native → universal → distro-native) apply regardless of language.

### Phase 1 — Language-Native (Day 1)

The fastest path to a working install. Requires no external infrastructure beyond GitHub.

**If Python:** PyPI + pipx. `pipx install lomax` installs into an isolated virtualenv, available on every Linux distro that ships Python 3 (i.e., all of them).

- `pyproject.toml`-based build (PEP 517/518, `hatchling` as build backend)
- `pipx install lomax` as the recommended user-facing install
- `pip install lomax` also works
- GitHub Releases with built wheel (`.whl`) attached
- Changelog in `CHANGELOG.md` (Keep a Changelog format)

**If Go:** `go install <module>@latest` for users with a Go toolchain. GitHub Releases with prebuilt binaries for `linux-amd64` and `linux-arm64`. Goreleaser automates the release pipeline.

**If Rust:** `cargo install lomax` for users with a Rust toolchain. GitHub Releases with prebuilt binaries (`linux-amd64`, `linux-arm64`). `cargo-dist` automates the release pipeline.

For all three languages, the GitHub Releases binaries are signed with a project GPG key and accompanied by SHA256 checksums.

### Phase 2 — Flatpak / Flathub

A universal, distro-agnostic install path that works on every Linux distro from Ubuntu 22.04 to Alpine to NixOS without per-distro packaging effort.

- Submit a manifest to [Flathub](https://flathub.org/).
- Flathub requires an OSI-approved license (Apache 2.0 qualifies).
- Sandbox model imposes filesystem permission requirements; see [Section 7 → Sandbox Awareness](#7-linux-platform-strategy).
- The Flatpak manifest pulls in FFmpeg and Chromaprint as Flatpak-side dependencies, so users do not need to install them separately.
- For Python: `flatpak-pip-generator` produces the dependency manifest fragment; some manifest verbosity is unavoidable but well-documented.
- For Go/Rust: drop the binary into the manifest; trivial.

Flatpak is a particular fit for this project because it is Linux-only by design and it solves the "user has FFmpeg but it's the wrong build" class of problems that comes up with media tools.

### Phase 3 — Distro-Native Packages

Slowest to set up, but the most native install experience for users who prefer their distro's package manager. These can be done concurrently or one at a time.

- **AUR (Arch Linux)** — Often happens organically once the project has any users; can also be seeded by the maintainer with a `PKGBUILD`. Lowest effort of the three.
- **Fedora COPR** — Personal RPM repository; a `.spec` file plus the build artifact. After the project stabilizes, submission to Fedora proper is possible but not required.
- **Debian/Ubuntu PPA** — `.deb` packaging via `debhelper`/`dh-make`. Highest setup cost; ongoing maintenance for each new release. Hosted on Launchpad. After stability, submission to Debian proper is possible but optional.

### Explicitly Not Used

- **Homebrew** — Homebrew is technically available on Linux (`linuxbrew`), but the overwhelming majority of Homebrew users are on macOS, which is out of scope for this project. Maintaining a Homebrew tap to serve a small fraction of Linux users isn't worth the effort.
- **Snap** — Functional but the centralized store and proprietary backend make it a poor fit for this project's principles. AUR/COPR/PPA + Flatpak cover the same need with better licensing.
- **AppImage** — Optional consideration for power users who want a single-file binary they can drop on any distro. Can be added in Phase 3 if there's demand; not a priority because Flatpak covers the same use case more robustly.

### Packaging Principles

- **Core install must work without a C compiler.** Dependencies with native components (libmtp, libgpod, libchromaprint bindings) belong in optional plugin packages, not core. This is essential for Python builds where ABI-incompatible wheels are an ongoing source of pain; Go/Rust users get this for free.
- **System binaries (`ffmpeg`, `fpcalc`, `metaflac`) are runtime dependencies, not bundled.** The package manifest declares them; the user's distro provides them. Documented in [Section 7](#7-linux-platform-strategy).
- **Same artifact across distros where possible.** A single Flatpak bundle covers Debian/Ubuntu/Fedora/Arch/openSUSE. Distro-specific packages are valuable but secondary.
- **GPG-signed releases.** Every GitHub Release tag is signed; release artifacts have detached signatures published alongside.

---

## 13. Community Infrastructure

These are **Phase 1 deliverables**, not afterthoughts. A project launched without them has a worse first impression and generates more duplicate noise.

### Repository Setup (Day 1)

- `LICENSE` — Apache 2.0 full text
- `README.md` — what it is, why it exists, quick install, quick start, link to docs. **Must include an "About the name" section attributing the project to Alan Lomax (1915–2002).**
- `NOTICE` — Apache 2.0 notice file; includes Alan Lomax attribution as the named-after-credit (the tool is independent of and unaffiliated with the Alan Lomax estate or the Association for Cultural Equity, which should be linked explicitly).
- `CONTRIBUTING.md` — how to report bugs, request features, submit PRs; development setup guide
- `CODE_OF_CONDUCT.md` — [Contributor Covenant](https://www.contributor-covenant.org/) is the standard choice
- `CHANGELOG.md` — [Keep a Changelog](https://keepachangelog.com/) format
- `SECURITY.md` — how to report security vulnerabilities privately (GitHub has a dedicated feature for this)
- `.github/ISSUE_TEMPLATE/` — separate templates for bug reports and feature requests
- `.github/pull_request_template.md` — PR checklist
- `.github/CODEOWNERS` — once there are multiple maintainers

### Attribution Requirements (Alan Lomax)

The name `lomax` is taken from Alan Lomax. The project is **not** affiliated with the Alan Lomax estate, the Association for Cultural Equity (ACE — the nonprofit Lomax founded), or the Library of Congress American Folklife Center which holds his collection. Attribution must:

1. **Appear in README.md** under a dedicated "About the name" / "Attribution" section, with one paragraph on who Alan Lomax was, why the tool carries his name, and a disambiguation note that the project is not an official Lomax-estate product.
2. **Appear in the docs site landing page** ([Section 13 — Documentation Site](#documentation-site)) in equivalent form.
3. **Appear in `lomax about` / `lomax --version` output** as a one-line credit: `Named after Alan Lomax (1915–2002). Independent project, not affiliated with the Lomax estate or ACE.`
4. **Link to the [Association for Cultural Equity](https://www.culturalequity.org/)** as the canonical source for Lomax's actual archive and legacy.

If at any point the Lomax estate / ACE objects to the use of the name (unlikely given the homage framing, but possible), the project commits to renaming. Track this risk in `docs/attribution.md`.

### GitHub Features to Enable Day 1

- **GitHub Sponsors** — donation link appears on the repository page; requires a Sponsors profile
- **Discussions** — better than Issues for questions and general community conversation
- **Security advisories** — private vulnerability reporting
- **Branch protection** on `main` — require PR reviews before merge, even for the sole maintainer initially (sets the norm for future contributors)

### Documentation Site

**Platform: MkDocs + Material** (decided 2026-05-27). Best out-of-box UX in the OSS docs space — built-in instant search, theming, dark mode, content tabs, admonitions, mermaid diagrams. Python is a CI-side build dependency only; users never need it. Deployed via GitHub Pages (`mkdocs gh-deploy` or a GitHub Action on push to `main`).

Repository layout:
```
docs/
├── index.md
├── install.md
├── getting-started.md
├── reference/
│   ├── commands/
│   ├── config.md
│   └── hooks.md          # plugin hook API spec
├── plugins/
│   ├── authoring.md
│   ├── transcode.md
│   └── sync-fs.md
└── distros.md            # per-distro install + ffmpeg/fpcalc setup
mkdocs.yml
.github/workflows/docs.yml
```

---

## 14. Differentiation From Beets

| Beets Pain Point | This Project's Answer |
|-----------------|----------------------|
| Slow import on large libraries | Async metadata resolution; parallel queries to multiple sources |
| Opaque matching — hard to understand why it chose a match | Rich diff view: exact before/after for every field, with source attribution ("from MusicBrainz release MBID: ...") |
| YAML config via bespoke `confuse` | TOML config, stdlib parser, JSON Schema for IDE support |
| Stringly-typed plugin hooks | Pluggy with typed hook specifications; versioned API |
| pip-only distribution | Flatpak (universal), AUR/COPR/PPA (distro-native), language-native (PyPI/cargo/`go install`) — see [Section 12](#12-distribution-plan) |
| Requires knowing Python to install plugins | `lomax plugin install plugin-name` command wrapping the language-native installer |
| No device sync beyond community plugins | Filesystem sync (incl. Rockbox / DAPs) and transcode-on-sync as maintained first-party plugins; MTP next |

---

## 15. Development Roadmap

### Milestone 0 — Repository Bootstrap
- [x] **Decide language** — Go 1.22+ (decided 2026-05-27, see [Section 6](#6-language--framework-choice))
- [x] **Choose project name** — `lomax` (decided 2026-05-27, see [Section 16](#16-open-decisions))
- [ ] Create GitHub repository at `github.com/<org-or-user>/lomax`
- [ ] Add LICENSE (Apache 2.0), NOTICE (with Lomax attribution), README (state Linux-only support clearly; include "About the name" section), CONTRIBUTING, CODE_OF_CONDUCT, CHANGELOG, SECURITY, `docs/attribution.md`
- [ ] Add `docs/distros.md` covering install commands and FFmpeg/Chromaprint setup per supported distro (see [Section 7](#7-linux-platform-strategy))
- [ ] Set up GitHub Sponsors
- [ ] Build configuration: `pyproject.toml` (Python) / `go.mod` (Go) / `Cargo.toml` (Rust)
- [ ] CI pipeline (GitHub Actions) matrix: Ubuntu 22.04 + Ubuntu 24.04 + Debian 12 + Fedora 39 (via container) + Arch (via container); lint, type check, test
- [ ] Pre-commit hooks for the chosen language (ruff+mypy / gofmt+golangci-lint / rustfmt+clippy)
- [ ] Verify a "hello world" build artifact installs cleanly from Phase-1 distribution channel on a fresh Ubuntu 24.04 VM before any feature work begins

### Milestone 1 — Core Tag Reading & Display
- [ ] Read audio files recursively from a path
- [ ] Parse tags via Mutagen abstraction layer
- [ ] Display current tags in a Rich table (`lomax inspect <path>`)
- [ ] Identify format, duration, bitrate, and encoding info

### Milestone 2 — Metadata Resolution
- [ ] MusicBrainz lookup by existing tags (artist + album + title)
- [ ] AcoustID fingerprint generation + MusicBrainz lookup for untagged files
- [ ] Display proposed changes as a Rich diff (before/after per field)
- [ ] Dry-run mode

### Milestone 3 — Tag Writing & File Organization
- [ ] Write approved tag changes to files
- [ ] Configurable naming template engine
- [ ] Move or copy files to target directory structure
- [ ] Import workflow (new files → library)
- [ ] Manage-in-place workflow (existing library → review + fix)

### Milestone 4 — Library Database
- [ ] SQLAlchemy schema for tracks, albums, artists
- [ ] Track which files are managed; detect moved/deleted files
- [ ] Query interface (`lomax query artist:"David Bowie" year:1972`)

### Milestone 5 — Plugin System
- [ ] Pluggy hook specification definitions
- [ ] Entry point discovery at startup
- [ ] `lomax plugin list` / `plugin install` / `plugin remove` commands
- [ ] Port Discogs and Last.fm sources as first-party plugins

### Milestone 6 — PyPI Release
- [ ] Version 0.1.0 to PyPI
- [ ] `pipx install lomax` install path working and documented
- [ ] GitHub Release with wheel artifact
- [ ] Documentation site live (MkDocs + Material)

### Milestone 7 — Mirror Libraries & Filesystem Sync (Post-1.0)
- [ ] `lomax-plugin-transcode`: sync profiles, transcode presets (mp3-320-cbr, mp3-v0, mp3-v2, aac-256, passthrough)
- [ ] Mirror state DB: incremental transcoding, orphan cleanup, preset-change detection
- [ ] Tag rewrite via Mutagen (not FFmpeg metadata copy); embedded art handling + optional resize
- [ ] `lomax-plugin-sync-fs`: filesystem sync to mounted targets (USB, SD cards, Rockbox devices, NAS)
- [ ] Device state DB per profile, orphan cleanup on device, `--dry-run` and `--prune`
- [ ] `lomax sync` command surface: per-profile, all-profiles, status, transcode-only, push-only
- [ ] Integration tests with a fake device target (tmpdir)

### Milestone 8 — Specialized Sync Targets (Post-1.0)
- [ ] `lomax-plugin-mtp`: Android/MTP device sync (composes with transcode plugin)
- [ ] `lomax-plugin-ipod`: stock iPod Classic via libgpod (community-driven, best-effort)
- [ ] `lomax-plugin-subsonic`: Navidrome/Airsonic-Advanced/Subsonic server sync

---

## 16. Open Decisions

These require a decision before or shortly after repository creation:

| Decision | Options | Recommendation | Notes |
|----------|---------|----------------|-------|
| **Implementation language** | ~~Python 3.11+ / Go 1.22+ / Rust stable~~ | **DECIDED: Go 1.22+** (2026-05-27) | See [Section 6](#6-language--framework-choice) for rationale. |
| **Project name** | ~~TBD~~ | **DECIDED: `lomax`** (2026-05-27) | Always lowercase. Named for Alan Lomax. Attribution mandatory in README, docs landing page, and `about` / `--version` output. |
| **Plugin licensing** | ~~Apache 2.0 vs. GPL v3~~ | **DECIDED: Apache 2.0** (2026-05-27) | Permissive across core + plugins. Subprocess-plugin model meant GPL would not have enforced copyleft anyway. |
| **Config file location** | XDG vs. dotfile in `$HOME` | XDG (`$XDG_CONFIG_HOME/lomax/config.toml`); system-wide `/etc/lomax/config.toml` also read | Detailed in [Section 7](#7-linux-platform-strategy) |
| **Monorepo vs. multi-repo for plugins** | ~~Single-module monorepo / multi-module monorepo / multi-repo~~ | **DECIDED: Multi-module monorepo** (2026-05-27) | One repo, each plugin its own `go.mod` under `plugins/<name>/`. Atomic cross-cutting refactors during pre-1.0 hook-API churn + per-plugin dep surface + per-plugin release tags. Community plugins live in their own repos. |
| **Documentation platform** | ~~MkDocs + Material / Hugo / Docusaurus~~ | **DECIDED: MkDocs + Material** (2026-05-27) | Best out-of-box UX (search, theming, admonitions, content tabs). Python only in CI, not on user machine. Deploy via GitHub Pages. |
| **Minimum runtime version** | Language-dependent | Python 3.11 / Go 1.22 / Rust MSRV ~6 mo behind latest | 3.11 is on all current LTS distros; 3.12 is not yet on Ubuntu 22.04 LTS |
| **Default MP3 preset for sync mirrors** | ~~320 CBR vs. V0 VBR~~ | **DECIDED: 320 CBR** (2026-05-27) | Broadest device compatibility; VBR seek tables are imperfect on some old hardware. |
| **Mirror change detection** | ~~mtime+size only vs. +content hash~~ | **DECIDED: mtime+size default; BLAKE3 opt-in** (2026-05-27) | Hashing is slow on large libraries; mtime+size catches the realistic cases. Config flag for paranoid users. |
| **Multi-machine sync state** | ~~Host DB vs. device manifest~~ | **DECIDED: Host-only for v1** (2026-05-27) | Documented as "one device = one host." Device-side manifest revisited in v2. |
| **Re-encoding existing lossy sources** | ~~Re-encode vs. copy as-is~~ | **DECIDED: Copy as-is by default; opt-in re-encode** (2026-05-27) | Avoids unnecessary quality loss when source is already MP3. |

---

## Sources

- [beets documentation](https://beets.io/) and [source code](https://github.com/beetbox/beets)
- [wrtag](https://github.com/sentriz/wrtag) — Go-based MusicBrainz tagger
- [MusicBrainz Picard](https://github.com/metabrainz/picard)
- [OneTagger](https://github.com/Marekkon5/onetagger) — Rust-based tagger
- [Mutagen documentation](https://mutagen.readthedocs.io/)
- [musicbrainzngs](https://python-musicbrainzngs.readthedocs.io/)
- [pyacoustid](https://github.com/beetbox/pyacoustid)
- [Typer documentation](https://typer.tiangolo.com/)
- [Rich documentation](https://rich.readthedocs.io/)
- [Pluggy documentation](https://pluggy.readthedocs.io/)
- [Dynaconf documentation](https://www.dynaconf.com/)
- [Python Plugin Architecture Guide](https://packaging.python.org/en/latest/guides/creating-and-discovering-plugins/)
- [libmtp](https://github.com/libmtp/libmtp)
- [libgpod](https://github.com/gtkpod/libgpod)
- [libimobiledevice](https://libimobiledevice.org/)
- [beets convert plugin](https://beets.readthedocs.io/en/stable/plugins/convert.html)
- [mtpsync](https://github.com/barsanuphe/mtpsync) — FLAC→MP3 transcode + MTP sync (prior art)
- [FlacSquisher](https://sourceforge.net/projects/flacsquisher/) — Batch FLAC→MP3 mirror tool
- [fre:ac](https://www.freac.org/) — Cross-platform audio converter
- [LAME MP3 encoder](https://lame.sourceforge.io/)
- [FFmpeg](https://ffmpeg.org/)
- [Rockbox](https://www.rockbox.org/) — Open-source firmware for portable music players
- [Flatpak pip generator](https://github.com/flatpak/flatpak-pip-generator)
- [Keep a Changelog](https://keepachangelog.com/)
- [Contributor Covenant](https://www.contributor-covenant.org/)
- [Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0)
- [SPDX License List](https://spdx.org/licenses/)
