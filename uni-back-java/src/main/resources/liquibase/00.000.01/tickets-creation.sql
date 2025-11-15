--changeset dev:nick-08

CREATE TABLE IF NOT EXISTS tickets
(
    id          TEXT PRIMARY KEY,
    user_id     BIGINT       NOT NULL,
    department  VARCHAR(255) NOT NULL,
    subject     VARCHAR(255) NOT NULL,
    message     TEXT         NOT NULL,
    response    TEXT,
    response_by VARCHAR(255),
    user_reply  TEXT,
    created_at  TIMESTAMP    NOT NULL,
    updated_at  TIMESTAMP    NOT NULL,
    status      VARCHAR(50)  NOT NULL,
    CONSTRAINT fk_tickets_user FOREIGN KEY (user_id) REFERENCES users(id)
);