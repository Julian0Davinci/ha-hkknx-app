#!/usr/bin/env python3
"""Bump the HA app only on a new upstream Docker image digest.

Release notes are attached only after verifying that a Docker Hub version tag
has EXACTLY the latest image digest and that GitHub has that exact release tag.
If not verifiable, the changelog explicitly states that the changes are unknown.
"""
from datetime import datetime, timezone
import json
from pathlib import Path
import re
from urllib.parse import quote
from urllib.request import Request, urlopen

ROOT = Path(__file__).resolve().parents[2]
CONFIG = ROOT / "hkknx/config.yaml"
DOCKERFILE = ROOT / "hkknx/Dockerfile"
CHANGELOG = ROOT / "hkknx/CHANGELOG.md"
PIN = ROOT / ".upstream-digest"
HUB = "https://hub.docker.com/v2/repositories/brutella/hkknx/tags/"
RELEASES = "https://api.github.com/repos/brutella/hkknx-public/releases/tags/"


def fetch_json(url):
    request = Request(url, headers={"User-Agent": "hkknx-ha-app-updates/1.0", "Accept": "application/json"})
    with urlopen(request, timeout=20) as response:
        return json.load(response)


def matching_release(digest):
    """Conservatively find exact release notes for a digest-matched Docker tag."""
    # Docker Hub tag listing is paginated. Limit to first 100 recent tags.
    url = HUB + "?page_size=100&ordering=last_updated"
    try:
        tags = fetch_json(url).get("results", [])
        candidates = []
        for tag in tags:
            version = tag.get("name", "")
            if tag.get("digest") != digest:
                continue
            if re.fullmatch(r"v?\d+\.\d+\.\d+(?:[-.]b\d+)?", version):
                candidates.append(version)
        # Prefer a stable release over prereleases if multiple tags match.
        candidates.sort(key=lambda v: ("-" in v or ".b" in v, v), reverse=False)
        for tag in candidates:
            release = fetch_json(RELEASES + quote(tag, safe=""))
            if release.get("tag_name") == tag and release.get("body", "").strip():
                return tag, release["body"].strip()
    except (OSError, ValueError, KeyError, TypeError) as err:
        print(f"Release-note lookup unavailable: {err}")
    return None


def changelog_entry(version, digest):
    stamp = datetime.now(timezone.utc).strftime("%Y-%m-%d")
    lines = [f"## {version} — {stamp}", "- Updated upstream Docker image `brutella/hkknx:latest`.", f"- Image digest: `{digest}`."]
    matched = matching_release(digest)
    if matched:
        tag, body = matched
        lines.append(f"- Official [HKKNX {tag} release notes](https://github.com/brutella/hkknx-public/releases/tag/{quote(tag, safe='')}), verified against a Docker tag with the same digest:")
        # GitHub Markdown release notes: quote line-by-line as blockquote.
        for line in body.splitlines()[:100]:
            lines.append("> " + line if line.strip() else ">")
    else:
        lines.append("- Upstream release details could not be verified for this image. No specific features or fixes are claimed.")
    return "\n".join(lines) + "\n\n"


def main():
    digest = fetch_json(HUB + "latest/").get("digest", "")
    if not re.fullmatch(r"sha256:[a-f0-9]{64}", digest):
        raise RuntimeError("Docker Hub returned no valid image digest")
    prior = PIN.read_text().strip() if PIN.exists() else None
    if digest == prior:
        print("No upstream change:", digest)
        return
    conf = CONFIG.read_text(encoding="utf-8")
    match = re.search(r'(?m)^version: "(\d+)\.(\d+)\.(\d+)"$', conf)
    if not match:
        raise RuntimeError("Missing app version x.y.z")
    major, minor, patch = map(int, match.groups())
    new_version = f"{major}.{minor}.{patch + 1}"
    docker = DOCKERFILE.read_text(encoding="utf-8")
    replaced, count = re.subn(r'(?m)^FROM brutella/hkknx(?::[^\s]+|@sha256:[a-f0-9]{64})$', "FROM brutella/hkknx@" + digest, docker)
    if count != 1:
        raise RuntimeError("Unexpected Dockerfile FROM")
    changelog = CHANGELOG.read_text(encoding="utf-8") if CHANGELOG.exists() else "# Changelog\n"
    if not changelog.startswith("# "):
        raise RuntimeError("Unexpected changelog header")
    entry = changelog_entry(new_version, digest)
    header, _, previous = changelog.partition("\n")
    updated_log = header + "\n\n" + entry + previous.lstrip("\n")
    # Write only after the Docker digest, version, and changelog are all validated.
    DOCKERFILE.write_text(replaced, encoding="utf-8")
    CONFIG.write_text(conf[:match.start()] + f'version: "{new_version}"' + conf[match.end():], encoding="utf-8")
    CHANGELOG.write_text(updated_log, encoding="utf-8")
    PIN.write_text(digest + "\n", encoding="utf-8")
    print(f"Updated app to {new_version}; upstream digest {prior} -> {digest}")


if __name__ == "__main__":
    main()
