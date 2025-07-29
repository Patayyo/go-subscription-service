# Подгрузка переменных из .env

include .env
export

MIGRATE_CMD=docker-compose run --rm --entrypoint "" migrate
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Применить все миграции
migrate-up:
	$(MIGRATE_CMD) /bin/sh -c "migrate -path=/migrations -database='$(DB_URL)' up"

# Откатить все миграции
migrate-down:
	$(MIGRATE_CMD) /bin/sh -c "migrate -path=/migrations -database='$(DB_URL)' down"

migrate-version:
	$(MIGRATE_CMD) /bin/sh -c "migrate -path=/migrations -database='$(DB_URL)' version"