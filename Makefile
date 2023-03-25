build:
	go build -o ./bin/app ./cmd/

migrate-up:
	./bin/app --config-path=./example/config/config.toml

migrate-down:
	./bin/app --config-path=./example/config/config.toml --rollback=true
