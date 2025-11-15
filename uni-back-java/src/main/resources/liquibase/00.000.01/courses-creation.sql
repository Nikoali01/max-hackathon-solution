--liquibase formatted sql

--changeset dev:nick-02
CREATE SEQUENCE IF NOT EXISTS course_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE CACHE 1;

CREATE TABLE IF NOT EXISTS courses
(
    id           BIGINT PRIMARY KEY DEFAULT nextval('course_id_seq'),
    course_name  VARCHAR(255) NOT NULL,
    course_year  INTEGER      NOT NULL
);

ALTER SEQUENCE course_id_seq OWNED BY courses.id;