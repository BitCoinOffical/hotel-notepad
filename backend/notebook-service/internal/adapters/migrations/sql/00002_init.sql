-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_first_name ON guests(first_name);
CREATE INDEX idx_mid_name ON guests(mid_name);
CREATE INDEX idx_last_name ON guests(last_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_first_name;
DROP INDEX IF EXISTS idx_mid_name;
DROP INDEX IF EXISTS idx_last_name;
-- +goose StatementEnd