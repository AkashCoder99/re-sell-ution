.PHONY: run migrate seed setup verify-search-plans

DBTOOL := go run ./cmd/dbtool

run:
	go run ./cmd/server

setup: migrate seed

migrate:
	$(DBTOOL) migrate

seed:
	$(DBTOOL) seed

verify-search-plans:
	$(DBTOOL) verify-search-plans
