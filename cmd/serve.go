package main

import (
	"context"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"news-bot/internal/bot"
	"news-bot/internal/botkit"
	"news-bot/internal/config"
	"os"
	"os/signal"
	"syscall"
)

func (app *application) serve() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	botAPI, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Printf("failed to create bot: %v", err)
		return err
	}

	newsBot := botkit.New(botAPI)
	newsBot.RegisterCmdView("start", bot.ViewCmdStart())

	go func(ctx context.Context) {
		err := app.fetcher.Start(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("failed to start fetcher: %v", err)
				return
			}
			log.Println("fetcher stopped")
		}
	}(ctx)

	go func(ctx context.Context) {
		err := app.notifier.Start(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("failed to start notifier: %v", err)
				return
			}
			log.Println("notifier stopped")
		}
	}(ctx)

	err = newsBot.Run(ctx)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("[ERROR] failed to start bot: %v", err)
			return err
		}

		log.Printf("bot stopped")
	}

	return nil

}
