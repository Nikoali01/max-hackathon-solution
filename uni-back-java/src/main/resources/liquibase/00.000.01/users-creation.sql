--liquibase formatted sql

--changeset dev:nick-01
CREATE SEQUENCE IF NOT EXISTS users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE CACHE 1;

CREATE TABLE IF NOT EXISTS users
(
    id             BIGINT PRIMARY KEY           DEFAULT nextval('users_id_seq'),
    name           VARCHAR(100)        NOT NULL,
    surname        VARCHAR(100)        NOT NULL,
    patronymic     VARCHAR(100),
    max_id         BIGINT,
    email          VARCHAR(255) UNIQUE NOT NULL,
    age            INTEGER,
    role           TEXT                NOT NULL,
    generated_code TEXT,
    verified       BOOLEAN             NOT NULL DEFAULT FALSE
);

ALTER SEQUENCE users_id_seq OWNED BY users.id;