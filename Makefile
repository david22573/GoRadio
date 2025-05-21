build:
	@go build -o bin/GoRadio ./cmd/app/main.go

run:
	@bin/GoRadio

buildr:
	@make build
	@make run
