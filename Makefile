.PHONY: run
run:
	go run . run -c config/config.yaml

test:
	go test ./...