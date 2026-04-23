# Tairitsu API Documentation

This document describes the current public-facing API shape used by the bundled frontend. It is intentionally concise and focuses on the endpoints that are part of the normal runtime and setup flows.

## Overview

- Base URL: `/api`
- All responses are JSON
- Most runtime endpoints require `Authorization: Bearer <token>`
- Setup endpoints are only available before initialization
- Runtime/admin access is enforced server-side

## Setup and System

### `GET /system/status`

Returns initialization and runtime availability information.

Example:

```json
{
  "initialized": true,
  "hasDatabase": true,
  "hasAdmin": true,
  "allowPublicRegistration": true,
  "ztStatus": {
    "version": "1.14.2",
    "address": "8789af2692",
    "online": true,
    "tcpFallbackAvailable": true,
    "apiReady": true
  }
}
```

### `POST /system/database`

Setup-only. Configures the database.

Request:

```json
{
  "type": "sqlite",
  "path": "data/tairitsu.db"
}
```

### `GET /system/zerotier/test`

Setup-only. Tests controller connectivity.

### `POST /system/zerotier/config`

Setup-only. Saves controller URL and token path.

Request:

```json
{
  "controllerUrl": "http://127.0.0.1:9993",
  "tokenPath": "/var/lib/zerotier-one/authtoken.secret"
}
```

### `POST /system/admin/init`

Setup-only. Prepares the admin creation step.

### `POST /system/initialized`

Setup-only. Marks the system initialized.

Request:

```json
{
  "initialized": true
}
```

### `GET /system/settings`

Runtime, admin-only. Returns instance runtime settings.

### `PUT /system/settings`

Runtime, admin-only. Updates instance runtime settings.

Request and response shape:

```json
{
  "allow_public_registration": true
}
```

Response:

```json
{
  "message": "实例设置已更新",
  "settings": {
    "allow_public_registration": true
  }
}
```

## Authentication and Sessions

### `POST /auth/register`

Registers a user account. During setup, the first account becomes `admin`. During runtime, public registration follows `allow_public_registration`.

Request:

```json
{
  "username": "alice",
  "password": "secret123"
}
```

Response:

```json
{
  "user": {
    "id": "uuid",
    "username": "alice",
    "role": "user",
    "createdAt": "2026-04-23T10:00:00Z"
  },
  "message": "注册成功"
}
```

### `POST /auth/login`

Authenticates a user and creates a tracked session.

Request:

```json
{
  "username": "alice",
  "password": "secret123",
  "remember_me": true
}
```

Response:

```json
{
  "token": "jwt-token",
  "user": {
    "id": "uuid",
    "username": "alice",
    "role": "user",
    "createdAt": "2026-04-23T10:00:00Z"
  },
  "session": {
    "id": "session-id",
    "userAgent": "Mozilla/5.0",
    "ipAddress": "127.0.0.1",
    "rememberMe": true,
    "lastSeenAt": "2026-04-23T10:05:00Z",
    "expiresAt": "2026-04-24T10:05:00Z",
    "createdAt": "2026-04-23T10:00:00Z",
    "updatedAt": "2026-04-23T10:05:00Z",
    "current": true
  }
}
```

### `POST /auth/logout`

Revokes the current session.

### `GET /profile`

Returns the authenticated user's profile.

### `PUT /profile/password`

Changes the current user's password. When `logout_other_sessions` is true, password change and session revocation are performed atomically.

Request:

```json
{
  "current_password": "old-secret",
  "new_password": "new-secret",
  "confirm_password": "new-secret",
  "logout_other_sessions": true
}
```

Response:

```json
{
  "message": "密码修改成功",
  "revoked_other_sessions": 2
}
```

### `GET /profile/sessions`

Returns the current user's session list.

### `DELETE /profile/sessions/:sessionId`

Revokes one session owned by the current user.

### `DELETE /profile/sessions/others`

Revokes all other sessions of the current user.

## Networks and Members

Network and member access is owner-scoped:

