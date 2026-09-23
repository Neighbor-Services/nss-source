.PHONY: run build test clean docker-build docker-run

run:
	go run ./cmd/server

build:
	go build -ldflags="-w -s" -o bin/server ./cmd/server

test:
	go test -v ./...

clean:
	rm -rf bin/

docker-build:
	docker build -t ns-backend-go:latest .

docker-run:
	docker run -p 8000:8000 --env-file .env ns-backend-go:latest
