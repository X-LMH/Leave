.PHONY: migrate

migrate:
	cd backend && go run ./cmd/migrate -config ./config/config.yaml