- users only see networks where they are the recorded `owner_id`
- accessing someone else's network returns `403`
- accessing a missing network returns `404`

### `GET /networks`

Returns lightweight owned network summaries.

Example item:

```json
{
  "id": "8056c2e21c000001",
  "name": "alpha",
  "description": "alpha-desc",
  "owner_id": "user-uuid",
  "member_count": 3,
  "authorized_member_count": 2,
  "pending_member_count": 1,
  "created_at": "2026-04-23T10:00:00Z",
  "updated_at": "2026-04-23T10:10:00Z"
}
```

### `POST /networks`

Creates a network owned by the current user.

### `GET /networks/:id`

Returns full network detail, database description, and current members.

### `PUT /networks/:id`

Updates network configuration.

### `PUT /networks/:id/metadata`

Updates network name and description.

### `DELETE /networks/:id`

Deletes an owned network.

### `GET /networks/:id/members`

Returns members for an owned network.

### `GET /networks/:id/members/:memberId`

Returns one member in an owned network.

### `PUT /networks/:id/members/:memberId`

Updates one member. Common fields include:

```json
{
  "authorized": true,
  "name": "node-1",
  "activeBridge": false,
  "noAutoAssignIps": false,
  "ipAssignments": ["10.10.10.5"]
}
```

### `DELETE /networks/:id/members/:memberId`

Removes a member from an owned network.

## User Governance

The system keeps a single-admin model. These endpoints are admin-only.

### `GET /users`

Returns all users.

### `POST /users`

Creates a normal user and returns a one-time temporary password.

Response:

```json
{
  "message": "用户创建成功，请通过其他方式安全告知用户临时密码",
  "user": {
    "id": "uuid",
    "username": "alice",
    "role": "user",
    "createdAt": "2026-04-23T10:00:00Z"
  },
  "temporary_password": "TempSecret123"
}
```

### `POST /users/transfer-admin`

Transfers the admin role to another user.

Request:

```json
{
  "user_id": "target-user-uuid"
}
```

### `POST /users/:userId/reset-password`

Resets a user's password, returns a one-time temporary password, and revokes existing sessions.

### `DELETE /users/:userId`

Deletes a normal user, transfers owned networks to the current admin, and revokes sessions.

Response:

```json
{
  "message": "用户已删除，名下网络已转让给当前管理员",
  "user": {
    "id": "uuid",
    "username": "alice",
    "role": "user",
    "createdAt": "2026-04-23T10:00:00Z"
  },
  "transferred_networks": 2,
  "revoked_sessions": 1
}
```

## Import Network

These endpoints are admin-only.

### `GET /admin/networks/importable`

Returns controller takeover candidates and summary counts.

Response:

```json
{
  "candidates": [
    {
      "network_id": "8056c2e21c000001",
      "name": "alpha",
      "description": "alpha-desc",
      "controller_status": "OK",
      "member_count": 3,
      "status": "available",
      "can_import": true,
      "reason_code": "unregistered",
      "reason_message": "网络尚未登记到 Tairitsu，可直接接管"
    }
  ],
  "summary": {
    "total": 1,
    "available": 1,
    "managed": 0,
    "blocked": 0
  }
}
```

### `POST /admin/networks/import`

Imports controller networks for a target owner.

Request:

```json
{
  "network_ids": ["8056c2e21c000001"],
  "owner_id": "user-uuid"
}
```

Response:

```json
{
  "target_owner": {
    "id": "user-uuid",
    "username": "alice"
  },
  "summary": {
    "requested": 1,
    "imported": 1,
    "failed": 0,
    "skipped": 0
  },
  "imported": [
    {
      "network_id": "8056c2e21c000001",
      "name": "alpha",
      "owner_id": "user-uuid",
      "owner_username": "alice"
    }
  ],
  "failed": [],
  "skipped": []
}
```

## Planet

`Planet` endpoints are admin-only and experimental:

- `GET /admin/planet/identity`
- `POST /admin/planet/keys`
- `POST /admin/planet/generate`

They are intentionally outside the normal mainline validation gate.
