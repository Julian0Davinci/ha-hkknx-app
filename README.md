# HKKNX – Home Assistant App

Run [HKKNX by brutella](https://github.com/brutella/hkknx) directly within Home Assistant.

HKKNX bridges KNX devices to Apple HomeKit, allowing compatible KNX devices to be controlled through the Apple Home app.

This repository provides a community-maintained Home Assistant App package for HKKNX.

## Features

- Runs HKKNX as a Home Assistant App
- Supports KNX-to-HomeKit bridging
- Managed through the Home Assistant App Store
- Persistent configuration and HomeKit pairing data
- App version management and update notifications
- Optional Home Assistant sidebar integration via Ingress (experimental)
- Compatible with supported Home Assistant OS architectures

## Installation

1. Open **Home Assistant → Settings → Apps → Install App**.
2. Open the three-dot menu and select **Repositories**.
3. Add the following repository URL:

   `https://github.com/Julian0Davinci/ha-hkknx-app`

4. Find **HKKNX** in the App Store.
5. Install the app and configure it according to your KNX network.
6. Start the app and open its web interface to configure HKKNX.

## Persistent Data

HKKNX stores its configuration, database, and HomeKit pairing information in the Home Assistant App's persistent `/data` directory.

This data survives:

- App restarts
- Home Assistant restarts
- Docker image replacements during App updates
- Normal App upgrades

**Important:** Uninstalling the app can remove its associated data. Create a Home Assistant backup before uninstalling or migrating between installations.

Existing installations with a different App ID require a separate data migration.

## Updates

The repository is designed to monitor the upstream HKKNX Docker image and publish new App versions when changes are detected.

Once a newer App version is published, Home Assistant can offer an update through its App management interface.

Updates are installed manually unless automatic App updates are enabled.

## Home Assistant Sidebar

The App includes experimental Home Assistant Ingress support.

Depending on the upstream HKKNX web interface, additional proxy adjustments may be necessary for full sidebar functionality.

The standalone web interface remains an alternative.

## Credits

- [brutella/hkknx](https://github.com/brutella/hkknx) – Original HKKNX project
- [Home Assistant](https://www.home-assistant.io/) – Home automation platform

This repository provides community packaging for Home Assistant and is not an official HKKNX or Home Assistant project.

## Support

For problems related to the Home Assistant App packaging, please open an issue in this repository.

For issues related to HKKNX itself, please refer to the original upstream project.