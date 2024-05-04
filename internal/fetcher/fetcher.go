package fetcher

import (
	"context"
	"log"
	"news-bot/internal/models"
	source2 "news-bot/internal/source"
	"strings"
	"sync"
	"time"
)

type ArticleStorage interface {
	Store(ctx context.Context, article models.Article) error
}

type SourceProvider interface {
	Sources(ctx context.Context) ([]models.Source, error)
}

type Source interface {
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]models.Item, error)
}

type Fetcher struct {
	articles ArticleStorage
	sources  SourceProvider

	fetchInterval  time.Duration
	filterKeywords []string
}

func New(
	articleStorage ArticleStorage,
	sourceProvider SourceProvider,
	fetchInternal time.Duration,
	filterKeywords []string,
) *Fetcher {
	return &Fetcher{
		articles:       articleStorage,
		sources:        sourceProvider,
		fetchInterval:  fetchInternal,
		filterKeywords: filterKeywords,
	}
}

func (f *Fetcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	err := f.Fetch(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.Fetch(ctx); err != nil {
				return err
			}
		}
	}

}

func (f *Fetcher) Fetch(ctx context.Context) error {
	sources, err := f.sources.Sources(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, src := range sources {
		wg.Add(1)

		rssSource := source2.NewRSSSourceFromModel(src)

		go func(source Source) {
			defer wg.Done()

			items, err := source.Fetch(ctx)
			if err != nil {
				log.Printf("[ERROR] Fetching items from source")
				return
			}

			err = f.processItems(ctx, source, items)
			if err != nil {
				log.Printf("[ERROR] Processing items from source")
				return
			}
		}(rssSource)

	}

	wg.Wait()

	return nil
}

func (f *Fetcher) processItems(ctx context.Context, source Source, items []models.Item) error {
	for _, item := range items {
		item.Date = item.Date.UTC()

		if f.itemShouldBeSkipped(item) {
			continue
		}

		err := f.articles.Store(ctx, models.Article{
			SourceID:    source.ID(),
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			PublishedAt: item.Date,
		})

		if err != nil {
			return err
		}

	}

	return nil

}

func (f *Fetcher) itemShouldBeSkipped(item models.Item) bool {
	categoriesSet := make(map[string]struct{})
	for _, category := range item.Categories {
		categoriesSet[category] = struct{}{}
	}

	for _, keyword := range f.filterKeywords {
		titleContainsKeyWord := strings.Contains(strings.ToLower(item.Title), keyword)

		if _, exists := categoriesSet[keyword]; exists || titleContainsKeyWord {
			return true
		}
	}

	return false
}
