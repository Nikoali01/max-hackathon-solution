--changeset dev:nick-11
CREATE TABLE IF NOT EXISTS allowed_users
(
    id    BIGINT PRIMARY KEY,
    email VARCHAR(255),
    role  VARCHAR(255)
);
