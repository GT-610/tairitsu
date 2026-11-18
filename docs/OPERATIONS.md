# Operations

This document summarizes what Tempest-ZT currently provides, what remains experimental, and which boundaries matter in day-to-day operation.

## Available Capabilities

Tempest-ZT includes:

- initialization and first admin creation
- user registration and login
- user governance by the platform admin
- owner-scoped network management
- member approval, rejection, editing, and removal
- per-network roles (owner, manager, operator, viewer)
- IPv4, IPv6, routes, DNS, and multicast settings
- import takeover of controller networks not yet owned in Tempest-ZT

### Upgrading from read-only shares

On the first start after upgrading from a release that used `network_viewers`,
Tempest-ZT copies every legacy read-only share into `network_grants` as a
`viewer` grant and removes the old table. Users keep their existing access and
do not need to sign in again or re-share anything. The server log records the
legacy row count and the number of inserted grants; no manual migration step is
required.

## Explicit Boundaries

The following are **not** supported today:

- any storage backend other than SQLite
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
even though users reach Tempest-ZT over HTTPS:

- `TEMPEST_ZT_TRUSTED_PROXIES` — comma separated addresses or CIDR ranges, for
  example `172.17.0.1,10.0.0.0/8`. Entries that cannot be parsed are rejected
  with a warning.
- `TEMPEST_ZT_PROXY_HEADER` — the header carrying the client address. Defaults to
  `X-Forwarded-For`, which nginx, Caddy, Traefik and HAProxy all set. Set it to
  `X-Real-IP` only for an nginx that sends that instead.
- `TEMPEST_ZT_TRUSTED_PROXY_LOOPBACK` — set to `false` to stop trusting loopback
  proxies entirely.

The client address also keys the rate limiters and the address recorded on a
session, so getting it wrong has two failure modes that are invisible from
outside:

- **No trusted proxy listed while one is in use.** Every caller is recorded as
  the proxy, so they share one rate limit budget and one noisy client can lock
  the rest out.
- **A proxy that passes the caller's own `X-Forwarded-For` through** instead of
  appending to it. The caller then chooses their own address and the rate limit
  stops applying to them.

The second is a configuration mistake rather than a Tempest-ZT setting, and it is
common enough to be worth stating: the header must *append* the peer address.
With nginx, `$proxy_add_x_forwarded_for` does that and `$http_x_forwarded_for`
does not:

```nginx
location / {
    proxy_pass http://127.0.0.1:3000;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header Host $host;
}
```

A forgotten password is throttled per account, with a delay that doubles after a
handful of failures, so a wrong proxy configuration cannot be used to guess
passwords without limit. Correcting the configuration is still what makes the
per-client limits meaningful.

## Upgrading from Tairitsu

The project is now called Tempest-ZT. A deployment that moves over keeps its data
directory, database, configuration file, and stored ZeroTier token: only the
names below changed.

| What | New | Old |
| --- | --- | --- |
| Image and container | `ghcr.io/gt-610/tempest-zt` | `ghcr.io/gt-610/tairitsu` |
| Binary | `tempest-zt` | `tairitsu` |
| Session cookie | `tempest-zt_session` | `tairitsu_session` |
| Log file | `logs/tempest-zt.log` | `logs/tairitsu.log` |
| Default database file | `data/tempest-zt.db` | `data/tairitsu.db` |

Environment variables are the exception: `TEMPEST_ZT_*` takes precedence, and the
old `TAIRITSU_*` names still work so an existing container definition is not
silently ignored. Each fallback is logged once at startup with the name to
replace, so rename them when convenient.

Both image names are published for now, so a deployment that still pulls
`ghcr.io/gt-610/tairitsu` keeps receiving releases. Switching is one edit to the
`image` field. The old name will stop being published at some point; from then
on, pulling it still succeeds, so nothing breaks visibly — the deployment simply
stays on the last version published under that name. That silence is why the
console tells the administrator about it: on sign-in, an administrator sees a
one-time notice naming both images, dismissible for the current visit or for
good. That choice is kept in the browser, so it is not shared between devices.

Published images are multi-architecture indexes covering `linux/amd64` and
`linux/arm64`, so one tag serves an x86 server and an arm64 board alike.

Everyone signs in again after the upgrade, because the session cookie is scoped
to the new name. A database file that setup already recorded in `config.json`
keeps its path; only the name used when no path is configured changed.

## Experimental Surface

`Planet` generation remains available for controlled testing, but it is **experimental**:

- it should be treated as a separate test surface
- it should be validated independently before production use

Six things about the signing keys, because getting them wrong is silent:

- Each generated planet names a successor key that its **own next update** must be
  signed by. The stored pair therefore advances on every generation: the key that was
  advertised becomes the signer, and a fresh successor is written. There is nothing to
  rotate by hand.
- Because of that, **deploy every generated planet in the order it was generated**. A
  generation consumes the signer for the next one, whether or not its output is
  deployed: generate twice and deploy only the second, and the deployed planet is
  signed with a key the previously deployed world does not accept. Generate and deploy
  one at a time.
- **A generation that returns an error can simply be retried.** The successor is written
  before the signer is promoted, so an attempt that fails leaves the stored pair exactly
  as usable as it was, and the retry produces a planet the deployed world still accepts.
- **A generation whose response you never received is the case to be careful with.** It
  may have succeeded: the signer would then have advanced with no planet to carry the
  advertisement, and a retry would sign an update the deployed planet rejects. Back up
  the signing-key directory before each generation (restricted access — it holds private
  keys), and if a response is lost or doubtful, restore that pre-generation pair before
  generating again. Do not retry it blind.
- A planet generated **without** a signing key directory uses a throwaway pair and
  stores no successor, so no later generation can sign an update for it. That mode is
  for a one-shot test world. To be able to publish updates, set `signing_key_path`
  (advanced mode, custom signing keys) and keep that directory.
- Generating keys refuses when a pair already exists. Replacing a stored pair breaks
  the update chain of any planet already deployed from it, so starting a new identity
  means removing the two key files deliberately.
