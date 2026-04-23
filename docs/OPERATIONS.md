# Operations and Support Boundaries

This document describes what Tairitsu includes, what stays experimental, and which boundaries are important for operators.

## Available Capabilities

Tairitsu includes:

- initialization and first admin creation
- user registration and login
- user governance by the platform admin
- owner-scoped network management
- member approval, rejection, editing, and removal
- IPv4, IPv6, routes, DNS, and multicast settings
- import takeover of controller networks not yet owned in Tairitsu

## Explicit Boundaries

The following are **not** supported today:

- MySQL or PostgreSQL support
- multi-tenant organizations or broader platform tenancy features
- email-based password recovery
- notification systems
- a polished public manual-installation guide

## Experimental Surface

`Planet` generation remains available for controlled testing, but it is **experimental**:

- it should be treated as a separate test surface
- it should be validated independently before production use

## Internal vs Public Documentation

- Public docs live under `docs/`
- Working notes, audits, and developer-oriented regression checklists stay under `.cache/`

That split is intentional: public docs should explain what operators need to run Tairitsu, while internal notes can stay more implementation-specific.
