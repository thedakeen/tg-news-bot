package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"news-bot/internal/models"
	"time"
)

type SourcePostgresStorage struct {
	DB *sqlx.DB
}

func NewSourcesStorage(db *sqlx.DB) *SourcePostgresStorage {
	return &SourcePostgresStorage{DB: db}
}

func (s *SourcePostgresStorage) Sources(ctx context.Context) ([]models.Source, error) {
	conn, err := s.DB.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var sources []dbSource

	err = conn.SelectContext(ctx, &sources, `SELECT * FROM sources`)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecordFound
		}
		return nil, err
	}

	return lo.Map(sources, func(source dbSource, _ int) models.Source { return models.Source(source) }), nil

}

func (s *SourcePostgresStorage) SourceByID(ctx context.Context, id int64) (*models.Source, error) {
	conn, err := s.DB.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var source dbSource

	err = conn.GetContext(ctx, &source, `SELECT * FROM sources WHERE id = $1`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecordFound
		}
		return nil, err
	}

	return (*models.Source)(&source), nil
}

func (s *SourcePostgresStorage) Add(ctx context.Context, source models.Source) (int64, error) {
	conn, err := s.DB.Connx(ctx)
	if err != nil {
		return 0, err
	}

	defer conn.Close()

	var id int64

	row := conn.QueryRowxContext(
		ctx,
		`INSERT INTO sources (name, feed_url, created_at) VALUES ($1,$2,$3) RETURNING id`,
		source.Name,
		source.FeedURL,
		source.CreatedAt)

	if err := row.Err(); err != nil {
		return 0, err
	}

	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil

}

func (s *SourcePostgresStorage) Delete(ctx context.Context, id int64) error {
	conn, err := s.DB.Connx(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	result, err := s.DB.ExecContext(ctx, `DELETE FROM sources WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffecte, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffecte == 0 {
		return models.ErrNoRecordFound
	}

	return nil
}

type dbSource struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	FeedURL   string    `db:"feed_url"`
	CreatedAt time.Time `db:"created_at"`
}
