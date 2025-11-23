# SnowOps Number Service

Сервис для работы с гос. номерами: нормализация, проверка по whitelist/blacklist, управление списками.

## Возможности

- Нормализация гос. номеров
- Проверка номеров по whitelist/blacklist
- Добавление/удаление номеров из списков
- Управление списками разрешённых/запрещённых номеров

## Технологии

- Go 1.22+
- GORM + PostgreSQL
- Gin (HTTP router)
- Zerolog (логирование)
- Viper (конфигурация)

## Запуск

### Локально

1. Убедитесь, что PostgreSQL запущен
2. Скопируйте `.env.example` в `app.env` и настройте параметры
3. Запустите сервис:

```bash
go run cmd/number-service/main.go
```

### Docker Compose

```bash
docker compose up --build
```

Сервис будет доступен на `http://localhost:8083`

## API Endpoints

### Health Checks

- `GET /health/live` - проверка работоспособности
- `GET /health/ready` - проверка готовности (включая БД)

### Number Operations

- `POST /api/v1/numbers/check` - проверка номера и получение информации о списках

Пример запроса:

```json
{
  "plate": "123 ABC 02"
}
```

Ответ:

```json
{
  "data": {
    "plate_id": 45,
    "plate": "123ABC02",
    "original": "123 ABC 02",
    "hits": [
      {
        "list_id": 1,
        "list_name": "default_blacklist",
        "list_type": "BLACKLIST"
      }
    ]
  }
}
```

- `POST /api/v1/numbers/whitelist` - добавить номер в whitelist
- `POST /api/v1/numbers/blacklist` - добавить номер в blacklist
- `DELETE /api/v1/numbers/whitelist?plate=123ABC02` - удалить номер из whitelist
- `DELETE /api/v1/numbers/blacklist?plate=123ABC02` - удалить номер из blacklist

## База данных

Сервис создаёт следующие таблицы:

- `plates` - номера (исходный и нормализованный)
- `lists` - списки (whitelist/blacklist)
- `list_items` - элементы списков

Миграции выполняются автоматически при старте сервиса.

