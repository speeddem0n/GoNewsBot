-- +goose Up
-- +goose StatementBegin
CREATE TABLE article
(
    id SERIAL PRIMARY KEY,
    source_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    link VARCHAR(255) NOT NULL UNIQUE,
    summary TEXT NOT NULL,
    published TIMESTAMP NOT NULL,
    created TIMESTAMP NOT NULL DEFAULT NOW(),
    posted TIMESTAMP,
    CONSTRAINT fk_article_source_id
        FOREIGN KEY (source_id)
            REFERENCES source (id)
            ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS article;
-- +goose StatementEnd
