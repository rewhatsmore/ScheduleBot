-- Переименовать колонку user_id в telegram_user_id в таблице users
ALTER TABLE "users" RENAME COLUMN "user_id" TO "telegram_user_id";

-- Удалить существующий внешний ключ на user_id в таблице appointments
ALTER TABLE "appointments" DROP CONSTRAINT appointments_user_id_fkey;

-- Удалить существующий внешний ключ на user_id в таблице appointments
ALTER TABLE users DROP CONSTRAINT users_pkey;

-- Добавить новую колонку internal_user_id с автоинкрементом в таблицу users
ALTER TABLE "users" ADD COLUMN "internal_user_id" SERIAL PRIMARY KEY;

-- Переименовать колонку user_id в internal_user_id в таблице appointments
ALTER TABLE "appointments" RENAME COLUMN "user_id" TO "internal_user_id";

-- Обновить значения в колонке internal_user_id в таблице appointments
-- Это предполагает, что у вас есть соответствие между telegram_user_id и internal_user_id
UPDATE "appointments" SET "internal_user_id" = (
    SELECT "internal_user_id" FROM "users" WHERE "telegram_user_id" = "appointments"."internal_user_id"
);

-- Добавить новый внешний ключ на internal_user_id в таблице appointments
ALTER TABLE "appointments" ADD FOREIGN KEY ("internal_user_id") REFERENCES "users" ("internal_user_id") ON DELETE CASCADE;

ALTER TABLE users
ADD CONSTRAINT unique_telegram_user_id UNIQUE (telegram_user_id);

Alter TABLE users drop CONSTRAINT unique_telegram_user_id

