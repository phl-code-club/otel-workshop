CREATE TABLE "Users" (
"id" SERIAL PRIMARY KEY,
"email" VARCHAR(255) UNIQUE NOT NULL,
"password_hash" VARCHAR(255) NOT NULL,
"role" INTEGER REFERENCES "roles"("id") DEFAULT 1,
"created_at" TIMESTAMP DEFAULT NOW()
);

CREATE TABLE "Roles" (
  "id" SERIAL PRIMARY KEY,
  "name" VARCHAR(50) UNIQUE NOT NULL
);

INSERT INTO "roles" ("name") VALUES ('user'), ('admin');
