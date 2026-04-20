build:
	@mkdir -p app/static
	@touch app/static/.gitkeep
	@go build -o bin/GoRadio ./cmd/app/main.go
	@go build -o bin/analyzer ./cmd/analyzer/main.go
	@go build -o bin/migrate ./cmd/migrate/main.go

migrate:
	@bin/migrate

analyze:
	@bin/analyzer

run:
	@bin/GoRadio

dev:
	@cd frontend && npm run dev

build-web:
	@mkdir -p app/static
	@cd frontend && npm run build
	@cp -rf frontend/build/* app/static

buildr:
	@make build
	@make run

