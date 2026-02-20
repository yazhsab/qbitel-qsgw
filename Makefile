.PHONY: all setup build test lint clean migrate docker-deps docker-all docker-down

# ==== Setup ====
setup: setup-go setup-rust setup-node setup-python
	@echo "Setup complete"

setup-go:
	go work sync

setup-rust:
	cargo fetch

setup-node:
	cd admin && npm install

setup-python:
	cd ai-engine && python3 -m venv .venv && .venv/bin/pip install -e ".[dev]"

# ==== Build ====
build: build-go build-rust build-node

build-go:
	cd control-plane && go build ./...

build-rust:
	cargo build --workspace

build-node:
	cd admin && npm run build

# ==== Test ====
test: test-go test-rust test-python

test-go:
	cd control-plane && go test ./...

test-rust:
	cargo test --workspace

test-python:
	cd ai-engine && .venv/bin/pytest

# ==== Lint ====
lint: lint-go lint-rust lint-python

lint-go:
	cd control-plane && golangci-lint run ./...

lint-rust:
	cargo clippy --workspace -- -D warnings
	cargo fmt --all -- --check

lint-python:
	cd ai-engine && .venv/bin/ruff check .

# ==== Database Migrations ====
migrate:
	migrate -path db/migrations -database "$${QSGW_DATABASE_URL}" up

migrate-down:
	migrate -path db/migrations -database "$${QSGW_DATABASE_URL}" down

# ==== Docker ====
docker-deps:
	docker compose -f infra/docker/docker-compose.deps.yml up -d

docker-all:
	docker compose -f infra/docker/docker-compose.yml up -d

docker-down:
	docker compose -f infra/docker/docker-compose.yml down

# ==== Clean ====
clean:
	cargo clean
	cd control-plane && go clean -cache
	rm -rf admin/dist admin/node_modules/.cache
