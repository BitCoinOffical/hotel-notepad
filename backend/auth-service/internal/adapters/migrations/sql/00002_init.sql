-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_username ON users(username);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_username;
-- +goose StatementEnd