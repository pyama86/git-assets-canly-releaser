test:
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
		--save-assets-path ./tmp
