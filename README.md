# HKKNX – Home Assistant App

Community-maintained Home Assistant app packaging for [HKKNX by brutella](https://github.com/brutella/hkknx-public), a KNX-to-Apple HomeKit bridge. This is not an official HKKNX or Home Assistant product.

## Features

- Fresh installations, with HKKNX's persistent database and HomeKit pairing data under `/data/hkknx`
- Home Assistant app management and manual updates
- An upstream Docker image digest check every six hours
- Automatic app version bumps and a changelog entry **only** when the image digest changes
- Official release notes included only when a version-tagged Docker image digest can be matched to the corresponding upstream release; otherwise the changelog explicitly states that details are unverified
- Experimental Home Assistant sidebar integration via Ingress (not yet tested with upstream HKKNX)

## Install

1. Open **Home Assistant → Settings → Apps → Install App → ⋮ → Repositories**.
2. Add `https://github.com/Julian0Davinci/ha-hkknx-app`.
3. Install the **HKKNX (Store)** app and start it.
4. Open its web interface, configure the KNX IP interface and set up Apple HomeKit pairing.

## Backups and persistence

The persistent `/data/hkknx` folder survives normal app restarts and updates. For backup/restore of the HKKNX configuration and HomeKit identity, use **HKKNX's own backup and restore function**. Create a backup before uninstalling an app: uninstalling may delete its data.

This app does **not** bundle a custom migration script or import old Home Assistant app data. If moving from another installation, use HKKNX's built-in backup/restore function and ensure that only one bridge is running at a time.

## Changelog and updates

See [`hkknx/CHANGELOG.md`](hkknx/CHANGELOG.md). The [GitHub Actions workflow](.github/workflows/sync-upstream.yml) checks `brutella/hkknx:latest` by immutable digest. A changed digest increments the **packaging** patch version and records the new digest. A matching official upstream release is linked only when verification succeeds. Otherwise the release details are explicitly marked unknown.

The app's version (for example, `0.2.2`) is **not** the upstream HKKNX software version (for example, `3.2.2`).

## Sidebar and support

Ingress has been configured, but its ability to display every part of the HKKNX web UI is not yet verified. If the sidebar view has issues, use the direct web interface on port 8080 and report the issue in this repository.

For HKKNX application bugs, consult [upstream releases and issues](https://github.com/brutella/hkknx-public).
