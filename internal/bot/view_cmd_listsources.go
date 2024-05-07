package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samber/lo"
	"news-bot/internal/botkit"
	"news-bot/internal/botkit/markup"
	"news-bot/internal/models"
	"strings"
)

type SourceLister interface {
	Sources(ctx context.Context) ([]models.Source, error)
}

func ViewCmdListSources(lister SourceLister) botkit.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		sources, err := lister.Sources(ctx)
		if err != nil {
			return err
		}

		var (
			sourceInfos = lo.Map(sources, func(source models.Source, _ int) string {
				return formatSource(source)
			})
			msgText = fmt.Sprintf(
				"List of sources \\(total %d\\):\n\n%s",
				len(sources),
				strings.Join(sourceInfos, "\n\n"))
		)

		reply := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		reply.ParseMode = "MarkdownV2"

		_, err = bot.Send(reply)
		if err != nil {
			return err
		}

		return nil
	}
}

func formatSource(source models.Source) string {
	return fmt.Sprintf(
		"*%s*\nID: `%d`\nURL: %s",
		markup.EscapeForMarkdown(source.Name),
		source.ID,
		markup.EscapeForMarkdown(source.FeedURL))
}
