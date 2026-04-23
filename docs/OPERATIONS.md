# Operations and Support Boundaries

This document describes what Tairitsu Phase 1 is meant to cover, what remains experimental, and what operators should verify before calling an instance healthy.

## Supported Product Shape

Tairitsu Phase 1 is intended for:

- SQLite-backed deployments
- single-instance self-hosted controller management
- small self-hosted teams
- day-to-day network and member administration

The current mainline includes:

- initialization and first admin creation
- user registration and login
- user governance by the platform admin
- owner-scoped network management
- member approval, rejection, editing, and removal
- IPv4, IPv6, routes, DNS, and multicast settings
- import takeover of controller networks not yet owned in Tairitsu

## Explicit Boundaries

The following are **not** current Phase 1 promises:

- MySQL or PostgreSQL support
- multi-tenant organizations or broader platform tenancy features
- email-based password recovery
- notification systems
- a polished public manual-installation guide

## Experimental Surface

`Planet` generation remains available for controlled testing, but it is **experimental**:

- it is outside the Phase 1 mainline support claim
- it should not block a release
- it should be validated independently before production use

## Release Baseline

The minimum automated baseline for a release-minded change is:

- `go test ./...`
- `bun test`
- `bun run lint`
- `bun run build`
- `docker build .`

The minimum mainline manual validation is:

1. complete first-run setup
2. register and log in as a normal user
3. create a network and confirm owner-scoped visibility
4. approve or remove a joining member
5. save IPv4, IPv6, routes, DNS, and multicast settings without cross-overwriting
6. import an existing controller network
7. complete admin user governance flows: create user, reset password, transfer admin, delete user

## Internal vs Public Documentation

- Public docs live under `docs/`
- Working notes, audits, and developer-oriented regression checklists stay under `.cache/`

That split is intentional: public docs should explain what operators need to run Tairitsu, while internal notes can stay more implementation-specific.
