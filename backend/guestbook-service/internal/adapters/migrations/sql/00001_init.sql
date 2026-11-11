-- +goose Up

CREATE TABLE IF NOT EXISTS guests (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY, 
    user_id UUID NOT NULL,
    first_name TEXT(255) NOT NULL,
    mid_name TEXT(255) NOT NULL,
    last_name TEXT(255) NOT NULL,
    note VARCHAR(2048),
    phone INTEGER(11) NOT NULL,
    is_delete BOOLEAN DEFAULT FALSE,
    current_date TIMESTAMPTZ NOT NULL,
    departure_date TIMESTAMPTZ NOT NULL,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS guests;