CREATE TABLE IF NOT EXISTS "generated_speech" (
    "hash" VARCHAR(64) NOT NULL,
    target_lang VARCHAR(10) NOT NULL,
    input_text TEXT NOT NULL,
    telegram_file_id BIGINT NOT NULL,
    requested_count BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_requested_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("hash", target_lang)
)
