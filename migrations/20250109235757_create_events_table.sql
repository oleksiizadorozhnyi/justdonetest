-- +goose Up
-- +goose StatementBegin
CREATE TABLE order_events (
                              id SERIAL PRIMARY KEY,
                              event_id UUID NOT NULL UNIQUE,
                              order_id UUID NOT NULL,
                              user_id UUID NOT NULL,
                              status TEXT NOT NULL,
                              created_at TIMESTAMP NOT NULL,
                              updated_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE order_events;
-- +goose StatementEnd
