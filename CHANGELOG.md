# Changelog

All notable changes to this project will be documented in this file.

## [1.0.3] - 2026-07-18

Dependency updates. No user-facing changes; upgrading is optional.

### Changed

- Updated goldmark from v1.8.2 to v1.8.4
- Updated golang.org/x/sys from v0.46.0 to v0.47.0

### Security Assessment

goldmark is the Markdown-to-HTML renderer, so its changes affect gomarkwiki's
output path directly. Two relevant fixes landed between v1.8.2 and v1.8.4:

| Fix | Version | Relevance to gomarkwiki |
|-----|---------|-------------------------|
| Guard against index-out-of-range panic on malformed Markdown ([#556]) | v1.8.3 | Hardens the parser against a crash on adversarial input |
| Disable SVG in `data:image` URLs | v1.8.4 | XSS hardening for the generated HTML |

golang.org/x/sys v0.47.0 contains platform and bug fixes only, with no security
fixes. It is a transitive dependency of fsnotify, used only in `-watch` mode.

No CVEs were assigned to either update.

[#556]: https://github.com/yuin/goldmark/issues/556

## [1.0.2] - 2026-06-15

Dependency updates. No user-facing changes; upgrading is optional.

### Changed

- Upgraded Go from 1.26.3 to 1.26.4
- Updated golang.org/x/sys from v0.44.0 to v0.46.0 (indirect)

## [1.0.1] - 2026-03-23

Dependency updates. No user-facing changes; upgrading is optional.

### Changed

- Upgraded Go from 1.25.6 to 1.26.1
- Updated goldmark from v1.7.16 to v1.7.17
- Updated golang.org/x/sys from v0.40.0 to v0.42.0
- Applied Go 1.26 modernizers via `go fix`

### Security Assessment

Go 1.26.1 includes fixes for five CVEs. None affect gomarkwiki:

| CVE | Package | Reason Not Affected |
|-----|---------|---------------------|
| CVE-2026-27137 | crypto/x509 | gomarkwiki does not use TLS or certificate verification |
| CVE-2026-27138 | crypto/x509 | Same as above |
| CVE-2026-27142 | html/template | gomarkwiki uses html/template but not `<meta http-equiv="refresh">`, the specific attack vector |
| CVE-2026-27139 | os | gomarkwiki does not use `os.Root` |
| CVE-2026-25679 | net/url | gomarkwiki does not use `net/url` |

## [1.0.0] - 2026-03-17

Initial release.
