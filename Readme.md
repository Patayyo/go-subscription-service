# Subscriptions Service

Сервис для управления онлайн-подписками пользователей.

## Стек:
* Golang
* PostgreSQL
* Gorilla Mux
* Swagger
* Docker + Docker compose
* Migrate CLI (через `migrate/migrate`)
* Makefile
* .env конфигурация

---

## Запуск проекта

### 1. Клонировать репозиторий

```bash
git clone https://github.com/Patayyo/go-subscriptions-service.git
```

### 2. Создать `.env` файл на основе шаблона `.env.example`

зайдайте переменные окружения:

```env
DB_HOST=db
DB_PORT=5432
DB_USER=вставьте ваше имя пользователя postgre
DB_PASSWORD=вставьте ваш пароль postgre
DB_NAME=subscription(вставьте ваше название бд)
```

### 3. запустите сервис

```bash
docker-compose up --build
```

База данных поднимется на `localhost:5432`, бэкенд — на `localhost:8080`.

---

## Swagger-документация

Документация доступна после запуска по адресу:

[http://localhost:8080/swagger/index.html]

В swagger доступны все ручки CRUD и ручка получения суммы подписок за период

### При необходимости swagger-документацию можно пересобрать командой:

```bash
swag init --generalInfo cmd/main.go --output docs
```

---

## Миграции

Для наката миграции:

```bash
make migrate-up - выполнять через терминал
```

Для отката миграции:

```bash
make migrate-down - выполнять через терминал
```

Для получения текущей версии миграции:

```bash
make migrate-version - выполнять через терминал
```

---

## Структура проекта

```

├── cmd/                   # Точка входа (main.go)
├── db/                    # Подключение к БД, загрузка .env
├── docs/                  # Сгенерированный Swagger (swagger.yaml / .json / docs.go)
├── internal/              # Основная бизнес-логика
│   ├── handler/           # HTTP Handlers
│   ├── service/           # Бизнес-логика
│   ├── repo/              # Работа с БД
│   ├── model/             # Модели БД
│   └── dto/               # DTO структуры
├── pgk/
│   ├──utils/              # Парсинг дат 
│   └── validator/         # Валидация данных
├── migrations/            # SQL миграции
├── build/Dockerfile       # Dockerfile приложения
├── docker-compose.yml     # Compose-файл
├── Makefile               # Команды для миграций
├── .env.example           # Шаблон переменных окружения
└── README.md              # Документация
```

---

## Примеры запросов

Все запросы можно протестировать в Swagger-интерфейсе, ниже даны curl-примеры:

### Создание подписки:

```bash
curl -X POST http://localhost:8080/subscription \
 -H "Content-Type: application/json" \
 -d '{
   "service_name": "Netflix",
   "price": 1000,
   "user_id": "d24e286e-fae2-4945-9c90-f124a84d4831",
   "start_date": "01-2024",
   "end_date": "01-2025"
}'
```

### Получение суммы подписок:

```bash
curl "http://localhost:8080/subscription/total_amount?user_id=d24e286e-fae2-4945-9c90-f124a84d4831&from=2024-01-01&to=2024-12-31"
```


