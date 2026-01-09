# Tairitsu

[**简体中文**](readme-i18n/README.zh-CN.md)

**NOTE: This project is still in development. Some features may not be fully implemented or subject to change.**

Tairitsu is a web-based controller interface for ZeroTier, providing a user-friendly GUI to manage ZeroTier networks, members, and configurations. It consists of a Golang backend that interfaces with the ZeroTier client API and a React-based web frontend.

## Features

- **Network Management**: Create, edit, and delete ZeroTier networks
- **Member Administration**: Manage network members, authorize devices, and assign IPs
- **Configuration Control**: Configure network settings including IP ranges, routes, and rules
- **Real-time Status**: Monitor network and member status
- **Multi-database Support**: Works with SQLite, MySQL, and PostgreSQL
- **Secure Authentication**: JWT-based authentication for secure access
- **Responsive Design**: Modern, responsive Material Design interface

## Deployment

### Docker / Podman

1. Create a `local.conf` file in ZeroTier's home directory (Usually `/var/lib/zerotier-one`). If you already have a `local.conf` file, skip this step.

2. Configure `allowManagementFrom` in ZeroTier's `local.conf`:

   ```json
   {
      "settings": {
         "allowManagementFrom": [
               "0.0.0.0/0",
               "::/0"
         ]
      }
   }
   ```

   This will make ZeroTier controller accessible from any IP address.

   Or for more restrictive access, but make sure this IP can be accessed by Tairitsu container:

   ```json
   {
      "settings": {
         "allowManagementFrom": [
               "<local-ip-cidr>",
         ]
      }
   }
   ```

   After modifying the configuration, restart the ZeroTier container.


3. **Run Tairitsu container**

   ```bash
   docker run -d \
       --name tairitsu \
       -p 3000:3000 \
       -v /var/lib/zerotier-one:/var/lib/zerotier-one \
       -v path/to/tairitsu/data:/app/data \
       ghcr.io/gt-610/tairitsu:latest
   ```

   Or through Compose:

   ```yaml
   services:
     tairitsu:
       image: ghcr.io/gt-610/tairitsu:latest
       ports:
         - 3000:3000
       volumes:
         - /var/lib/zerotier-one:/var/lib/zerotier-one
         - path/to/tairitsu/data:/app/data
   ```

### Manual Installation

Not ready yet.

### Development

#### Prerequisites

- Go 1.25 or later with CGO enabled
- Node.js 22 or later with npm
- ZeroTier controller installed and running

#### Backend
```bash
CGO_ENABLED=1 go build -o tairitsu ./cmd/tairitsu
```

#### Frontend
```bash
cd web
npm run dev
```

The frontend development server will start on port 3000 by default, and will proxy API requests to the backend server.

## Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.

## License

[GNU GPL v3](LICENSE).

## Legal Notice

Since version 1.16.0, ZeroTier's controller component is licensed under a [commercial source-available non-free license](https://github.com/zerotier/ZeroTierOne/blob/main/nonfree/LICENSE.md). Tairitsu does not redistribute any ZeroTier controller code and is fully compliant with ZeroTier's licensing terms.

### ZeroTier License Compliance
Tairitsu is a standalone management interface for ZeroTier networks. This project **DOES NOT** include, distribute, or modify any ZeroTier source code or binaries.

This software communicates with a **separately installed** ZeroTier controller via its official API. Users must deploy their own ZeroTier controller under the terms of its license.

The "Generate Planet" feature is modified from [ztnodeid-go](https://github.com/kmahyyg/ztnodeid-go) under GNU GPL v3 License, **not** from ZeroTier itself.

--- 

Tairitsu is **NOT** a ZeroTier product. It is **NOT affiliated with, endorsed by, or supported by ZeroTier, Inc**.