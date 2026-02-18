.PHONY: run migrate seed setup

run:
	go run ./cmd/server

setup: migrate seed

migrate:
	psql "$${DATABASE_URL}" -f migrations/0001_init.sql
	psql "$${DATABASE_URL}" -f migrations/0002_password_reset_tokens.sql
	psql "$${DATABASE_URL}" -f migrations/0003_password_reset_otps.sql
	psql "$${DATABASE_URL}" -f migrations/0004_add_user_bio.sql
	psql "$${DATABASE_URL}" -f migrations/0005_b5_constraints_indexes.sql
	psql "$${DATABASE_URL}" -f migrations/0007_soft_delete_audit.sql
	psql "$${DATABASE_URL}" -f migrations/0006_notifications.sql

seed:
	psql "$${DATABASE_URL}" -f seeds/001_categories.sql
