-- +goose Up
CREATE TABLE users
(
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    login         VARCHAR(256) NOT NULL UNIQUE,
    password_hash VARCHAR(256) NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);


-- +goose Down
DROP TABLE users;
