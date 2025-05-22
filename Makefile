build:
	@go build -o bin/GoRadio ./cmd/app/main.go

run:
	@bin/GoRadio

run-web:
	@cd frontend && npm run dev

build-web:
	@cd frontend && npm run build
	@cp -rf frontend/build/* app/static

buildr:
	@make build
	@make run
