-- +goose Up
-- +goose StatementBegin
CREATE TABLE source (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    feed_url VARCHAR(255) NOT NULL,
    created TIMESTAMP NOT NULL DEFAULT NOW(),
    updated TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS source;
-- +goose StatementEnd
