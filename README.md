# WAP Bot

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## Overview

WAP Bot helps music-sharing communities manage their discussions.
Either by extracting Spotify, YouTube, and YouTube Music links from Slack threads or creating new threads, handling votes etc.

> Because of some slack limitations you can submit commands for this bot via mentions!

## Features

- When mentioned with "summarize", it generates a CSV file containing song titles, artists, URLs, and platform types.
  (currently supported platforms: Spotify, YouTube and YouTube Music)

## Development Workflow

### Prerequisites

- **Go 1.25.3+** (or use mise to auto-install)
- **mise** - [Install mise](https://mise.jdx.dev/getting-started.html)
- **Docker** with Docker Compose
- **Slack App** with tokens:
  - Bot User OAuth Token (`xoxb-*`)
  - App-Level Token (`xapp-*`)
  - See `deploy/slack-app.yaml` for required scopes

### Local Development

1. **Setup environment:**

   ```bash
   mise init
   # Edit .env with your Slack tokens
   ```

2. **Start development stack:**

   ```bash
   mise start
   ```

3. **View logs and metrics:**
   - Check container logs: `docker compose logs -f bot`
   - Open Grafana: http://localhost:3000
   - View traces and metrics in real-time

4. **Common Commands:**
   - `mise task` to list available tasks
   - `mise lint` to lint the codebase
   - `mise test` to run tests

### Project Structure Guidelines

- **`internal/`** - Private code, not importable by other projects
  - `config/` - Environment and configuration management
  - `domain/` - Core business logic, independent of infrastructure
  - `services/` - External integrations (Slack API)
  - `telemetry/` - Cross-cutting observability concerns
- **`pkg/`** - Public libraries that could be extracted/reused
  - `musicextractors/` - Music link extraction (Spotify, YouTube, YouTube Music)
- **`cmd/`** - Application entrypoints, thin layer that wires everything together
