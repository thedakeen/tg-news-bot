package main

import (
	"context"
	"errors"
	"log"
	"news-bot/internal/bot"
	"news-bot/internal/bot/middleware"
	"news-bot/internal/config"
	"os"
	"os/signal"
	"syscall"
)

func (app *application) serve() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app.bot.RegisterCmdView("start", bot.ViewCmdStart())
	app.bot.RegisterCmdView("addsource", middleware.AdminOnly(
		config.Get().TelegramChannelID,
		bot.ViewCmdAddSource(app.sources),
	))

	app.bot.RegisterCmdView("listsources", middleware.AdminOnly(
		config.Get().TelegramChannelID,
		bot.ViewCmdListSources(app.sources),
	))

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

	err := app.bot.Run(ctx)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("[ERROR] failed to start bot: %v", err)
			return err
		}

		log.Printf("bot stopped")
	}

	return nil

}
