gen-swag:
	go install github.com/swaggo/swag/cmd/swag@latest
	swag init
.PHONY: gen-swag

run:
	go run .
.PHONY: run

test:
	go test -v ./... -race
.PHONY: test

update-snapshot:
	UPDATE_SNAPSHOTS=true go test ./...
.PHONY: update-snapshot

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out