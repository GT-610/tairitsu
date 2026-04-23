# Installation and Deployment

Tairitsu is published for **single-instance self-hosted deployments**. The main supported path is Docker / Podman with a separately installed ZeroTier controller.

## What You Need

- A ZeroTier controller already running on the same host or reachable network
- Access to the controller token file
- A persistent directory for Tairitsu application data
- The published `linux/amd64` container image: `ghcr.io/gt-610/tairitsu:latest`

## ZeroTier Controller Preparation

Tairitsu talks to the local controller API. That API must be reachable from the Tairitsu process.

Prepare `local.conf` in the ZeroTier data directory, usually `/var/lib/zerotier-one/local.conf`:

```json
{
  "settings": {
    "allowManagementFrom": [
      "0.0.0.0/0",
      "::/0"
    ]
  }
}
```

If you prefer a narrower rule, replace those entries with the CIDR that covers the Tairitsu host or container network.

After changing `local.conf`, restart the ZeroTier controller.

## Docker / Podman Run

```bash
docker run -d \
  --name tairitsu \
  --restart unless-stopped \
  -p 3000:3000 \
  -v /var/lib/zerotier-one:/var/lib/zerotier-one \
  -v /path/to/tairitsu-data:/app/data \
  ghcr.io/gt-610/tairitsu:latest
```

The first mount exposes the ZeroTier token and controller-side files. The second mount persists the Tairitsu SQLite database and runtime data.

## First-Run Setup

Open `http://<host>:3000` and complete the setup wizard.

Recommended values:

1. **Controller URL**
   - Host deployment: `http://127.0.0.1:9993`
   - Container-to-container deployment: the URL that reaches the controller from inside the Tairitsu container
2. **Token path**
   - Host-style mount example: `/var/lib/zerotier-one/authtoken.secret`
3. **Database**
   - Type: `sqlite`
   - Path: keep the default unless you have a specific reason to change it
4. **Initial admin**
   - Create the first administrator account

After setup completes, Tairitsu switches into runtime mode and `/setup` is no longer the main entrypoint.

## Upgrade and Backup Notes

- Back up the Tairitsu data directory before upgrading.
- Back up the ZeroTier controller data directory separately if the controller and Tairitsu share the same host.
- Upgrades are expected to preserve the current SQLite-first runtime model; there is no public migration guide for switching database engines because MySQL and PostgreSQL are not supported today.

## Public Validation Checklist

A public deployment should verify at least:

1. The container starts and serves the UI on port `3000`.
2. The setup wizard can reach the ZeroTier controller.
3. SQLite setup succeeds.
4. The first admin can log in after initialization.
5. An owned network can be created and opened.
6. A pending member can be approved or removed.

For product boundaries and operational caveats, see [Operations and Support Boundaries](OPERATIONS.md).
For the release-minded regression checklist, see [Validation Checklist](VALIDATION.md).
