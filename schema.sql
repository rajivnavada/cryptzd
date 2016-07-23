CREATE TABLE IF NOT EXISTS "users" (
    "id" integer not null primary key autoincrement,
    "name" varchar(255),
    "email" varchar(255) unique,
    "comment" varchar(255),
    "created_at" datetime,
    "updated_at" datetime
);

CREATE TABLE IF NOT EXISTS "public_keys" (
    "id" integer not null primary key autoincrement,
    "user_id" integer,
    "fingerprint" varchar(255) unique,
    "key_data" blob,
    "created_at" datetime,
    "updated_at" datetime,
    "activated_at" datetime,
    "expires_at" datetime,
    FOREIGN KEY("user_id") REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS "encrypted_messages" (
    "id" integer not null primary key autoincrement,
    "sender_id" integer,
    "public_key_id" integer,
    "subject" varchar(255),
    "cipher" blob,
    "created_at" datetime,
    "updated_at" datetime
);

CREATE TABLE IF NOT EXISTS "projects" (
    "id" integer not null primary key autoincrement,
    "name" varchar(255),
    "environment" varchar(255),
    "default_access_level" varchar(255),
    "created_at" datetime,
    "updated_at" datetime
);

CREATE TABLE IF NOT EXISTS "project_members" (
    "id" integer not null primary key autoincrement,
    "project_id" integer,
    "user_id" integer,
    "access_level" varchar(255),
    "created_at" datetime,
    "updated_at" datetime
);

CREATE TABLE IF NOT EXISTS "project_credential_keys" (
    "id" integer not null primary key autoincrement,
    "project_id" integer,
    "key" varchar(255),
    "created_at" datetime,
    "updated_at" datetime
);

CREATE TABLE IF NOT EXISTS "project_credential_values" (
    "id" integer not null primary key autoincrement,
    "credential_id" integer,
    "public_key_id" integer,
    "cipher" blob,
    "created_at" datetime,
    "updated_at" datetime,
    "expires_at" datetime
);
