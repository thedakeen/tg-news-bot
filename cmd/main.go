package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"

	"news-bot/internal/config"
	"news-bot/internal/fetcher"
	"news-bot/internal/notifier"
	"news-bot/internal/storage"
	"news-bot/internal/summary"
	"sync"
)

type application struct {
	articles *storage.ArticlePostgresStorage
	sources  *storage.SourcePostgresStorage
	fetcher  fetcher.Fetcher
	notifier notifier.Notifier

	wg sync.WaitGroup
}

func main() {
	botAPI, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Printf("failed to create bot: %v", err)
		return
	}

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)
	if err != nil {
		log.Printf("failed to connect to database %v", err)
		return
	}
	defer db.Close()

	app := &application{
		articles: storage.NewArticlesStorage(db),
		sources:  storage.NewSourcesStorage(db),
	}

	app.fetcher = *fetcher.New(
		app.articles,
		app.sources,
		config.Get().FetchInterval,
		config.Get().FilterKeywords)

	app.notifier = *notifier.New(
		app.articles,
		summary.NewOpenAISummarizer(config.Get().OpenAIKey, config.Get().OpenAIPrompt),
		botAPI,
		config.Get().NotificationInterval,
		2*config.Get().FetchInterval,
		config.Get().TelegramChannelID)

	err = app.serve()
	if err != nil {
		log.Fatal(err)
	}
}
