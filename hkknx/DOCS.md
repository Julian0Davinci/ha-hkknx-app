# HKKNX — Store-App

Die Datenbank und HomeKit-Pairings liegen persistent unter `/data/hkknx`. Das private `/data`-Volume bleibt bei App-Updates erhalten und wird in App-Backups aufgenommen.

## Erste Installation / Umstellung von local_hkknx

Diese App ist absichtlich **migrationsgeschützt**: beim ersten Start ohne vorhandene Daten muss `/share/hkknx-migration/.ready` mit dem Datenauszug der alten App existieren. Ansonsten startet HKKNX nicht, damit keine zweite HomeKit-Bridge-ID entsteht.

1. Alte App `local_hkknx` sichern.
2. In der alten App einmalig den mitgelieferten Export-Patch einsetzen und starten; er legt `/share/hkknx-migration` an.
3. Alte App stoppen.
4. Neue Store-App installieren und starten. Sie kopiert die Daten **einmalig** nach `/data/hkknx`.
5. KNX und HomeKit testen, erst danach den Transferordner löschen. Alte App als Rückfallmöglichkeit zunächst behalten.

## Seitenleiste

Nach erfolgreichem Start auf der App-Info-Seite „In Seitenleiste anzeigen“ aktivieren. Ingress erwartet die Weboberfläche auf Port 8080. Der HKKNX-Webserver wurde nicht speziell auf einen Ingress-Unterpfad angepasst. Falls CSS, JavaScript oder WebSocket-Verbindungen fehlschlagen, ist ein Ingress-kompatibler Reverse-Proxy als spätere Anpassung nötig; bis dahin die direkte Weboberfläche verwenden.

## Updates

Eine GitHub Action prüft alle 6 Stunden das Digest des Docker-Hub-Tags `brutella/hkknx:latest`. Bei Änderung wird das Dockerfile auf das Digest festgelegt und die App-Patchversion hochgezählt. Im Home-Assistant-App-Store wird anschließend nach dem nächsten Repository-Refresh ein Update angezeigt. Keine automatischen Installationen ohne Test.
