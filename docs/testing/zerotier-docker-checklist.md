# ZeroTier Docker Manual Test Checklist

This checklist is the Phase 1 manual integration baseline for Tairitsu with a Dockerized ZeroTier controller.

## Environment

- ZeroTier data dir: `build/.docker/zerotier-one`
- Tairitsu data dir: `build/.docker/tairitsu-data`
- Backend binary output: `build/tairitsu`
- Frontend commands: `bun run dev`, `bun run lint`, `bun run build`

## Suggested Commands

### Start ZeroTier container

```bash
docker run -d \
  --name tairitsu-zerotier \
  --restart unless-stopped \
  -p 9993:9993 \
  -v "$(pwd)/build/.docker/zerotier-one:/var/lib/zerotier-one" \
  zerotier/zerotier:latest
```

### Build and run backend

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go build -o ./build/tairitsu ./cmd/tairitsu
./build/tairitsu
```

### Run frontend

```bash
cd web
bun run dev
```

## ZeroTier Controller Prep

Ensure `build/.docker/zerotier-one/local.conf` allows management access. Typical example:

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

Then restart the ZeroTier container.

## Setup Flow Checklist

1. Open the frontend and confirm the app redirects to `/setup`.
2. Confirm non-setup pages are not reachable before initialization.
3. In setup step 1, use the controller URL that matches your environment.
4. If backend runs on host, use token path like `/root/code/tairitsu/build/.docker/zerotier-one/authtoken.secret`.
5. Verify ZeroTier configuration saves successfully.
6. In database setup, confirm SQLite is the only supported path.
7. Leave SQLite path empty once, confirm default path behavior works.
8. Create the initial admin account.
9. Complete setup and confirm the page reloads into initialized mode.

## Runtime Flow Checklist

1. Confirm `/setup` is no longer the main path after initialization.
2. Confirm login works with the created admin account.
3. Confirm incorrect password shows a meaningful backend-aware error.
4. Confirm network list loads successfully.
5. Open a network detail page and confirm metadata loads.
6. Open member list and verify members are shown.
7. Update member authorization and verify success feedback.
8. Edit member metadata and verify persistence.
9. Delete a member only in a disposable test scenario and verify feedback.

## Error Path Checklist

1. Stop the ZeroTier container and confirm runtime requests fail with clear errors.
2. Corrupt the token path during setup and confirm setup reports the backend error.
3. Try accessing runtime APIs before setup and confirm they are blocked.
4. Try calling setup APIs after initialization and confirm they are blocked.

## Notes

- Import Network and Planet remain experimental and are not part of Phase 1 acceptance.
- If Go test or build dependency fetches fail inside sandboxed execution, rerun with an approved external sandbox and `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod`.
