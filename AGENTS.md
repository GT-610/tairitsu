## Build instructions
1. Go Build binary should be in `./build` folder. DON'T BUILD ELSEWHERE.
2. Frontend uses `bun`, not `npm`/`node`. Use `bun install`, `bun run dev`, `bun run lint`, and `bun run build`.
3. Frontend dev server has hot reload, no need to rerun it after source changes.

## Code writing expectations
1. Prefer the best maintainable end state over superficial uniformity. Do not refactor only to make styles look the same.
2. Frontend and backend are version-locked and shipped together. Do not keep legacy compatibility layers, old payload fallbacks, or duplicate request/response formats unless explicitly requested.
3. Keep responsibilities clear:
   - Backend owns protocol adaptation, ZeroTier field compatibility, permission rules, response shaping, and stable API semantics.
   - Frontend owns presentation, interaction state, and UI-only derived text.
   - Do not let the frontend guess semantics that the backend can return directly.
4. Only show information that is trustworthy:
   - If data can be derived from the official API, it may be displayed.
   - If a value is missing, zero-valued, or not reliably available, hide it or mark it unavailable instead of fabricating a plausible-looking display.
   - Do not render zero-value timestamps or pseudo activity times as real user-facing times.
5. Remove dead code, obsolete routes, unused compatibility helpers, and duplicated UI flows when they no longer serve the current product path.
6. Avoid placeholder product UI:
   - Do not add fake entry points, fake toggles, or “coming soon” interactions that look usable.
   - Experimental features must be clearly marked as experimental.
7. When cleaning code, prefer deleting obsolete logic over adding another abstraction layer around it.
8. Keep public docs user-facing. Internal notes, audits, temporary plans, and work-in-progress materials should stay out of public docs when they must be kept locally.

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
