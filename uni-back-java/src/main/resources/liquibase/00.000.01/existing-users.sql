--liquibase formatted sql

--changeset dev:nick-02

CREATE TABLE IF NOT EXISTS allowed_users
(
    id    BIGINT PRIMARY KEY,
    email text NOT NULL,
    role  text NOT NULL
);
