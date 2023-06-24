dev:
	go run --race -v .

build:
	go build -v -ldflags "-r -v" -o main .