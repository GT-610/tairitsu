# Tairitsu

[**简体中文**](readme-i18n/README.zh-CN.md)

Tairitsu is a self-hosted web console for a separately installed ZeroTier controller. It focuses on the day-to-day management path: setup, login and registration, network and member administration, network settings, import takeover, and user governance.

Tairitsu is aimed at **SQLite-backed, single-instance self-hosted deployments**. It is not trying to be a broader multi-tenant platform like `ztnet`, and `Planet` generation remains experimental.

## Current Scope

- Setup wizard for controller connection, SQLite, and initial admin creation
- User registration, login, password change, session handling, admin transfer, admin-created users, password reset, and user deletion
- Owner-scoped network management and member approval
- IPv4, IPv6, managed routes, DNS, and multicast settings
- Import takeover for controller networks not yet owned in Tairitsu

## Access Model

- Each ZeroTier network has exactly one `owner` in Tairitsu.
- The network owner can manage that network and its members.
- Platform `admin` users manage setup, user governance, and network import.
- Platform `admin` does not automatically gain read/write access to every owned network.

## Deployment

The main public deployment path is Docker / Podman:

- Image: `ghcr.io/gt-610/tairitsu:latest`
- Published image target: `linux/amd64`
- Controller requirement: a separately installed ZeroTier controller with API access enabled through `local.conf`

Detailed docs:

- [Installation and Deployment](docs/INSTALLATION.md)
- [Operations and Support Boundaries](docs/OPERATIONS.md)
- [API documentation](docs/api/Tairitsu_API_Documentation.md)

Manual host installation is possible for development, but Docker / Podman is the public deployment path documented today.

## Development

Prerequisites:

- Go 1.25 or later with CGO enabled
- Bun 1.3 or later
- A local or Dockerized ZeroTier controller

Build backend:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go build -o ./build/tairitsu ./cmd/tairitsu
```

Run frontend:

```bash
cd web
bun install
bun run dev
```

Release baseline:

- `go test ./...`
- `bun test`
- `bun run lint`
- `bun run build`
- `docker build .`

## License

[GNU GPL v3](LICENSE)

## Legal Notice

Tairitsu does not redistribute any ZeroTier controller code. It talks to a **separately installed** ZeroTier controller through the official API and stays outside ZeroTier's non-free controller codebase.

The `Generate Planet` feature is modified from [ztnodeid-go](https://github.com/kmahyyg/ztnodeid-go) under GNU GPL v3 License, **not** from ZeroTier itself.

Tairitsu is **not** a ZeroTier product and is **not affiliated with, endorsed by, or supported by ZeroTier, Inc.**
