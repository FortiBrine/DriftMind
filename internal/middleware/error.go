package middleware

import (
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func NewCustomErrorHandler(logger *slog.Logger) th.ErrorHandler {
	return func(
		ctx *th.Context,
		update telego.Update,
		err error,
	) {
		if err == nil {
			return
		}

		attrs := []any{"update_id", update.UpdateID, "error", err}

		switch {
		case update.Message != nil:
			attrs = append(attrs,
				"chat_id", update.Message.Chat.ID,
				"from_id", update.Message.From.ID,
				"message_id", update.Message.MessageID,
			)
		case update.CallbackQuery != nil:
			attrs = append(attrs, "from_id", update.CallbackQuery.From.ID)
			if update.CallbackQuery.Message != nil {
				attrs = append(attrs, "chat_id", update.CallbackQuery.Message.GetChat().ID)
			}
		}

		logger.Error("error handling update", attrs...)
	}
}
