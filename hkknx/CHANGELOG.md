# Changelog

## 0.2.9 — Fix HKKNX WebSocket Origin verification
- HKKNX rejected Ingress WebSockets with HTTP 403 because the client Origin (Home Assistant on port 8123) did not match the proxied upstream Host (127.0.0.1:8080).
- Normalize only WebSocket upgrade request Origin to the local upstream origin, after authenticated Supervisor Ingress access has been verified.
- Keep the client's normal HTTP Origin untouched; preserve existing frontend adaptations, port 8080, the database, and HomeKit pairings.
- Adds regression tests for both WebSocket and ordinary HTTP Origin handling.

## 0.2.8 — Safe JavaScript and WebSocket diagnostics
- Stop rewriting bundled JavaScript files to prevent corruption of regex literals and minified scripts.
- Support HTMX ws-connect values using ws:/ and wss:/ prefixes.
- Log WebSocket upgrade attempts and handshake outcomes without sensitive headers or tokens.
- Preserve HKKNX port 8080, Ingress port 8099, database and HomeKit pairing data.

## 0.2.7 — Live device status over Ingress WebSockets
- Fixes HKKNX live accessory status streams staying on `Laden` because the HTMX WebSocket connection was opened at `ws://<ha-host>:8123/projects/.../hx-ws-accessory-states` without the Ingress session prefix.
- Adds a scoped WebSocket constructor adapter before HKKNX scripts. It rewrites same-origin `ws://` or `wss://` URLs under `/projects/` or `/api/`, and leaves external/already-prefixed destinations alone.
- Also prefixes root-relative `ws-connect` and `data-ws-connect` markup attributes. The existing Go reverse proxy supports WebSocket upgrade requests.
- No change to HKKNX port 8080, Ingress port 8099, database, KNX communication or HomeKit pairings.
- Verify after updating: device status values should replace `Laden` through the Home Assistant sidebar.

## 0.2.6 — Large KNX project uploads through Ingress
- Enables Supervisor `ingress_stream: true` to allow KNX project uploads larger than the default 16 MiB body limit.
- The existing Go reverse proxy forwards streamed request bodies to the unchanged native HKKNX web server on port 8080.
- Authentication, HTMX URL rewriting, persistent database and HomeKit pairings remain unchanged.
- Please verify the upload in Home Assistant Ingress; if the application itself has a separate file-size limit, use the native web UI.

## 0.2.5 — Ingress HTMX request routing (experimental)
- Corrects root-relative `hx-get`, `hx-post`, other HTMX request attributes and their `data-hx-*` variants to remain inside the Ingress session path.
- Extends the fetch/XMLHttpRequest Ingress shim to cover dynamically generated `/projects/...` requests as well as `/api/...`.
- Addresses the reported diagram-fragment 404 and login URL failures. Requires live verification after installation.
- Leaves the native HKKNX service on port 8080, the Ingress proxy on 8099, and all existing HomeKit and database data unchanged.

## 0.2.4 — Ingress login API routing (experimental)
- Corrects failed fresh logins through Home Assistant Ingress when HKKNX calls `/api/validate-credentials` on the Home Assistant origin instead of the HKKNX Ingress endpoint.
- Injects a minimal client-side shim before HKKNX scripts to route same-origin `/api/...` calls made using `fetch` or `XMLHttpRequest` through the active `X-Ingress-Path` prefix.
- Leaves other origins, already-prefixed requests, HKKNX port 8080, data and HomeKit pairings unchanged.
- Needs a fresh-session login test in Safari private mode and the Home Assistant iOS app.

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
