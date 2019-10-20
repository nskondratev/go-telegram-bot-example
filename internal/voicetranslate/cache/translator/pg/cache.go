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
	selectQuery = `UPDATE "text_translation"
SET
	"last_requested_at" = CURRENT_TIMESTAMP,
	"requested_count" = "requested_count" + 1
WHERE "hash" = $1 AND "target_lang" = $2
RETURNING "translated_text"
`
	insertQuery = `INSERT INTO "text_translation"
("hash", "target_lang", "input_text", "translated_text")
VALUES
($1, $2, $3, $4)`
)

type Cache struct {
	db *pgx.ConnPool
}

func New(db *pgx.ConnPool) *Cache {
	return &Cache{db: db}
}

func (c *Cache) Get(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	textHash, err := util.Hash(text)
	if err != nil {
		return "", err
	}
	row := c.db.QueryRowEx(ctx, selectQuery, &pgx.QueryExOptions{}, textHash, targetLang)
	var translated string
	err = row.Scan(&translated)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", voicetranslate.ErrNotFoundInCache
		}
		return "", fmt.Errorf("failed to select from translation cache: %w", err)
	}
	return translated, nil
}

func (c *Cache) Store(ctx context.Context, text, translation, sourceLang, targetLang string) error {
	textHash, err := util.Hash(text)
	if err != nil {
		return err
	}
	_, err = c.db.ExecEx(ctx, insertQuery, &pgx.QueryExOptions{}, textHash, targetLang, text, translation)
	if err != nil {
		return fmt.Errorf("failed to store text translation cache: %w", err)
	}
	return nil
}
