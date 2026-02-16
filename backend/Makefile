.PHONY: run migrate

run:
	go run ./cmd/server

migrate:
	psql "$${DATABASE_URL}" -f migrations/0001_init.sql
	psql "$${DATABASE_URL}" -f migrations/0002_password_reset_tokens.sql
	psql "$${DATABASE_URL}" -f migrations/0003_add_user_bio.sql
