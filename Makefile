test:
	mkdir -p tmp
	go test ./... -coverprofile=coverage.out -covermode=count

ci: lint test

lint: devdeps
	@staticcheck ./...
	go vet ./...

devdeps:
	@which staticcheck > /dev/null || go install honnef.co/go/tools/cmd/staticcheck@latest

release_deps:
	which goreleaser > /dev/null || go install github.com/goreleaser/goreleaser@latest

release: release_deps
	goreleaser --clean

build:
	go build -o dist/gacr main.go

run_example:
	GOOS=linux GOARCH=amd64 make build
	docker compose rm -f
	docker compose up --build

.PHONY: test devdeps lint release
