build:
	@go build -o bin/GoRadio ./src/

run:
	@bin/GoRadio

buildr:
	@make build
	@make run
