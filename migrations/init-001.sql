-- -------------------------------------------------------------
-- TablePlus 5.3.4(492)
--
-- https://tableplus.com/
--
-- Database: users
-- Generation Time: 2023-03-11 21:42:22.4370
-- -------------------------------------------------------------


-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS codes_id_seq;

-- Table Definition
CREATE TABLE "public"."codes" (
                                  "id" int4 NOT NULL DEFAULT nextval('codes_id_seq'::regclass),
                                  "code" int4,
                                  "user_id" int4,
                                  "created_at" timestamp,
                                  "expire_at" timestamp,
                                  "action" varchar,
                                  "data" varchar,
                                  "is_used" bool NOT NULL DEFAULT false,
                                  PRIMARY KEY ("id")
);

-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq;

-- Table Definition
CREATE TABLE "public"."users" (
                                  "id" int4 NOT NULL DEFAULT nextval('id_seq'::regclass),
                                  "email" varchar(255),
                                  "password" varchar(60),
                                  "created_at" timestamp,
                                  "updated_at" timestamp,
                                  "role" varchar(255),
                                  "is_active" bool DEFAULT false,
                                  "google_id" varchar(255)
);

ALTER TABLE "public"."codes" ADD FOREIGN KEY ("user_id") REFERENCES "public"."users"("id");
