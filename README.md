# CommentTree — Древовидные комментарии

Сервис для работы с древовидными комментариями на Go с использованием фреймворка [wbf v0.0.13](https://github.com/wb-go/wbf).

## Возможности

- POST /api/comments — создание комментария (с указанием родительского)
- GET /api/comments?parent={id} — получение комментария и всех вложенных
- DELETE /api/comments/{id} — удаление комментария и всех вложенных
- Постраничная навигация (limit, offset)
- Полнотекстовый поиск (POST /api/comments/search)
- Web-интерфейс (HTML + JS)

## Миграции

Миграции находятся в папке `migrations/`.

### Автоматическое применение

**Up миграция** применяется автоматически при первом запуске контейнера PostgreSQL.

### Ручное применение

**Up (создание таблиц):**
```bash
docker exec commenttree-db psql -U postgres -d commenttree -f /docker-entrypoint-initdb.d/001_init.sql
```

**Down (удаление таблиц):**
```bash
docker exec commenttree-db psql -U postgres -d commenttree -f /docker-entrypoint-initdb.d/002_down.sql
```

> ⚠️ Down миграция удалит все данные без возможности восстановления!

---

## Быстрый старт

### 1. Запуск PostgreSQL с миграциями

```bash
docker-compose up -d
```

> ⚠️ Миграции применяются автоматически при первом запуске контейнера.
> Файлы в `migrations/` выполняются в порядке нумерации.

### 2. Установка зависимостей

```bash
go mod tidy
```

### 3. Запуск приложения

```bash
go run ./cmd/app
```

Сервис доступен на http://localhost:8080

## API

### Создание комментария

```bash
curl -X POST http://localhost:8080/api/comments \
  -H "Content-Type: application/json" \
  -d '{"author":"Имя","content":"Текст комментария"}'
```

### Создание ответа

```bash
curl -X POST http://localhost:8080/api/comments \
  -H "Content-Type: application/json" \
  -d '{"parent_id":1,"author":"Имя","content":"Ответ"}'
```

### Получение комментариев

```bash
curl "http://localhost:8080/api/comments?limit=20&offset=0"
```

### Получение вложенных комментариев

```bash
curl "http://localhost:8080/api/comments?parent=1"
```

### Удаление комментария

```bash
curl -X DELETE http://localhost:8080/api/comments/1
```

### Поиск комментариев

```bash
curl -X POST http://localhost:8080/api/comments/search \
  -H "Content-Type: application/json" \
  -d '{"query":"ключевые слова","limit":50,"offset":0}'
```

## Структура проекта

```
.
├── cmd/
│   └── app/
│       └── main.go           # Точка входа, DI
├── internal/
│   ├── config/
│   │   └── config.go         # Загрузка конфигурации (wbf/config)
│   ├── domain/
│   │   ├── comment.go        # Сущность комментария
│   │   ├── repository.go     # Интерфейс репозитория
│   │   └── service.go        # Интерфейс сервиса
│   ├── models/
│   │   └── comment.go        # DTO модели
│   ├── handler/
│   │   └── comment.go        # HTTP API (wbf/ginext)
│   ├── infrastructure/
│   │   └── postgres/
│   │       └── comment_repository.go  # Репозиторий (dbpg)
│   └── usecase/
│       └── comment_usecase.go         # Бизнес-логика
├── web/
│   └── index.html            # Frontend
├── migrations/
│   └── 001_create_comments.sql
├── docker-compose.yml        # PostgreSQL 17
├── config.yaml
├── go.mod
└── README.md
```

## Используемые пакеты wbf

| Пакет | Назначение |
|-------|------------|
| `github.com/wb-go/wbf/config` | Загрузка конфигурации |
| `github.com/wb-go/wbf/logger` | Логирование (zerolog) |
| `github.com/wb-go/wbf/dbpg` | Работа с PostgreSQL |
| `github.com/wb-go/wbf/ginext` | HTTP-роутинг (Gin) |

## Особенности реализации

- **ltree** — расширение PostgreSQL для эффективного хранения иерархии
- **Каскадное удаление** — при удалении комментария удаляются все вложенные
- **Полнотекстовый поиск** — с поддержкой русского языка
- **Clean Architecture** — разделение на domain/usecase/infrastructure/handler
