# Tairitsu

![Tairitsu banner](assets/readme-banner.png)

[**简体中文**](readme-i18n/README.zh-CN.md)

Tairitsu is a self-hosted web console for a separately installed ZeroTier controller. It provides a web UI for setup, user accounts, network and member management, network settings, and controller network import.

## Features

- Setup wizard for controller connection, SQLite, and initial admin creation
- User registration, login, password change, session handling, admin transfer, admin-created users, password reset, and user deletion
- Network creation, detail view, and member approval / rejection / removal
- IPv4, IPv6, managed routes, DNS, and multicast settings
- Import takeover for controller networks not yet owned in Tairitsu
- Experimental `Planet` generation

## Installation

Prerequisites:

- A ZeroTier controller that Tairitsu can reach
- Access to the controller token file
- A persistent directory for Tairitsu data
- Container runtime such as Docker or Podman

Prepare the ZeroTier controller so Tairitsu can access the local controller API. Edit `/var/lib/zerotier-one/local.conf`:

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

Restart the ZeroTier controller after changing `local.conf`.

Run Tairitsu with Docker:

```bash
docker run -d \
  --name tairitsu \
  --restart unless-stopped \
  -p 3000:3000 \
  -v /var/lib/zerotier-one:/var/lib/zerotier-one \
  -v /path/to/tairitsu-data:/app/data \
  ghcr.io/gt-610/tairitsu:latest
```

Or use Docker Compose:

```yaml
services:
  tairitsu:
    image: ghcr.io/gt-610/tairitsu:latest
    container_name: tairitsu
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - /var/lib/zerotier-one:/var/lib/zerotier-one
      - /path/to/tairitsu-data:/app/data
```

Open `http://<host>:3000` and complete the setup wizard:

1. Enter the ZeroTier controller URL.
2. Enter the token file path.
3. Configure the SQLite database path if needed.
4. Create the initial administrator account.

Sessions are held in an HttpOnly cookie. Serve Tairitsu over HTTPS in any
deployment reachable beyond a trusted network; on plain HTTP the cookie cannot
carry the `Secure` flag and will also travel unencrypted.

## Documentation

- [Operations](docs/OPERATIONS.md)

Tairitsu serves its web console and its HTTP surface from a single process. The
frontend and server are version-locked and ship in one binary, so the internal
endpoints between them are not a supported public API. See the
[internal HTTP reference](docs/api/Tairitsu_API_Documentation.md) if you need to
read it; expect it to change in any release.

## Development

Prerequisites:

- Go 1.26 or later with CGO enabled
- Node.js or Bun latest
- A local or Dockerized ZeroTier controller

Build frontend and backend for a single-binary release:

```bash
cd web
bun install
bun run build
cd ..

# The frontend bundle is embedded, so build it before the server.
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go build -o ./build/tairitsu ./cmd/tairitsu
./build/tairitsu   # serves the console and the API on :3000
```

The listening port is fixed at 3000 and is not configurable; in a container,
decide the port users reach with your own mapping (for example `-p 8080:3000`).

For frontend development with hot reload, run the server and the Vite dev
server together. The dev server listens on 3001 and proxies `/api` to the
server on 3000:

```bash
# terminal 1
go run ./cmd/tairitsu

# terminal 2
cd web
bun run dev
```

## License

[GNU GPL v3](LICENSE)

## Legal Notice

Since version 1.16.0, ZeroTier's controller component is licensed under a [commercial source-available non-free license](https://github.com/zerotier/ZeroTierOne/blob/main/nonfree/LICENSE.md).

Tairitsu does not redistribute any ZeroTier controller code. It talks to a **separately installed** ZeroTier controller through the official API and stays outside ZeroTier's non-free controller codebase.

The `Generate Planet` feature is modified from [ztnodeid-go](https://github.com/kmahyyg/ztnodeid-go) under GNU GPL v3 License, **not** from ZeroTier itself.

Tairitsu is **not** a ZeroTier product and is **not affiliated with, endorsed by, or supported by ZeroTier, Inc.**
