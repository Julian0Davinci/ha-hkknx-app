# HKKNX for Home Assistant

This community app runs the official upstream HKKNX Docker image. HKKNX and HomeKit pairing data persist in `/data/hkknx` and survive app updates. Create an app backup before updating or uninstalling.

## Direct web UI

HKKNX's native web server runs on port **8080** with `host_network: true`. Use the app's **Web UI** button or `http://<HOME_ASSISTANT_IP>:8080` on your local network.

## Home Assistant sidebar (experimental)

The app runs a separate small reverse proxy on port **8099** for Home Assistant Ingress. It uses the request header `X-Ingress-Path` and fixes root-relative redirects like `/projects/...`, common static resource paths, and cookie paths. Supervisor must be able to connect to this port. Do not expose port 8099 through a router or independent reverse proxy.

Important: HKKNX does not natively support subpath hosting. Dynamically constructed URLs or client-side navigation may still fail even with the proxy. If this happens, use the **direct web UI on port 8080** and report the failing request path; do not modify the HKKNX database.

## Uploading large KNX projects

Supervisor normally buffers Ingress requests and rejects uploads larger than 16 MiB with `Maximum request body size 16777216 exceeded`. This app enables `ingress_stream: true` (since 0.2.6), forwarding upload bodies as a stream through the Go reverse proxy. Home Assistant may send requests with chunked transfer encoding and no `Content-Length`, which the Go HTTP server supports.

If the web upload still fails because of a separate HKKNX application limit or client behavior, use the direct web UI on port 8080. The app's existing database and HomeKit pairings are untouched.

## Install / restore

Start the app, configure HKKNX, and pair it with Apple Home. If moving from a previous installation, use HKKNX's **built-in backup and restore** feature. Only one HKKNX bridge instance should run at a time. This app has no custom migration code.

## Updates

The GitHub Action checks the upstream image digest every six hours. On digest changes it increments the packaging version and writes changelog entries; it includes verified release notes when available. The app is not updated automatically unless you have enabled automatic app updates.
