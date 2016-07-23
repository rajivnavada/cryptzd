CREATE TABLE IF NOT EXISTS "users" (
    "id" integer not null primary key autoincrement,
    "name" varchar(255),
    "email" varchar(255) not null unique,
    "comment" varchar(255),
    "created_at" datetime not null,
    "updated_at" datetime not null
);

CREATE TABLE IF NOT EXISTS "public_keys" (
    "id" integer not null primary key autoincrement,
    "user_id" integer not null,
    "fingerprint" varchar(255) not null unique,
    "key_data" blob,
    "created_at" datetime not null,
    "updated_at" datetime not null,
    "activated_at" datetime,
    "expires_at" datetime not null,
    FOREIGN KEY("user_id") REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "encrypted_messages" (
    "id" integer not null primary key autoincrement,
    "sender_id" integer not null,
    "public_key_id" integer not null,
    "subject" varchar(255),
    "cipher" blob not null,
    "created_at" datetime not null,
    "updated_at" datetime not null,
    FOREIGN KEY("sender_id") REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY("public_key_id") REFERENCES public_keys(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "projects" (
    "id" integer not null primary key autoincrement,
    "name" varchar(255),
    "environment" varchar(255),
    "default_access_level" varchar(255),
    "created_at" datetime not null,
    "updated_at" datetime not null
);

CREATE TABLE IF NOT EXISTS "project_members" (
    "id" integer not null primary key autoincrement,
    "project_id" integer not null,
    "user_id" integer not null,
    "access_level" varchar(255),
    "created_at" datetime not null,
    "updated_at" datetime not null,
    FOREIGN KEY("project_id") REFERENCES projects(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY("user_id") REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "project_credential_keys" (
    "id" integer not null primary key autoincrement,
    "project_id" integer not null,
    "key" varchar(255),
    "created_at" datetime not null,
    "updated_at" datetime not null,
    FOREIGN KEY("project_id") REFERENCES projects(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "project_credential_values" (
    "id" integer not null primary key autoincrement,
    "credential_id" integer not null,
    "public_key_id" integer not null,
    "cipher" blob not null,
    "created_at" datetime not null,
    "updated_at" datetime not null,
    "expires_at" datetime not null,
    FOREIGN KEY("credential_id") REFERENCES project_credential_keys(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY("public_key_id") REFERENCES public_keys(id) ON UPDATE CASCADE ON DELETE CASCADE
);
