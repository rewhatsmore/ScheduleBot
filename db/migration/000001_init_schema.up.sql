CREATE TABLE "users" (
    "user_id" bigint PRIMARY KEY,
    "full_name" varchar NOT NULL,
    "is_admin" boolean NOT NULL DEFAULT FALSE,
    "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "trainings" (
    "training_id" bigserial PRIMARY KEY,
    "place" varchar NOT NULL,
    "type" varchar NOT NULL DEFAULT 'рукоходы/силовая',
    "date_and_time" timestamptz NOT NULL,
    "price" bigint NOT NULL DEFAULT 600,
    "trainer" varchar NOT NULL DEFAULT 'Роман Заколодкин'
);

CREATE TABLE "appointments" (
    "appointment_id" bigserial PRIMARY KEY,
    "training_id" bigint NOT NULL,
    "user_id" bigint NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX ON "trainings" ("date_and_time");

CREATE INDEX ON "appointments" ("training_id");

CREATE UNIQUE INDEX ON "appointments" ("user_id", "training_id");

COMMENT ON COLUMN "users"."user_id" IS 'chat_id from telegram';

ALTER TABLE "appointments" ADD FOREIGN KEY ("training_id") REFERENCES "trainings" ("training_id") ON DELETE CASCADE;

ALTER TABLE "appointments" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") ON DELETE CASCADE;
