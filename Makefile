.PHONY: run build test vet fmt
.PHONY: run build test vet fmt docker-build docker-dev

run:
	go run ./cmd/api

build:
	go build -o bin/ ./cmd/...

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w cmd config db

docker-build:
	docker build -t auth-api .

docker-dev:
	docker run --rm -it \
		-p 8090:8080 \
		--env-file .env \
		-e HTTP_ADDR=0.0.0.0:8080 \
		-v "$(PWD)":/app \
		-v /app/tmp \
		auth-api