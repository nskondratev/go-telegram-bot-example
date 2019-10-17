CREATE TABLE IF NOT EXISTS "text_translation" (
    "hash" VARCHAR(64) NOT NULL,
    target_lang VARCHAR(10) NOT NULL,
    input_text TEXT NOT NULL,
    translated_text TEXT NOT NULL,
    requested_count BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_requested_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("hash", target_lang)
)
