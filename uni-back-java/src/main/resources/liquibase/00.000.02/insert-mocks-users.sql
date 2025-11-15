--liquibase formatted sql

--changeset dev:nick-04
INSERT INTO users (name, surname, patronymic, max_id, email, age, role, generated_code, verified)
VALUES ('Анатолий', 'Белов', 'Иванович', 201, 'anatoly.belov@example.com', 46, 'professor', NULL, TRUE),
       ('Валерий', 'Гордеев', 'Петрович', 202, 'valeriy.gordeev@example.com', 51, 'professor', NULL, TRUE),
       ('Георгий', 'Дьяков', 'Сергеевич', 203, 'georgiy.dyakov@example.com', 39, 'professor', NULL, TRUE),
       ('Денис', 'Ефимов', 'Алексеевич', 204, 'denis.efimov@example.com', 43, 'professor', NULL, TRUE),
       ('Евгений', 'Жарков', 'Дмитриевич', 205, 'evgeniy.zharkov@example.com', 56, 'professor', NULL, TRUE),
       ('Захар', 'Зайцев', 'Михайлович', 206, 'zakhar.zaycev@example.com', 48, 'professor', NULL, TRUE),
       ('Игорь', 'Иванов', 'Николаевич', 207, 'igor.ivanov@example.com', 54, 'professor', NULL, TRUE),
       ('Константин', 'Крылов', 'Владимирович', 208, 'konstantin.krylov@example.com', 50, 'professor', NULL, TRUE),
       ('Леонид', 'Лебедев', 'Евгеньевич', 209, 'leonid.lebedev@example.com', 42, 'professor', NULL, TRUE),
       ('Максим', 'Морозов', 'Анатольевич', 210, 'maksim.morozov@example.com', 45, 'professor', NULL, TRUE);

