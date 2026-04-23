# Validation Checklist

This checklist is the mainline validation baseline for Tairitsu. It is intended to be repeatable before a release or after higher-risk changes.

## Automated Baseline

The following checks should remain green:

- `go test ./...`
- `bun test`
- `bun run lint`
- `bun run build`
- `docker build .`

## Mainline Manual Validation

### 1. Setup and Runtime Entry

1. Start from an uninitialized instance.
2. Complete the setup wizard with:
   - ZeroTier controller URL
   - token path
   - SQLite database
   - initial admin account
3. Confirm the app switches into runtime mode after setup.
4. Confirm `/setup` is no longer the normal runtime entrypoint.

### 2. Account and Access Flow

1. Register a normal user when public registration is enabled.
2. Log in as that user.
3. Confirm the user can create a network and only sees owned networks.
4. Log in as admin and confirm admin-only pages remain restricted to admin accounts.

### 3. Network and Member Flow

1. Create a network and confirm it remains private.
2. Open the detail page and confirm members load.
3. Let a member send a join request and confirm:
   - pending member appears
   - authorize succeeds
   - reject succeeds
4. Confirm an authorized member can be edited and removed.

### 4. Network Settings

1. Save IPv4 subnet and assignment pool changes.
2. Save IPv6 settings.
3. Save managed routes without overwriting primary subnets.
4. Save DNS when empty and when populated.
5. Save multicast settings.

### 5. Admin Governance

1. Toggle public registration and confirm login/register entry behavior matches.
2. Create a normal user as admin.
3. Reset a user's password and confirm old sessions are revoked.
4. Transfer admin to another user.
5. Delete a normal user and confirm owned networks move to the current admin.

### 6. Import Network

1. Confirm controller-only networks appear as importable.
2. Confirm already managed networks appear as managed.
3. Confirm abnormal controller-read cases appear as blocked.
4. Import a network and confirm it appears in the target owner's network list.

## Outside the Mainline Gate

`Planet` remains experimental and is not part of the mainline validation gate.
