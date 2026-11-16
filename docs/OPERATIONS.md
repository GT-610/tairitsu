# Operations

This document summarizes what Tairitsu currently provides, what remains experimental, and which boundaries matter in day-to-day operation.

## Available Capabilities

Tairitsu includes:

- initialization and first admin creation
- user registration and login
- user governance by the platform admin
- owner-scoped network management
- member approval, rejection, editing, and removal
- per-network roles (owner, manager, operator, viewer)
- IPv4, IPv6, routes, DNS, and multicast settings
- import takeover of controller networks not yet owned in Tairitsu

### Upgrading from read-only shares

On the first start after upgrading from a release that used `network_viewers`,
Tairitsu copies every legacy read-only share into `network_grants` as a
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
even though users reach Tairitsu over HTTPS:

- `TAIRITSU_TRUSTED_PROXIES` — comma separated addresses or CIDR ranges, for
  example `172.17.0.1,10.0.0.0/8`. Entries that cannot be parsed are rejected
  with a warning.
- `TAIRITSU_PROXY_HEADER` — the header carrying the client address. Defaults to
  `X-Forwarded-For`, which nginx, Caddy, Traefik and HAProxy all set. Set it to
  `X-Real-IP` only for an nginx that sends that instead.
- `TAIRITSU_TRUSTED_PROXY_LOOPBACK` — set to `false` to stop trusting loopback
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

The second is a configuration mistake rather than a Tairitsu setting, and it is
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

## Experimental Surface

`Planet` generation remains available for controlled testing, but it is **experimental**:

- it should be treated as a separate test surface
- it should be validated independently before production use

Five things about the signing keys, because getting them wrong is silent:

- Each generated planet names a successor key that its **own next update** must be
  signed by. The stored pair therefore advances on every generation: the key that was
  advertised becomes the signer, and a fresh successor is written. There is nothing to
  rotate by hand.
- Because of that, **deploy every generated planet in the order it was generated**. A
  generation consumes the signer for the next one, whether or not its output is
  deployed: generate twice and deploy only the second, and the deployed planet is
  signed with a key the previously deployed world does not accept. Generate and deploy
  one at a time.
- **Do not blindly retry a generation whose response you did not receive.** The keys
  advance before the result is handed back, so an attempt you cannot see the outcome of
  may still have consumed the signer — and the retry then produces a planet the
  deployed world rejects. Back up the signing-key directory (restricted access, it holds
  private keys) before each generation; if a response is lost, restore that pre-generation
  pair before generating again.
- A planet generated **without** a signing key directory uses a throwaway pair and
  stores no successor, so no later generation can sign an update for it. That mode is
  for a one-shot test world. To be able to publish updates, set `signing_key_path`
  (advanced mode, custom signing keys) and keep that directory.
- Generating keys refuses when a pair already exists. Replacing a stored pair breaks
  the update chain of any planet already deployed from it, so starting a new identity
  means removing the two key files deliberately.
