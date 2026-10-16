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

## Reverse Proxy

Forwarding headers are only believed when they come from a proxy that was
explicitly listed. Only loopback is trusted by default, which covers a proxy in
the same container or on the host's loopback; everything else is ignored, so a
client cannot claim to be someone else.

If TLS terminates in a sibling container or on the Docker gateway, list that
proxy, otherwise the session cookie loses its `Secure` flag and HSTS is not sent
even though users reach Tairitsu over HTTPS:

- `TAIRITSU_TRUSTED_PROXIES` — comma separated addresses or CIDR ranges, for
  example `172.17.0.1,10.0.0.0/8`. Entries that cannot be parsed are rejected
  with a warning.
- `TAIRITSU_PROXY_HEADER` — the header carrying the client address. Defaults to
  `X-Forwarded-For`, which nginx, Caddy, Traefik and HAProxy all set. Set it to
  `X-Real-IP` only for an nginx that sends that instead.
- `TAIRITSU_TRUSTED_PROXY_LOOPBACK` — set to `false` to stop trusting loopback
  proxies entirely.

## Experimental Surface

`Planet` generation remains available for controlled testing, but it is **experimental**:

- it should be treated as a separate test surface
- it should be validated independently before production use
