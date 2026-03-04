-- Down миграция для удаления таблицы комментариев
-- Выполняется вручную при необходимости

-- Удаление таблицы и всех зависимостей
DROP TABLE IF EXISTS comments CASCADE;

-- Удаление расширения ltree (если не используется другими таблицами)
-- DROP EXTENSION IF EXISTS ltree CASCADE;

-- Для выполнения миграции:
-- docker exec commenttree-db psql -U postgres -d commenttree -f /docker-entrypoint-initdb.d/002_down.sql
