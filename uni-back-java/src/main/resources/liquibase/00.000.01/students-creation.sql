--liquibase formatted sql

--changeset dev:nick-01
CREATE SEQUENCE IF NOT EXISTS students_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE CACHE 1;

CREATE TABLE IF NOT EXISTS students
(
    id        BIGINT PRIMARY KEY DEFAULT nextval('students_id_seq'),
    user_id   BIGINT,
    course_id BIGINT,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES courses (id)
);

ALTER SEQUENCE students_id_seq OWNED BY students.id;