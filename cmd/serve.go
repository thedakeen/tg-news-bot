package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func (app *application) serve() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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

	//go func(ctx context.Context) {
	err := app.notifier.Start(ctx)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("failed to start notifier: %v", err)
			return err
		}
		log.Println("notifier stopped")
	}
	//}(ctx)

	return nil

}
