.PHONY: run migrate

run:
	go run ./cmd/server

migrate:
	psql "$${DATABASE_URL}" -f migrations/0001_init.sql
