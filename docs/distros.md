# Supported Distros & System Dependencies

`lomax` is Linux-only by design (see
[Goals & Non-Goals](music-cli-plan.md#2-goals--non-goals)). This page covers
which distros are supported and how to install the system binaries `lomax`
shells out to.

## Support tiers

| Distro | Tier | Notes |
|--------|------|-------|
| Ubuntu 22.04 LTS, 24.04 LTS | 1 | CI runs both on every PR |
| Debian 12 (Bookworm) | 1 | Stable baseline; CI runs |
| Fedora 39+ | 1 | Modern reference distro |
| Arch Linux (rolling) | 1 | AUR distribution path |
| openSUSE Tumbleweed | 2 | Best-effort; nightly CI only |
| RHEL / Rocky / Alma 9 | 2 | Older glibc; known-good, not CI-gated |
| Alpine Linux | 2 | musl libc; community-supported |
| Ubuntu 20.04, RHEL 8, Debian 11 | — | Unsupported (pre-glibc 2.34) |

**System baseline:** kernel 5.15+, glibc 2.35+. systemd is not assumed (`lomax`
is a CLI, not a daemon). Wayland vs X11 is irrelevant — it's a terminal program.

## System binaries

`lomax` and its first-party plugins shell out to a few native tools rather than
bundling them (every distro packages them, and bundling FFmpeg is a
licensing minefield). `lomax` checks for these at startup and prints an install
hint if one is missing.

| Tool | Required by | Purpose |
|------|-------------|---------|
| `ffmpeg` | transcode plugin | Decode/encode for mirror transcoding |
| `fpcalc` (Chromaprint) | metadata resolver | AcoustID audio fingerprinting |
| `metaflac` (optional) | FLAC tag operations | Fallback FLAC tag writer |

### Install commands

**Debian / Ubuntu**
```bash
sudo apt update
sudo apt install ffmpeg libchromaprint-tools flac
```

**Fedora** (FFmpeg needs [RPM Fusion](https://rpmfusion.org/) enabled)
```bash
sudo dnf install ffmpeg chromaprint-tools flac
```

**Arch Linux**
```bash
sudo pacman -S ffmpeg chromaprint flac
```

**openSUSE Tumbleweed** (FFmpeg via Packman)
```bash
sudo zypper install ffmpeg chromaprint-fpcalc flac
```

**Alpine Linux**
```bash
sudo apk add ffmpeg chromaprint flac
```

## Installing lomax

> **Status:** pre-release. No published packages yet. The paths below are the
> planned [distribution channels](music-cli-plan.md#12-distribution-plan); they
> become live starting at Milestone 6.

| Method | Command | Availability |
|--------|---------|--------------|
| Go toolchain | `go install github.com/ferro-dev/lomax@latest` | M6 |
| GitHub Releases | Download `linux-amd64` / `linux-arm64` binary | M6 |
| Flatpak (Flathub) | `flatpak install flathub <app-id>` | Phase 2 |
| AUR | `<aur-helper> -S lomax` | Phase 3 |
| Fedora COPR | `sudo dnf copr enable <owner>/lomax && sudo dnf install lomax` | Phase 3 |
| Debian/Ubuntu PPA | via Launchpad | Phase 3 |

### Flatpak note

The Flatpak build pulls FFmpeg and Chromaprint in as Flatpak-side dependencies,
so the system-binary install step above is **not** needed under Flatpak. The
sandbox does impose filesystem-permission requirements (library path, mounted
devices); those are documented separately and reflected in the published
manifest.
