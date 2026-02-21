# Development Guide

This guide covers local development setup, project structure, build instructions, testing, and conventions for contributing to QSGW.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Local Setup](#local-setup)
- [Rust Development](#rust-development)
- [Go Development](#go-development)
- [Python Development](#python-development)
- [Admin UI Development](#admin-ui-development)
- [Database Migrations](#database-migrations)
- [Code Style and Conventions](#code-style-and-conventions)
- [Makefile Targets](#makefile-targets)
- [IDE Setup](#ide-setup)

---

## Prerequisites

Ensure the following tools are installed on your development machine:

| Tool         | Minimum Version | Purpose                              |
|------------- |-----------------|--------------------------------------|
| Rust         | 1.75+           | Gateway engine, crypto, TLS crates   |
| Go           | 1.23+           | Control plane REST API               |
| Python       | 3.11+           | AI engine (anomaly/bot detection)    |
| Node.js      | 22+             | Admin dashboard (React 19)           |
| Docker       | 24+             | Local services (PostgreSQL, etcd)    |
| Docker Compose | 2.20+         | Orchestrating local services         |
| PostgreSQL client | 16+        | Database migrations and debugging    |

**Verify your environment:**

```bash
rustc --version       # 1.75.0 or later
go version            # go1.23 or later
python3 --version     # 3.11 or later
node --version        # v22 or later
docker --version      # 24.0 or later
docker compose version # 2.20 or later
```

---

## Local Setup

### Step 1: Clone the Repository

```bash
git clone https://github.com/yazhsab/qbitel-qsgw.git
cd qbitel-qsgw
```

### Step 2: Start Infrastructure Services

Start PostgreSQL and etcd using Docker Compose:

```bash
docker compose up -d postgres etcd
```

Wait for the services to be healthy:

```bash
docker compose ps
```

### Step 3: Configure Environment Variables

Copy the example environment file and customize it:

```bash
cp .env.example .env
```

Ensure the following variables are set:

```bash
QSGW_DATABASE_URL=postgres://qsgw:password@localhost:5432/qsgw?sslmode=disable
QSGW_JWT_SECRET=dev-jwt-secret-change-in-production
QSGW_API_KEY=qsgw_k_dev_key
QSGW_ETCD_ENDPOINTS=http://localhost:2379
QSGW_LOG_LEVEL=debug
```

### Step 4: Run Database Migrations

```bash
psql -h localhost -U qsgw -d qsgw -f migrations/001_initial_schema.sql
```

### Step 5: Build and Start All Services

Using the Makefile:

```bash
make dev
```

Or start each service individually (see sections below).

### Step 6: Verify the Setup

```bash
# Control plane health
curl http://localhost:8085/health

# AI engine health
curl http://localhost:8086/health

# Admin dashboard
open http://localhost:3003
```

---

## Rust Development

The Rust code is organized as a Cargo workspace with four crates:

```
gateway/     Main gateway engine (Axum, Tokio, reverse proxy)
crypto/      Post-quantum cryptography (ML-KEM, ML-DSA, SLH-DSA)
tls/         TLS integration (rustls configuration, PQC cipher suites)
types/       Shared types and data structures
```

### Gateway Engine (`gateway/`)

The gateway engine is an async HTTP reverse proxy built on Axum and Tokio. It handles:

- TLS termination with configurable PQC policies
- Request routing based on path prefix and priority
- Middleware pipeline (auth, rate limiting, PQC enforcement, security headers)
- Connection pooling to upstream services
- Real-time telemetry to the AI engine

**Key files:**

| Path                           | Description                        |
|--------------------------------|------------------------------------|
| `gateway/src/main.rs`          | Application entry point            |
| `gateway/src/server.rs`        | Axum server and router setup       |
| `gateway/src/proxy.rs`         | Reverse proxy handler              |
| `gateway/src/middleware/`      | Middleware implementations         |
| `gateway/src/config.rs`        | Configuration loading              |

### Crypto Crate (`crypto/`)

Implements NIST post-quantum cryptographic primitives:

- **ML-KEM** (FIPS 203): Key encapsulation mechanism for key exchange
- **ML-DSA** (FIPS 204): Digital signature algorithm for authentication
- **SLH-DSA** (FIPS 205): Stateless hash-based signatures

### TLS Crate (`tls/`)

Configures rustls with post-quantum cipher suites and hybrid key exchange. Implements the four TLS policies (`PQC_ONLY`, `PQC_PREFERRED`, `HYBRID`, `CLASSICAL_ALLOWED`).

### Build Commands

```bash
# Build all crates
cargo build --workspace

# Build in release mode
cargo build --workspace --release

# Run the gateway
cargo run -p gateway

# Run tests for all crates
cargo test --workspace

# Run tests for a specific crate
cargo test -p crypto
cargo test -p gateway

# Run benchmarks
cargo bench --workspace

# Run clippy (linter)
cargo clippy --workspace -- -D warnings

# Format code
cargo fmt --all

# Check formatting without modifying files
cargo fmt --all -- --check
```

### Adding a New Middleware

1. Create a new file in `gateway/src/middleware/`.
2. Implement the middleware as an Axum layer or extractor.
3. Register the middleware in `gateway/src/server.rs` in the appropriate position in the middleware stack.
4. Add tests in the same file or in `gateway/tests/`.

---

## Go Development

The Go control plane is a REST API built with Chi v5 and pgx v5 for PostgreSQL access.

### Project Structure

```
control-plane/
  cmd/
    server/
      main.go              Entry point
  internal/
    config/                Configuration loading
    handlers/              HTTP handlers (gateways, upstreams, routes, threats)
    services/              Business logic layer
    repositories/          Database access layer (pgx)
    middleware/             Shared middleware (auth, rate limiting, CORS, logging)
    models/                Data models and types
  migrations/              SQL migration files
  go.mod
  go.sum
```

### Adding a New Handler

1. **Define the model** in `internal/models/`.
2. **Create the repository** in `internal/repositories/` with pgx queries.
3. **Implement the service** in `internal/services/` with business logic.
4. **Write the handler** in `internal/handlers/` with request validation and response formatting.
5. **Register routes** in the handler's `Routes()` method and mount in `cmd/server/main.go`.

**Example handler pattern:**

```go
// internal/handlers/example.go
type ExampleHandler struct {
    service services.ExampleService
}

func (h *ExampleHandler) Routes() chi.Router {
    r := chi.NewRouter()
    r.Get("/", h.List)
    r.Post("/", h.Create)
    r.Get("/{id}", h.GetByID)
    return r
}
```

### Build and Run

```bash
cd control-plane

# Build
go build -o bin/control-plane ./cmd/server

# Run
go run ./cmd/server

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Vet (static analysis)
go vet ./...

# Format code
gofmt -w .
```

### Shared Middleware

The Go middleware package provides reusable middleware components:

| Middleware    | Description                                       |
|-------------- |---------------------------------------------------|
| `auth`        | JWT validation and API key verification           |
| `ratelimit`   | Per-IP sliding window rate limiter                |
| `cors`        | CORS headers for admin dashboard access           |
| `database`    | Connection pool injection into request context    |
| `logging`     | Structured request logging with zap               |

---

## Python Development

The AI engine is a FastAPI application with scikit-learn and numpy for threat detection models.

### Project Structure

```
ai-engine/
  main.py                  FastAPI application entry point
  detectors/
    anomaly_detector.py    Anomaly detection model
    bot_detector.py        Bot detection model
  models/                  Serialized ML models
  training/                Model training scripts
  tests/
    test_anomaly.py
    test_bot.py
  requirements.txt
  pyproject.toml
```

### Setting Up the Python Environment

```bash
cd ai-engine

# Create and activate a virtual environment
python3 -m venv .venv
source .venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Install development dependencies
pip install -r requirements-dev.txt
```

### Running the AI Engine

```bash
# Development mode with auto-reload
uvicorn main:app --host 0.0.0.0 --port 8086 --reload

# Production mode with multiple workers
uvicorn main:app --host 0.0.0.0 --port 8086 --workers 4
```

### Adding a New Detector

1. Create a new file in `ai-engine/detectors/` (e.g., `custom_detector.py`).
2. Implement the detector class with an `analyze()` method that accepts a feature dictionary and returns a detection result.
3. Register the detector in `main.py` by adding it to the analysis pipeline.
4. Write tests in `ai-engine/tests/`.
5. Update the API endpoint if new input features are required.

### Testing and Linting

```bash
# Run tests
pytest

# Run tests with coverage
pytest --cov=. --cov-report=term-missing

# Run a specific test file
pytest tests/test_anomaly.py

# Lint with ruff
ruff check .

# Auto-fix lint issues
ruff check . --fix

# Format with ruff
ruff format .

# Type checking
mypy .
```

---

## Admin UI Development

The admin dashboard is a React 19 application built with Vite 6 and TypeScript 5.7.

### Project Structure

```
admin/
  src/
    components/            Reusable UI components
    pages/                 Page-level components
      GatewaysPage.tsx
      UpstreamsPage.tsx
      RoutesPage.tsx
      ThreatsPage.tsx
      DashboardPage.tsx
    hooks/                 Custom React hooks
    services/              API client functions
    types/                 TypeScript type definitions
    App.tsx                Root component with routing
    main.tsx               Entry point
  public/
  index.html
  vite.config.ts
  tsconfig.json
  package.json
```

### Running the Development Server

```bash
cd admin

# Install dependencies
npm install

# Start the development server with hot reload
npm run dev
```

The admin dashboard runs on `http://localhost:3003` and proxies API requests to the control plane at `http://localhost:8085`.

### Building for Production

```bash
# Type checking
npm run type-check

# Lint
npm run lint

# Build
npm run build

# Preview the production build
npm run preview
```

### Adding a New Page

1. Create a new page component in `admin/src/pages/`.
2. Add API client functions in `admin/src/services/`.
3. Define TypeScript types in `admin/src/types/`.
4. Add the route in `admin/src/App.tsx`.
5. Add navigation in the sidebar component.

---

## Database Migrations

Database migrations are plain SQL files stored in the `migrations/` directory.

### Creating a New Migration

```bash
# Create a new migration file
touch migrations/002_add_feature.sql
```

**Migration file naming convention:** `NNN_description.sql` where `NNN` is a zero-padded sequence number.

**Migration file structure:**

```sql
-- migrations/002_add_feature.sql
-- Description: Add feature X to the Y table

BEGIN;

ALTER TABLE example ADD COLUMN new_field TEXT;
CREATE INDEX idx_example_new_field ON example(new_field);

COMMIT;
```

### Applying Migrations

```bash
# Apply a specific migration
psql -h localhost -U qsgw -d qsgw -f migrations/002_add_feature.sql

# Apply all migrations in order
for f in migrations/*.sql; do
  psql -h localhost -U qsgw -d qsgw -f "$f"
done
```

### Database Schema

The current schema includes:

| Table            | Description                              |
|------------------|------------------------------------------|
| `gateways`       | Gateway instance configuration           |
| `upstreams`      | Backend service definitions              |
| `routes`         | Routing rules (gateway -> upstream)      |
| `tls_sessions`   | TLS session metadata for analysis        |
| `threat_events`  | Detected threat events                   |
| `qsgw_audit_log` | Administrative action audit trail        |

---

## Code Style and Conventions

### General

- Use descriptive variable and function names.
- Write doc comments for all public functions and types.
- Keep functions focused and short (under 50 lines where practical).
- Prefer returning errors over panicking or using exceptions.

### Rust

- Follow the [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/).
- Use `clippy` with `-D warnings` -- all warnings must be resolved.
- Use `cargo fmt` for formatting (rustfmt defaults).
- Error handling: use `thiserror` for library crates, `anyhow` for application code.
- Async: use `tokio` runtime; avoid blocking in async contexts.

### Go

- Follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments).
- Use `gofmt` for formatting.
- Error wrapping: use `fmt.Errorf("context: %w", err)`.
- Naming: exported functions use PascalCase, unexported use camelCase.
- Tests: use table-driven tests where appropriate.

### Python

- Follow [PEP 8](https://peps.python.org/pep-0008/) with enforcement via `ruff`.
- Type hints: use type annotations on all function signatures.
- Docstrings: use Google-style docstrings for public functions and classes.
- Maximum line length: 100 characters.

### TypeScript

- Strict mode enabled (`strict: true` in `tsconfig.json`).
- Use functional components with hooks (no class components).
- Prefer `interface` over `type` for object shapes.
- Use `const` by default; use `let` only when reassignment is necessary.

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`

**Scopes:** `gateway`, `control-plane`, `ai-engine`, `admin`, `crypto`, `tls`, `types`, `docs`

**Examples:**

```
feat(gateway): add connection pooling for upstream services
fix(control-plane): handle duplicate gateway names correctly
docs(api): add rate limiting endpoint documentation
test(crypto): add ML-KEM-1024 round-trip tests
```

---

## Makefile Targets

The project includes a Makefile with common development targets:

| Target             | Description                                        |
|--------------------|----------------------------------------------------|
| `make dev`         | Start all services in development mode             |
| `make build`       | Build all components                               |
| `make test`        | Run all tests (Rust, Go, Python, TypeScript)       |
| `make lint`        | Run all linters (clippy, go vet, ruff, eslint)     |
| `make fmt`         | Format all code                                    |
| `make fmt-check`   | Check formatting without modifying files           |
| `make docker-build`| Build Docker images for all components             |
| `make docker-up`   | Start all services via Docker Compose              |
| `make docker-down` | Stop all Docker Compose services                   |
| `make migrate`     | Run database migrations                            |
| `make clean`       | Remove build artifacts                             |
| `make bench`       | Run Rust benchmarks                                |

**Usage:**

```bash
# Run all tests
make test

# Build and start everything
make build && make dev

# Lint before committing
make lint
```

---

## IDE Setup

### Visual Studio Code

Recommended extensions:

| Extension                   | Purpose                          |
|-----------------------------|----------------------------------|
| `rust-analyzer`             | Rust language support            |
| `golang.go`                 | Go language support              |
| `ms-python.python`          | Python language support          |
| `charliermarsh.ruff`        | Python linting (ruff)            |
| `dbaeumer.vscode-eslint`    | TypeScript/JavaScript linting    |
| `esbenp.prettier-vscode`    | Code formatting                  |
| `bradlc.vscode-tailwindcss` | Tailwind CSS IntelliSense        |

Recommended workspace settings (`.vscode/settings.json`):

```json
{
  "editor.formatOnSave": true,
  "rust-analyzer.check.command": "clippy",
  "go.lintTool": "golangci-lint",
  "python.linting.enabled": true,
  "python.analysis.typeCheckingMode": "basic",
  "[rust]": {
    "editor.defaultFormatter": "rust-lang.rust-analyzer"
  },
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  },
  "[python]": {
    "editor.defaultFormatter": "charliermarsh.ruff"
  },
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "[typescriptreact]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  }
}
```

### JetBrains IDEs

- **RustRover** or **IntelliJ with Rust plugin:** For gateway, crypto, and TLS crate development.
- **GoLand:** For control plane development. Configure the Go SDK path and enable `go vet` on save.
- **PyCharm:** For AI engine development. Configure the Python interpreter to use the virtual environment in `ai-engine/.venv`.
- **WebStorm:** For admin dashboard development. Enable ESLint and Prettier integration.

### Neovim

For a multi-language setup, use `nvim-lspconfig` with the following language servers:

- `rust_analyzer` for Rust
- `gopls` for Go
- `pyright` or `ruff-lsp` for Python
- `ts_ls` for TypeScript
