test:
	mkdir -p tmp
	go test ./... -coverprofile=coverage.out -covermode=count

run:
	mkdir -p tmp
	rm -rf tmp/*
	go run main.go --repo STNS/STNS \
		--deploy-command scripts/deploy \
		--rollback-command scripts/rollback \
		--healthcheck-command scripts/healthcheck \
		--package-name-pattern ".*" \
		--log-level debug \
		--health-check-interval 2s \
		--canary-rollout-window 10s \
		--repository-polling-interval 10s \
		--save-assets-path ./tmp \
		--state-file-path ./tmp/state.json

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

.PHONY: test devdeps lint release
