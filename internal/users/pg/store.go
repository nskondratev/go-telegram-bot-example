package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/nskondratev/go-telegram-translator-bot/internal/users"
)

const (
	selectUserByTelegramIDQuery = `
SELECT
	telegram_user_id,
	username,
	first_name,
	last_name,
	language,
	source_lang,
	target_lang,
	points
FROM "users"
WHERE telegram_user_id = $1
`
	insertUserQuery = `
INSERT INTO "users"
	(telegram_user_id, username, first_name, last_name, language, source_lang, target_lang, points)
VALUES
	($1, $2, $3, $4, $5, $6, $7, $8)
`
	updateTranslationLangsQuery = `UPDATE "users"
SET "source_lang" = $1, "target_lang" = $2
WHERE "telegram_user_id" = $3`
	updateSourceLangQuery = `UPDATE "users"
SET "source_lang" = $1
WHERE "telegram_user_id" = $2`
	updateTargetLangQuery = `UPDATE "users"
SET "target_lang" = $1
WHERE "telegram_user_id" = $2`
	updateChargeCostQuery = `UPDATE "users"
SET "points" = "points" - $1
WHERE "telegram_user_id" = $2`
)

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) GetUserByTelegramUserID(ctx context.Context, tgUserID int64) (users.User, error) {
	var user users.User
	row := s.db.QueryRow(ctx, selectUserByTelegramIDQuery, tgUserID)
	err := row.Scan(&user.TelegramUserID, &user.UserName, &user.FirstName, &user.LastName, &user.Lang, &user.SourceLang,
		&user.TargetLang, &user.Points)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, users.ErrUserNotFound
		}
		return user, fmt.Errorf("failed to scan user by telegram user id: %w", err)
	}
	return user, nil
}

func (s *Store) StoreUser(ctx context.Context, user *users.User) error {
	_, err := s.db.Exec(ctx, insertUserQuery, user.TelegramUserID, user.UserName, user.FirstName, user.LastName,
		user.Lang, user.SourceLang, user.TargetLang, user.Points)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (s *Store) UpdateTranslationLangs(ctx context.Context, tgUserID int64, sourceLang, targetLang string) error {
	_, err := s.db.Exec(ctx, updateTranslationLangsQuery, sourceLang, targetLang, tgUserID)
	if err != nil {
		return fmt.Errorf("failed to update user translation languages: %w", err)
	}
	return nil
}

func (s *Store) UpdateSourceLang(ctx context.Context, tgUserID int64, sourceLang string) error {
	_, err := s.db.Exec(ctx, updateSourceLangQuery, sourceLang, tgUserID)
	if err != nil {
		return fmt.Errorf("failed to update user source lang: %w", err)
	}
	return nil
}

func (s *Store) UpdateTargetLang(ctx context.Context, tgUserID int64, targetLang string) error {
	_, err := s.db.Exec(ctx, updateTargetLangQuery, targetLang, tgUserID)
	if err != nil {
		return fmt.Errorf("failed to update user target lang: %w", err)
	}
	return nil
}

func (s *Store) ChargeCost(ctx context.Context, tgUserID, cost int64) error {
	_, err := s.db.Exec(ctx, updateChargeCostQuery, cost, tgUserID)
	if err != nil {
		return fmt.Errorf("failed to charge cost for user: %w", err)
	}
	return nil
}
