--liquibase formatted sql

--changeset dev:nick-02
--rollback DELETE FROM news WHERE id BETWEEN 1 AND 10;

INSERT INTO news (id, text, date)
VALUES (nextval('news_feed_id_seq'), 'В Иннополисе опубликовано новое расписание занятий на весенний семестр.',
        CURRENT_DATE - INTERVAL '10 days'),
       (nextval('news_feed_id_seq'), 'Открыта регистрация на курсы повышения квалификации для сотрудников и студентов.',
        CURRENT_DATE - INTERVAL '9 days'),
       (nextval('news_feed_id_seq'), 'В университете прошёл международный семинар по искусственному интеллекту.',
        CURRENT_DATE - INTERVAL '8 days'),
       (nextval('news_feed_id_seq'), 'Начался приём заявок на летнюю стажировку в ведущих IT-компаниях.',
        CURRENT_DATE - INTERVAL '7 days'),
       (nextval('news_feed_id_seq'), 'Итоги университетского чемпионата по программированию опубликованы.',
        CURRENT_DATE - INTERVAL '6 days'),
       (nextval('news_feed_id_seq'),
        'Профессор Иванов получил государственную награду за исследования в области кибербезопасности.',
        CURRENT_DATE - INTERVAL '5 days'),
       (nextval('news_feed_id_seq'), 'Объявлена стипендия для лучших студентов по направлению Data Science.',
        CURRENT_DATE - INTERVAL '4 days'),
       (nextval('news_feed_id_seq'), 'В Иннополисе стартовал проект поддержки молодых стартапов.',
        CURRENT_DATE - INTERVAL '3 days'),
       (nextval('news_feed_id_seq'), 'Прошёл благотворительный хакатон для поддержки образовательных инициатив.',
        CURRENT_DATE - INTERVAL '2 days'),
       (nextval('news_feed_id_seq'), 'Опубликована стратегия развития университета на ближайшие пять лет.',
        CURRENT_DATE);
