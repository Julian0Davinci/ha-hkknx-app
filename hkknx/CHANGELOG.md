# Changelog

## 0.2.3 — Ingress adapter (experimental)
- Added a separate Home Assistant Ingress adapter on port `8099`, leaving HKKNX's native interface on port `8080` unchanged.
- Rewrites root-relative HTTP redirects such as `/projects/...` so they stay inside the authenticated Home Assistant Ingress path.
- Best-effort rewriting of root-relative HTML/CSS/JavaScript links and root-scoped cookies.
- Uses `X-Ingress-Path` dynamically. The proxy accepts only Supervisor or loopback connections.
- Ingress remains experimental until the HKKNX interface is tested end to end. Direct access on port 8080 is the fallback.
- Keeps the existing persistent `/data/hkknx` database and all HomeKit pairings; no migration, reset, or import code.

## 0.2.2 — Home Assistant app packaging
- Fresh installations are now supported: HKKNX creates its own persistent database in `/data/hkknx`.
- Removed the custom migration/import mechanism and the `/share` mount. To transfer an existing HKKNX configuration, use HKKNX's own backup and restore feature.
- The update workflow can now append upstream release notes when it can reliably match an official versioned Docker tag to the updated image digest. Otherwise it reports the image update without claiming unverified changes.
- The Home Assistant sidebar (Ingress) remains experimental until tested with the upstream web UI.

## 0.2.1 — Initial upstream image pin
- Recorded the Docker image digest to enable automated checks. The first pin was not evidence of a new upstream release.

## 0.2.0 — Initial community package
- Initial Home Assistant app packaging with persistent `/data` and experimental Ingress configuration.
