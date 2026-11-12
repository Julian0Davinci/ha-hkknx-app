# HKKNX for Home Assistant

This community app runs the upstream HKKNX Docker image. The application database and HomeKit pairings are stored persistently in `/data/hkknx` and survive app restarts and updates. Back up the app before uninstalling.

## Installation

1. Install HKKNX from this custom Home Assistant app repository.
2. Start the app and open the HKKNX web interface on port 8080.
3. Configure the KNX IP interface and pair the bridge with Apple Home in HKKNX.

## Restoring an existing HKKNX configuration

Use HKKNX's **built-in backup and restore** function in the HKKNX web interface. Stop the old HKKNX instance before starting the new bridge to avoid duplicate HomeKit advertisements. No custom migration folder or script is used by this app.

## Sidebar

Ingress on port 8080 is configured, but compatibility of the upstream HKKNX web interface with Ingress path rewriting is not yet verified. If its pages/assets fail, use the standalone web UI and open an issue.

## Updates and changelog

The GitHub Action checks the upstream Docker Hub `latest` digest every six hours. If it changes, the app version is bumped and the changelog receives release notes only when a matching version-tag digest and upstream GitHub release are found. Otherwise the entry explicitly says the upstream changes have not been independently identified. App updates are not installed automatically unless configured by the user.
