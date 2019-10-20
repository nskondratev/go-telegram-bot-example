package pg

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/nskondratev/go-telegram-translator-bot/internal/util"
	"github.com/nskondratev/go-telegram-translator-bot/internal/voicetranslate"
)

const (
	selectQuery = `UPDATE "generated_speech"
SET
	"last_requested_at" = CURRENT_TIMESTAMP,
	"requested_count" = "requested_count" + 1
WHERE "hash" = $1 AND "target_lang" = $2
RETURNING "telegram_file_id"
`
	insertQuery = `INSERT INTO "generated_speech"
("hash", "target_lang", "input_text", "telegram_file_id")
VALUES
($1, $2, $3, $4)`
)

type Cache struct {
	db *pgx.ConnPool
}

func New(db *pgx.ConnPool) *Cache {
	return &Cache{db: db}
}

func (c *Cache) Get(ctx context.Context, text, lang string) (voicetranslate.TextToSpeechResult, error) {
	var res voicetranslate.TextToSpeechResult
	textHash, err := util.Hash(text)
	if err != nil {
		return res, err
	}
	row := c.db.QueryRowEx(ctx, selectQuery, &pgx.QueryExOptions{}, textHash, lang)
	var telegramFileID string
	err = row.Scan(&telegramFileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, voicetranslate.ErrNotFoundInCache
		}
		return res, fmt.Errorf("failed to select from speech cache: %w", err)
	}
	res.FileID = telegramFileID
	return res, nil
}

func (c *Cache) Store(ctx context.Context, fileID, text, lang string) error {
	textHash, err := util.Hash(text)
	if err != nil {
		return err
	}
	_, err = c.db.ExecEx(ctx, insertQuery, &pgx.QueryExOptions{}, textHash, lang, text, fileID)
	if err != nil {
		return fmt.Errorf("failed to store speech translation cache: %w", err)
	}
	return nil
}
