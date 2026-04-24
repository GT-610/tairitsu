# Operations

This document summarizes what Tairitsu currently provides, what remains experimental, and which boundaries matter in day-to-day operation.

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
