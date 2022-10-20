export DATABASE_URL = postgres://postgres:postgres@127.0.0.1:5432/todos?sslmode=disable

todos:
	@go build .

.PHONY: run-breach
run-breach: todos
	./envbreach ./todos

.PHONY:
dev/vault:
	vault server -dev 2>&1 | tee vault.log

.PHONY: dev/psql
dev/psql:
	psql "${DATABASE_URL}"

.PHONY: dev/docker-db
dev/docker-db:
	@docker run --rm \
		--name todos-db \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=todos \
		-p 5432:5432 \
		postgres:alpine
