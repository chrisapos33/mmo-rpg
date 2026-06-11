.PHONY: dev db db-down backend frontend test harness

# Start only postgres (preferred for local dev)
db:
	docker compose up -d postgres
	@echo "PostgreSQL running on :5432"

db-down:
	docker compose down

# Run backend (loads .env if present)
backend:
	@set -a; [ -f .env ] && . ./.env; set +a; cd backend && go run ./cmd/api

# Run frontend dev server
frontend:
	cd frontend && npm run dev

# Run both in parallel (requires two terminals or use tmux)
dev:
	@echo "Run 'make db' first, then 'make backend' and 'make frontend' in separate terminals."

# Build backend binary
build-backend:
	cd backend && go build -o bin/api ./cmd/api

# Run backend tests
test:
	cd backend && go test ./...

# Dev harness: run scoring/build pipeline against real DB without HTTP or OAuth.
# Usage: make harness               — all users, reads MOCK_AI from .env
#        make harness ARGS="-users 1,2 -mock-ai=false"
harness:
	@set -a; [ -f .env ] && . ./.env; set +a; cd backend && go run ./cmd/harness $(ARGS)

# Copy env example if .env missing
.env:
	cp .env.example .env
	@echo "Created .env from example. Edit it before running."
