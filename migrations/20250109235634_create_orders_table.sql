-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders (
                        id SERIAL PRIMARY KEY,
                        order_id UUID NOT NULL UNIQUE,
                        user_id UUID NOT NULL,
                        status TEXT NOT NULL,
                        created_at TIMESTAMP NOT NULL,
                        updated_at TIMESTAMP NOT NULL,
                        meta JSONB
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders;
-- +goose StatementEnd
