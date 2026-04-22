# Tairitsu Phase 1 Audit

## Scope

Phase 1 targets a stable SQLite-first, single-instance, self-hosted ZeroTier controller UI. The goal is to reduce architectural ambiguity and freeze incomplete surfaces before adding new features.

## Immediate Blockers

### 1. Setup and runtime boundaries were previously mixed

- Severity: Critical
- Fix benefit: High
- Status: Mitigated in current refactor

Previously, setup endpoints and runtime endpoints were registered together, allowing login, registration, and setup operations to coexist without a strict state boundary. This made it possible for uninitialized systems to expose runtime routes and for initialized systems to keep exposing setup-only actions.

Current direction:

- Uninitialized mode exposes only setup-related endpoints.
- Initialized mode exposes runtime auth and management endpoints.
- `system/status` remains available in both modes as the single status probe.

### 2. Initialized mode could continue after DB or ZeroTier init failure

- Severity: Critical
- Fix benefit: High
- Status: Mitigated in current refactor

Startup previously logged warnings and continued even when the persisted system state said initialization was complete. That created a partially alive process with missing core dependencies.

Current direction:

- If config says the instance is initialized, DB and ZeroTier initialization must succeed or startup fails.
- Uninitialized mode is the only state allowed to continue without full dependencies.

### 3. Multi-database support was implied but not actually complete

- Severity: High
- Fix benefit: High
- Status: Mitigated in current refactor

MySQL and PostgreSQL were still visible in product surfaces, while reset/init branches were incomplete and could silently continue.

Current direction:

- SQLite is the only formally supported Phase 1 database.
- MySQL/PostgreSQL abstractions may remain in code, but setup and reset paths now reject them explicitly for Phase 1.

## Modules Frozen or Degraded in Phase 1

### Settings

- Status: Degraded
- Supported: password change
- Frozen: language switching, account deletion, broader settings management

### Import Network

- Status: Experimental
- Not part of Phase 1 acceptance
- Entry is retained for later cleanup, but should not be treated as a core path

### Planet

- Status: Experimental
- Not part of Phase 1 acceptance
- Requires isolated validation before any production use

## Stability Risks Still Worth Tracking

### 1. Global mutable dependencies still exist

- `config.AppConfig`
- global DB instance
- `services.GlobalZTClient`

These are reduced in impact but not fully eliminated yet. A later stage should keep dependencies explicit across server assembly and request handlers.

### 2. Setup flow still spans multiple endpoints

The current phase tightened the state model, but setup is still composed of several API calls rather than a single orchestration endpoint. This is acceptable for Phase 1, but a future refactor could consolidate setup into a clearer transaction-like flow.

### 3. Runtime reload hook still exists

`/api/system/reload` remains as an internal compatibility mechanism. It should not be part of the user-facing initialization story and should be removable after the remaining dependency refresh paths are simplified.

## Test and Validation Gaps

### Current baseline

- Frontend `bun run lint`
- Frontend `bun run build`
- Middleware/state boundary tests

### Gaps

- Core handler tests for setup/auth happy-path and failure-path behavior
- Reproducible backend test environment when Go module fetches are restricted
- End-to-end Docker ZeroTier integration coverage remains manual

## Phase 1 Acceptance Baseline

The project should be considered Phase 1 ready only when the following path is reliable:

1. Fresh start in uninitialized mode
2. Configure ZeroTier controller
3. Configure SQLite database
4. Create initial admin
5. Mark initialized
6. Restart and log in successfully
7. Manage networks and members against a Docker ZeroTier controller

## Phase 2 Candidates

- Revisit MySQL/PostgreSQL only after setup and runtime contracts are fully stable
- Rebuild settings into a clearly scoped account/system settings surface
- Decide whether import-network should be promoted or removed
- Decide whether planet tooling belongs in-repo or as a separate admin utility
