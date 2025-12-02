# Tairitsu API Documentation

## Table of Contents
1. [Introduction](#introduction)
2. [API Overview](#api-overview)
3. [Authentication](#authentication)
4. [API Endpoints](#api-endpoints)
   - [System APIs](#system-apis)
   - [Authentication APIs](#authentication-apis)
   - [Network APIs](#network-apis)
   - [Member APIs](#member-apis)
   - [User Profile APIs](#user-profile-apis)
   - [User Management APIs](#user-management-apis)
5. [Data Models](#data-models)
6. [Common Responses](#common-responses)
7. [Error Codes](#error-codes)

## Introduction

This document provides detailed information about the Tairitsu API, which is a web-based interface for managing ZeroTier networks. The API allows users to perform various operations such as creating networks, managing network members, and configuring system settings.

## API Overview

- Base URL: `/api`
- All API responses are in JSON format
- Authentication is required for most endpoints except system initialization endpoints
- Error responses follow a consistent format
- API rate limiting is applied to all endpoints to prevent abuse
  - Default limit: 100 requests per IP address, with 10 requests refilled per second
  - When the limit is exceeded, a 429 Too Many Requests response is returned

## Authentication

Most API endpoints require authentication using JWT (JSON Web Tokens). To authenticate:

1. Obtain a token by logging in via the `/api/auth/login` endpoint
2. Include the token in the `Authorization` header of subsequent requests:
   ```
   Authorization: Bearer <your_token>
   ```

## API Endpoints

### System APIs

#### Get System Status
- **URL**: `/api/system/status`
- **Method**: `GET`
- **Description**: Retrieves the current system status including initialization status, database configuration, and ZeroTier connection status
- **Authentication**: Not required
- **Parameters**: None
- **Response**:
  ```json
  {
    "initialized": true,
    "hasDatabase": true,
    "hasAdmin": true,
    "ztStatus": {
      "version": "1.8.5",
      "address": "abcd1234",
      "online": true,
      "tcpFallbackAvailable": true,
      "apiReady": true
    }
  }
  ```

#### Configure Database
- **URL**: `/api/system/database`
- **Method**: `POST`
- **Description**: Configures the database connection settings
- **Authentication**: Not required (during setup only)
- **Request Body**:
  ```json
  {
    "type": "sqlite",
    "path": "/path/to/database.db"
  }
  ```
- **Response**:
  ```json
  {
    "message": "数据库配置成功",
    "config": {
      "type": "sqlite",
      "path": "/path/to/database.db"
    }
  }
  ```

#### Test ZeroTier Connection
- **URL**: `/api/system/zerotier/test`
- **Method**: `GET`
- **Description**: Tests connectivity to the ZeroTier controller
- **Authentication**: Not required (during setup only)
- **Parameters**: None
- **Response**:
  ```json
  {
    "version": "1.8.5",
    "address": "abcd1234",
    "online": true,
    "tcpFallbackAvailable": true,
    "apiReady": true
  }
  ```

#### Save ZeroTier Configuration
- **URL**: `/api/system/zerotier/config`
- **Method**: `POST`
- **Description**: Saves ZeroTier configuration and initializes connection
- **Authentication**: Not required (during setup only)
- **Request Body**:
  ```json
  {
    "controllerUrl": "http://localhost:9993",
    "tokenPath": "/var/lib/zerotier-one/authtoken.secret"
  }
  ```
- **Response**:
  ```json
  {
    "message": "ZeroTier配置保存成功",
    "config": {
      "controllerUrl": "http://localhost:9993",
      "tokenPath": "/var/lib/zerotier-one/authtoken.secret"
    },
    "status": {
      "version": "1.8.5",
      "address": "abcd1234",
      "online": true,
      "tcpFallbackAvailable": true,
      "apiReady": true
    }
  }
  ```

#### Initialize ZeroTier Client
- **URL**: `/api/system/zerotier/init`
- **Method**: `POST`
- **Description**: Initializes the ZeroTier client for the application
- **Authentication**: Not required (during setup only)
- **Parameters**: None
- **Response**:
  ```json
  {
    "message": "ZeroTier客户端初始化成功"
  }
  ```

#### Initialize Admin Creation
- **URL**: `/api/system/admin/init`
- **Method**: `POST`
- **Description**: Prepares the system for admin account creation
- **Authentication**: Not required (during setup only)
- **Parameters**: None
- **Response**:
  ```json
  {
    "message": "管理员账户创建步骤初始化成功",
    "resetDone": true,
    "databaseType": "sqlite"
  }
  ```

#### Set System Initialization Status
- **URL**: `/api/system/initialized`
- **Method**: `POST`
- **Description**: Updates the system initialization status
- **Authentication**: Not required (during setup only)
- **Request Body**:
  ```json
  {
    "initialized": true
  }
  ```
- **Response**:
  ```json
  {
    "message": "初始化状态更新成功"
  }
  ```

#### Reload Routes
- **URL**: `/api/system/reload`
- **Method**: `POST`
- **Description**: Reloads application routes
- **Authentication**: Not required (during setup only)
- **Parameters**: None
- **Response**:
  ```json
  {
    "message": "路由重新加载成功"
  }
  ```

### Authentication APIs

#### User Registration
- **URL**: `/api/auth/register`
- **Method**: `POST`
- **Description**: Registers a new user account
- **Authentication**: Not required
- **Request Body**:
  ```json
  {
    "username": "user123",
    "password": "password123",
    "email": "user@example.com"
  }
  ```
- **Response**:
  ```json
  {
    "user": {
      "id": "uuid-string",
      "username": "user123",
      "email": "user@example.com",
      "role": "user",
      "created_at": "2023-01-01T00:00:00Z"
    },
    "message": "注册成功"
  }
  ```

#### User Login
- **URL**: `/api/auth/login`
- **Method**: `POST`
- **Description**: Authenticates a user and returns a JWT token
- **Authentication**: Not required
- **Request Body**:
  ```json
  {
    "username": "user123",
    "password": "password123"
  }
  ```
- **Response**:
  ```json
  {
    "token": "jwt-token-string",
    "user": {
      "id": "uuid-string",
      "username": "user123",
      "email": "user@example.com",
      "role": "user",
      "created_at": "2023-01-01T00:00:00Z"
    }
  }
  ```

### Network APIs

#### Get ZeroTier Status
- **URL**: `/api/status`
- **Method**: `GET`
- **Description**: Gets the ZeroTier network status
- **Authentication**: Required
- **Parameters**: None
- **Response**:
  ```json
  {
    "version": "1.8.5",
    "address": "abcd1234",
    "online": true,
    "tcpFallbackAvailable": true,
    "apiReady": true
  }
  ```

#### Get All Networks
- **URL**: `/api/networks`
- **Method**: `GET`
- **Description**: Retrieves a list of all networks
- **Authentication**: Required
- **Parameters**: None
- **Response**:
  ```json
  [
    {
      "id": "abcdef1234567890",
      "name": "My Network",
      "description": "A sample network",
      "config": {
        "private": true,
        "allowPassivePortForwarding": false,
        "ipAssignmentPools": [],
        "routes": [],
        "tags": [],
        "rules": [],
        "v4AssignMode": {
          "zt": true
        },
        "v6AssignMode": {
          "zt": false
        }
      },
      "creationTime": 1640995200000,
      "lastModifiedTime": 1640995200000
    }
  ]
  ```

#### Create Network
- **URL**: `/api/networks`
- **Method**: `POST`
- **Description**: Creates a new network
- **Authentication**: Required
- **Request Body**:
  ```json
  {
    "name": "New Network",
    "description": "A new network for testing"
  }
  ```
- **Response**:
  ```json
  {
    "id": "new-network-id",
    "name": "New Network",
    "description": "A new network for testing",
    "config": {
      "private": true,
      "allowPassivePortForwarding": false,
      "ipAssignmentPools": [],
      "routes": [],
      "tags": [],
      "rules": [],
      "v4AssignMode": {
        "zt": true
      },
      "v6AssignMode": {
        "zt": false
      }
    },
    "creationTime": 1640995200000,
    "lastModifiedTime": 1640995200000
  }
  ```

#### Get Network by ID
- **URL**: `/api/networks/{id}`
- **Method**: `GET`
- **Description**: Retrieves details of a specific network
- **Authentication**: Required
- **Parameters**: 
  - `id` (path parameter): Network ID
- **Response**:
  ```json
  {
    "id": "network-id",
    "name": "Network Name",
    "description": "Network Description",
    "config": {
      "private": true,
      "allowPassivePortForwarding": false,
      "ipAssignmentPools": [],
      "routes": [],
      "tags": [],
      "rules": [],
      "v4AssignMode": {
        "zt": true
      },
      "v6AssignMode": {
        "zt": false
      }
    },
    "creationTime": 1640995200000,
    "lastModifiedTime": 1640995200000
  }
  ```

#### Update Network
- **URL**: `/api/networks/{id}`
- **Method**: `PUT`
- **Description**: Updates a network's configuration
- **Authentication**: Required
- **Parameters**: 
  - `id` (path parameter): Network ID
- **Request Body**:
  ```json
  {
    "name": "Updated Network Name",
    "description": "Updated description"
  }
  ```
- **Response**:
  ```json
  {
    "id": "network-id",
    "name": "Updated Network Name",
    "description": "Updated description",
    "config": {
      "private": true,
      "allowPassivePortForwarding": false,
      "ipAssignmentPools": [],
      "routes": [],
      "tags": [],
      "rules": [],
      "v4AssignMode": {
        "zt": true
      },
      "v6AssignMode": {
        "zt": false
      }
    },
    "creationTime": 1640995200000,
    "lastModifiedTime": 1640995300000
  }
  ```

#### Delete Network
- **URL**: `/api/networks/{id}`
- **Method**: `DELETE`
- **Description**: Deletes a network
- **Authentication**: Required
- **Parameters**: 
  - `id` (path parameter): Network ID
- **Response**:
  ```json
  {
    "message": "网络删除成功"
  }
  ```

### Member APIs

#### Get Network Members
- **URL**: `/api/networks/{networkId}/members`
- **Method**: `GET`
- **Description**: Retrieves all members of a network
- **Authentication**: Required
- **Parameters**: 
  - `networkId` (path parameter): Network ID
- **Response**:
  ```json
  [
    {
      "id": "member-id",
      "address": "member-address",
      "config": {
        "authorized": true,
        "activeBridge": false,
        "ipAssignments": ["10.10.10.10"],
        "tags": [],
        "natTraversal": true,
        "capabilities": [],
        "noAutoAssignIps": false
      },
      "identity": "member-identity",
      "name": "Member Name",
      "description": "Member Description",
      "clientVersion": "1.8.5",
      "online": true,
      "lastOnline": 1640995200000,
      "creationTime": 1640995200000
    }
  ]
  ```

#### Get Network Member
- **URL**: `/api/networks/{networkId}/members/{memberId}`
- **Method**: `GET`
- **Description**: Retrieves details of a specific network member
- **Authentication**: Required
- **Parameters**: 
  - `networkId` (path parameter): Network ID
  - `memberId` (path parameter): Member ID
- **Response**:
  ```json
  {
    "id": "member-id",
    "address": "member-address",
    "config": {
      "authorized": true,
      "activeBridge": false,
      "ipAssignments": ["10.10.10.10"],
      "tags": [],
      "natTraversal": true,
      "capabilities": [],
      "noAutoAssignIps": false
    },
    "identity": "member-identity",
    "name": "Member Name",
    "description": "Member Description",
    "clientVersion": "1.8.5",
    "online": true,
    "lastOnline": 1640995200000,
    "creationTime": 1640995200000
  }
  ```

#### Update Network Member
- **URL**: `/api/networks/{networkId}/members/{memberId}`
- **Method**: `PUT`
- **Description**: Updates a network member's configuration
- **Authentication**: Required
- **Parameters**: 
  - `networkId` (path parameter): Network ID
  - `memberId` (path parameter): Member ID
- **Request Body**:
  ```json
  {
    "config": {
      "authorized": true,
      "activeBridge": false,
      "ipAssignments": ["10.10.10.10"]
    },
    "name": "Updated Member Name",
    "description": "Updated Member Description"
  }
  ```
- **Response**:
  ```json
  {
    "id": "member-id",
    "address": "member-address",
    "config": {
      "authorized": true,
      "activeBridge": false,
      "ipAssignments": ["10.10.10.10"],
      "tags": [],
      "natTraversal": true,
      "capabilities": [],
      "noAutoAssignIps": false
    },
    "identity": "member-identity",
    "name": "Updated Member Name",
    "description": "Updated Member Description",
    "clientVersion": "1.8.5",
    "online": true,
    "lastOnline": 1640995200000,
    "creationTime": 1640995200000
  }
  ```

#### Delete Network Member
- **URL**: `/api/networks/{networkId}/members/{memberId}`
- **Method**: `DELETE`
- **Description**: Removes a member from a network
- **Authentication**: Required
- **Parameters**: 
  - `networkId` (path parameter): Network ID
  - `memberId` (path parameter): Member ID
- **Response**:
  ```json
  {
    "message": "成员删除成功"
  }
  ```

### User Profile APIs

#### Get Current User Profile
- **URL**: `/api/profile`
- **Method**: `GET`
- **Description**: Retrieves the profile information of the currently authenticated user
- **Authentication**: Required
- **Parameters**: None
- **Response**:
  ```json
  {
    "id": "user-id",
    "username": "username",
    "email": "user@example.com",
    "role": "user",
    "created_at": "2023-01-01T00:00:00Z"
  }
  ```

#### Change Password
- **URL**: `/api/profile/password`
- **Method**: `PUT`
- **Description**: Updates the password for the currently authenticated user
- **Authentication**: Required
- **Request Body**:
  ```json
  {
    "current_password": "oldpassword123",
    "new_password": "newpassword123",
    "confirm_password": "newpassword123"
  }
  ```
- **Response**:
  ```json
  {
    "message": "密码修改成功"
  }
  ```

### User Management APIs

#### Get All Users
- **URL**: `/api/users`
- **Method**: `GET`
- **Description**: Retrieves a list of all users (admin only)
- **Authentication**: Required (admin role)
- **Parameters**: None
- **Response**:
  ```json
  [
    {
      "id": "user-id-1",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "created_at": "2023-01-01T00:00:00Z"
    },
    {
      "id": "user-id-2",
      "username": "user1",
      "email": "user1@example.com",
      "role": "user",
      "created_at": "2023-01-02T00:00:00Z"
    }
  ]
  ```

#### Update User Role
- **URL**: `/api/users/:userId/role`
- **Method**: `PUT`
- **Description**: Updates the role of a specific user (admin only, cannot modify own role)
- **Authentication**: Required (admin role)
- **Parameters**:
  - `userId` (path parameter): User ID
- **Request Body**:
  ```json
  {
    "role": "admin"
  }
  ```
- **Response**:
  ```json
  {
    "message": "用户角色更新成功",
    "user": {
      "id": "user-id",
      "username": "username",
      "email": "user@example.com",
      "role": "admin",
      "created_at": "2023-01-01T00:00:00Z"
    }
  }
  ```

## Data Models

### User
| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier for the user |
| username | string | User's username |
| password | string | User's password (hashed, not returned in responses) |
| email | string | User's email address |
| role | string | User's role (admin or user) |
| created_at | string | Timestamp when the user was created |
| updated_at | string | Timestamp when the user was last updated |

### Network
| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier for the network |
| name | string | Network name |
| description | string | Network description |
| config | object | Network configuration |
| creationTime | integer | Timestamp when the network was created |
| lastModifiedTime | integer | Timestamp when the network was last modified |

### Member
| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier for the member |
| address | string | Member's ZeroTier address |
| config | object | Member configuration |
| identity | string | Member's identity |
| name | string | Member's name |
| description | string | Member's description |
| clientVersion | string | Version of the ZeroTier client |
| online | boolean | Whether the member is online |
| lastOnline | integer | Timestamp when the member was last online |
| creationTime | integer | Timestamp when the member was created |

### Status
| Field | Type | Description |
|-------|------|-------------|
| version | string | ZeroTier version |
| address | string | Node address |
| online | boolean | Whether the node is online |
| tcpFallbackAvailable | boolean | Whether TCP fallback is available |
| apiReady | boolean | Whether the API is ready |

## Common Responses

### Success Response
```json
{
  "message": "Operation successful"
}
```

### Error Response
```json
{
  "error": "Error description"
}
```

## Error Codes

| HTTP Status Code | Error Code | Description |
|------------------|------------|-------------|
| 200 | OK | Request successful |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Authentication required or failed |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 429 | Too Many Requests | API rate limit exceeded |
| 500 | Internal Server Error | Server-side error |