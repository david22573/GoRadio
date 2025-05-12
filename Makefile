build:
	@go build -o bin/GoRadio ./cmd/main.go

run:
	@bin/GoRadio

buildr:
	@make build
	@make run
