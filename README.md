# Tairitsu

**NOTE: This project is still in development. Some features may not be fully implemented or subject to change.**

Tairitsu is a web-based controller interface for ZeroTier, providing a user-friendly GUI to manage ZeroTier networks, members, and configurations. It consists of a Golang backend that interfaces with the ZeroTier client API and a React-based web frontend.

## Features

- **Network Management**: Create, edit, and delete ZeroTier networks
- **Member Administration**: Manage network members, authorize devices, and assign IPs
- **Configuration Control**: Configure network settings including IP ranges, routes, and rules
- **Real-time Status**: Monitor network and member status
- **Multi-database Support**: Works with SQLite, MySQL, and PostgreSQL
- **Secure Authentication**: JWT-based authentication for secure access
- **Responsive Design**: Modern, responsive UI built with Material UI

## Getting Started

### Prerequisites

- Go 1.22 or later
- Node.js 18 or later with npm
- ZeroTier controller installed and running

### Docker / Podman
Not ready yet.

### Manual Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/GT-610/tairitsu.git
   cd tairitsu
   ```

2. **Backend Setup**
   ```bash
   # Install Go dependencies
   go mod download
   
   # Build the backend
   go build -o tairitsu ./cmd/tairitsu
   ```

3. **Frontend Setup**
   ```bash
   cd web
   
   # Install npm dependencies
   npm install
   
   # Build the frontend
   npm run build
   
   cd ..
   ```

### Configuration

Create a `.env` file in the project root with the following configuration:

```
# Server Configuration
SERVER_PORT=8080
```

### Running the Application

```bash
# Run the compiled binary
./tairitsu
```

The backend will start on port 8080 by default. You will see the notes when accessing through the browser.

For the frontend, just download the static files and host it on a web server on the same host.

### Development Mode

#### Backend
```bash
go run ./cmd/tairitsu
```

#### Frontend
```bash
cd web
npm run dev
```

The frontend development server will start on port 5173 by default, and will proxy API requests to the backend server.

## Usage
Not ready yet.

## Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.

## License

This project is licensed under GNU GPL v3 License - see the [LICENSE](LICENSE) file for details.

## Sponsor
Currently we have not prepared for accepting sponsorships. However, if you find Tairitsu useful, please consider starring the project, or recommend it to others.

---

Tairitsu is an independent project and is **NOT officially affiliated** with ZeroTier, Inc.