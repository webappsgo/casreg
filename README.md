# casreg - Self-Hosted Docker Registry Platform

A complete, production-ready Docker Registry platform with integrated authentication, vulnerability scanning, signature verification, and comprehensive management tools.

## Features

- **Full Docker Registry V2 API** - Complete OCI-compliant container registry
- **User Management** - Multi-user support with role-based access control
- **Organization Support** - Team collaboration with organizational registries
- **Vulnerability Scanning** - Integrated Trivy scanner for security analysis
- **Signature Verification** - Cosign integration for image signing
- **Web Interface** - Modern Svelte-based UI with Dracula theme
- **CLI Tool** - Bubbletea-powered interactive terminal interface

## Quick Start

```bash
# Download and run
./casreg serve
```

Default credentials: `admin` / `changeme`
Access: http://localhost:8080

## Build from Source

```bash
make build
./build/casreg serve
```

See CLAUDE.md for complete specifications.
