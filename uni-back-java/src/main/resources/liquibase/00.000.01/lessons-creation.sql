--liquibase formatted sql

--changeset dev:lessons-01
CREATE SEQUENCE IF NOT EXISTS lessons_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE CACHE 1;

CREATE TABLE IF NOT EXISTS lessons
(
    id           BIGINT PRIMARY KEY DEFAULT nextval('lessons_id_seq'),
    lesson_name  VARCHAR(255) NOT NULL,
    professor_id BIGINT       NOT NULL,
    place        TEXT         NOT NULL,
    description  TEXT         NOT NULL,
    course_id    BIGINT       NOT NULL,
    date_time    TIMESTAMP    NOT NULL,
    CONSTRAINT fk_lessons_professor FOREIGN KEY (professor_id) REFERENCES users (id),
    CONSTRAINT fk_lessons_course FOREIGN KEY (course_id) REFERENCES courses (id)
);

ALTER SEQUENCE lessons_id_seq OWNED BY lessons.id;
