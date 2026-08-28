.PHONY: generate sqlc run dev tools

TEMPL_VERSION := v0.3.1020
SQLC_VERSION := v1.31.1

generate:
	templ generate

sqlc:
	sqlc generate

run: generate
	go run ./cmd/server

dev: generate
	air

tools:
	go install github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION)
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)
	go install github.com/air-verse/air@latest
