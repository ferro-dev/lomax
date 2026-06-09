# Security Policy

## Supported versions

`lomax` is pre-1.0 and pre-release. Until a `0.1.0` tag exists, only the `main`
branch is supported. Once releases begin, this section will list which versions
receive security fixes.

## Reporting a vulnerability

**Do not open a public issue for security vulnerabilities.**

Use GitHub's private vulnerability reporting:

1. Go to the [Security tab](https://github.com/ferro-dev/lomax/security/advisories).
2. Click **Report a vulnerability**.
3. Provide a description, reproduction steps, affected version/commit, and
   impact.

You can expect an initial acknowledgement within a few days. Once the report is
triaged, we will coordinate a fix and a disclosure timeline with you.

## Scope

Relevant concerns for a local CLI music manager include, but are not limited to:

- Path traversal when reading source files or writing to library/mirror/device
  paths.
- Command injection via tag values, filenames, or config passed to shelled-out
  tools (`ffmpeg`, `fpcalc`, `metaflac`).
- Unsafe handling of untrusted metadata returned by remote sources
  (MusicBrainz, AcoustID, Discogs, Last.fm).
- Plugin subprocess trust boundaries (the `go-plugin` gRPC interface).

## Out of scope

- Vulnerabilities in third-party system binaries themselves (report those
  upstream to FFmpeg, Chromaprint, etc.).
- Issues that require an already-compromised local account.
