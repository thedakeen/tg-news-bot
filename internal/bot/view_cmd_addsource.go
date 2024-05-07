package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"news-bot/internal/botkit"
	"news-bot/internal/models"
)

type SourceStorage interface {
	Add(ctx context.Context, source models.Source) (int64, error)
}

func ViewCmdAddSource(storage SourceStorage) botkit.ViewFunc {
	type addSourceArgs struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		args, err := botkit.ParseJSON[addSourceArgs](update.Message.CommandArguments())
		if err != nil {
			message := "Invalid input."
			_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, message))
			return err
		}

		source := models.Source{
			Name:    args.Name,
			FeedURL: args.URL,
		}

		sourceID, err := storage.Add(ctx, source)
		if err != nil {
			return err
		}

		var (
			msgText = fmt.Sprintf(
				"Source added with ID: `%d`\\ Use this ID for managing source\\.", sourceID)
			reply = tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		)

		reply.ParseMode = "MarkdownV2"

		_, err = bot.Send(reply)
		if err != nil {
			return err
		}

		return nil
	}
}
