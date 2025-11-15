--liquibase formatted sql

--changeset dev:nick-01
--rollback select 1;

CREATE SEQUENCE IF NOT EXISTS news_feed_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE CACHE 1;

CREATE TABLE IF NOT EXISTS news
(
    id   BIGINT PRIMARY KEY DEFAULT nextval('news_feed_id_seq'),
    text TEXT NOT NULL,
    date timestamp not null
);

ALTER SEQUENCE news_feed_id_seq OWNED BY news.id;
