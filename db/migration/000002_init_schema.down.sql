-- Удалить внешний ключ на internal_user_id в таблице appointments
ALTER TABLE "appointments" DROP CONSTRAINT appointments_internal_user_id_fkey;

-- Восстановить колонку user_id в таблице appointments
ALTER TABLE "appointments" RENAME COLUMN "internal_user_id" TO "user_id";

-- Восстановить значения в колонке user_id в таблице appointments
-- Это предполагает, что у вас есть соответствие между internal_user_id и telegram_user_id
UPDATE "appointments" SET "user_id" = (
    SELECT "telegram_user_id" FROM "users" WHERE "internal_user_id" = "appointments"."user_id"
);

-- Удалить колонку internal_user_id из таблицы users
ALTER TABLE "users" DROP COLUMN "internal_user_id";

-- Восстановить внешний ключ на user_id в таблице appointments
ALTER TABLE "appointments" ADD CONSTRAINT appointments_user_id_fkey FOREIGN KEY ("user_id") REFERENCES "users" ("user_id");

-- Восстановить первичный ключ на user_id в таблице users
ALTER TABLE "users" ADD CONSTRAINT users_pkey PRIMARY KEY ("user_id");

-- Восстановить колонку user_id в таблице users
ALTER TABLE "users" RENAME COLUMN "telegram_user_id" TO "user_id";