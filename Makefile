build:
	go build -o gophermart ./cmd/gophermart ;

run_server:
	./gophermart -d "postgres://ak:postgres@localhost:5432/gophermart" -a "http://localhost:5555"