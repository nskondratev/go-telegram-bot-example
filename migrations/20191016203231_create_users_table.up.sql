CREATE TABLE IF NOT EXISTS "users" (
    telegram_user_id BIGINT NOT NULL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    language VARCHAR(10),
    source_lang VARCHAR(10),
    target_lang VARCHAR(10),
    points BIGINT NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz
)
