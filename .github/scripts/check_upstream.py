#!/usr/bin/env python3
"""Track the upstream multi-platform Docker image digest and bump the HA app patch version."""
import json
from pathlib import Path
import re
from urllib.request import Request, urlopen

ROOT = Path(__file__).resolve().parents[2]
CONFIG = ROOT / 'hkknx' / 'config.yaml'
DOCKERFILE = ROOT / 'hkknx' / 'Dockerfile'
PIN = ROOT / '.upstream-digest'
URL = 'https://hub.docker.com/v2/repositories/brutella/hkknx/tags/latest/'


def main():
    request = Request(URL, headers={'User-Agent': 'homeassistant-hkknx-image-monitor/1.0'})
    with urlopen(request, timeout=30) as response:
        data = json.load(response)
    digest = data.get('digest', '')
    if not re.fullmatch(r'sha256:[a-f0-9]{64}', digest):
        raise RuntimeError('Docker Hub lieferte keinen gueltigen Digest; nichts aendern')
    prior = PIN.read_text().strip() if PIN.is_file() else None
    if prior == digest:
        print('Keine Aenderung am Image:', digest)
        return

    cfgtxt = CONFIG.read_text(encoding='utf-8')
    m = re.search(r'(?m)^version: "(\d+)\.(\d+)\.(\d+)"$', cfgtxt)
    if m is None:
        raise RuntimeError('Keine Patch-Version x.y.z in config.yaml')
    a,b,c = map(int,m.groups())
    newversion = f'{a}.{b}.{c+1}'
    df = DOCKERFILE.read_text(encoding='utf-8')
    out, hits = re.subn(r'(?m)^FROM brutella/hkknx(?::[^\s]+|@sha256:[a-f0-9]{64})$',
                        'FROM brutella/hkknx@'+digest, df)
    if hits != 1:
        raise RuntimeError('Unbekanntes FROM-Format; nichts aendern')
    # Erst nach vollstaendiger Validierung schreiben.
    cfgtxt = cfgtxt[:m.start()] + f'version: "{newversion}"' + cfgtxt[m.end():]
    DOCKERFILE.write_text(out, encoding='utf-8')
    CONFIG.write_text(cfgtxt, encoding='utf-8')
    PIN.write_text(digest+'\n', encoding='utf-8')
    print(f'Digest gewechselt: {prior} -> {digest}, App-Version {newversion}')


if __name__ == '__main__':
    main()
