package storage

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"news-bot/internal/models"
	"time"
)

type ArticlePostgresStorage struct {
	DB *sqlx.DB
}

func NewArticlesStorage(db *sqlx.DB) *ArticlePostgresStorage {
	return &ArticlePostgresStorage{DB: db}
}

func (s *ArticlePostgresStorage) Store(ctx context.Context, article models.Article) error {
	conn, err := s.DB.Connx(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = conn.ExecContext(
		ctx,
		`INSERT INTO articles(source_id, title, link, summary, published_at) 
               VALUES ($1,$2,$3,$4,$5)
               ON CONFLICT DO NOTHING`,
		article.SourceID,
		article.Title,
		article.Link,
		article.Summary,
		article.PublishedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *ArticlePostgresStorage) AllNotPosted(ctx context.Context, since time.Time, limit uint64) ([]models.Article, error) {
	conn, err := s.DB.Connx(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var articles []dbArticle
	err = conn.SelectContext(
		ctx,
		&articles,
		`SELECT * FROM articles WHERE posted_at IS NULL AND published_at >= $1::timestamp ORDER BY published_at DESC LIMIT $2`,
		since.UTC().Format(time.RFC3339),
		limit)

	if err != nil {
		return nil, err
	}

	return lo.Map(articles, func(article dbArticle, _ int) models.Article {
		return models.Article{
			ID:          article.ID,
			SourceID:    article.SourceID,
			Title:       article.Title,
			Link:        article.Link,
			Summary:     article.Summary.String,
			PublishedAt: article.PublishedAt,
			CreatedAt:   article.CreatedAt,
		}
	}), nil
}

func (s *SourcePostgresStorage) MarkPosted(ctx context.Context, id int64) error {
	conn, err := s.DB.Connx(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = conn.ExecContext(
		ctx,
		`UPDATE articles SET posted_at = $1::timestamp WHERE id = $2`,
		time.Now().UTC().Format(time.RFC3339),
		id,
	)

	if err != nil {
		return err
	}

	return nil
}

type dbArticle struct {
	ID          int64          `db:"a_id"`
	SourceID    int64          `db:"s_id"`
	Title       string         `db:"a_title"`
	Link        string         `db:"a_link"`
	Summary     sql.NullString `db:"a_summary"`
	PublishedAt time.Time      `db:"a_published_at"`
	PostedAt    sql.NullTime   `db:"a_posted_at"`
	CreatedAt   time.Time      `db:"a_created_at"`
}
