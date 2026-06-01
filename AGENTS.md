## Build
- Go binary output: `./build` folder only.
- Frontend: use `bun`, `node`, or `deno` — no runtime-specific APIs are used. Pick whichever is available.
- Frontend dev server has hot reload; no restart needed after source changes.

## Code expectations
- Maintainable end state over superficial uniformity.
- Frontend and backend are version-locked and shipped together; no legacy compatibility layers unless explicitly requested.
- Responsibilities: backend owns protocol adaptation, ZeroTier field compatibility, permissions, response shaping, and stable API semantics. Frontend owns presentation, interaction state, and UI-only derived text.
- Only display trustworthy data. Hide or mark unavailable anything missing, zero-valued, or unreliable. Never fabricate timestamps or pseudo activity.
- Remove dead code, obsolete routes, unused helpers, and duplicated UI flows when they no longer serve the product path.
- Prefer deleting obsolete logic over wrapping it in another abstraction.
- No fake toggles, placeholder entry points, or "coming soon" UI. Experimental features must be clearly labeled.
- Keep public docs user-facing; internal notes and WIP materials stay out of public docs.

## Docker test environment
- All local Docker test data: `build/.docker/`
- ZeroTier data: `build/.docker/zerotier-one/`
- Tairitsu persistent data: `build/.docker/tairitsu-data/`
- ZeroTier container: mount `build/.docker/zerotier-one` → `/var/lib/zerotier-one`
- Host Tairitsu backend: use host path for ZeroTier token, e.g. `.../build/.docker/zerotier-one/authtoken.secret`
- Docker Tairitsu + ZeroTier on same network: use container paths, e.g. `http://zerotier:9993`
- Local controller API access: ensure `build/.docker/zerotier-one/local.conf` has `allowManagementFrom` configured
