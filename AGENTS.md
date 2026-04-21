## Build instructions
1. Go Build binary should be in `./build` folder. DON'T BUILD ELSEWHERE.
2. Frontend uses `bun`, not `npm`/`node`. Use `bun install`, `bun run dev`, `bun run lint`, and `bun run build`.
3. Frontend dev server has hot reload, no need to rerun it after source changes.

## Docker test environment
1. Docker-related local test data lives under `build/.docker`.
2. ZeroTier container data directory is `build/.docker/zerotier-one`.
3. Tairitsu local persistent data directory is `build/.docker/tairitsu-data`.
4. Do not place ad-hoc Docker test data outside `build/.docker` unless explicitly requested.

## ZeroTier container testing
1. Prefer mounting `build/.docker/zerotier-one` to `/var/lib/zerotier-one` in the ZeroTier container.
2. If Tairitsu backend runs on the host, the setup wizard's ZeroTier token path should use the host path, e.g. `/root/code/tairitsu/build/.docker/zerotier-one/authtoken.secret`.
3. If Tairitsu runs in Docker on the same network as ZeroTier, the setup wizard can use container paths like `/var/lib/zerotier-one/authtoken.secret` and service URL like `http://zerotier:9993`.
4. For local controller API access from Tairitsu, `build/.docker/zerotier-one/local.conf` should allow management access, e.g. via `allowManagementFrom`.
