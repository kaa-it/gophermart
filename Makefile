build:
	go build -o gophermart ./cmd/gophermart ;

run_server:
	./gophermart -d "postgres://ak:postgres@localhost:5432/gophermart" -a ":8081" -r "http://localhost:8080"