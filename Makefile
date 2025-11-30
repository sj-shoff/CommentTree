.PHONY: run build migrate-up migrate-down docker-up docker-down
include .env
export

run:
	go run cmd/comments-system/main.go

build:
	go build -o bin/comments-system cmd/comments-system/main.go

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

migrate-up:
	goose -dir migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_DB}?sslmode=disable" down