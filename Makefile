build:
	@cd ./frontend  && pnpm build
	@cp -rf ./frontend/dist/* ./app/static/
	@go build -o bin/GoRadio ./cmd/app/main.go

run:
	@bin/GoRadio

buildr:
	@make build
	@make run


web-dev:
	@cd ./frontend  && pnpm dev
web-build:
	@cd ./frontend  && pnpm build
