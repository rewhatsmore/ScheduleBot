package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	db "schedule.sqlc.dev/app/db/sqlc"
)

func StartBot(token string, queries *db.Queries) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	handleUpdates(updates, bot, queries)
}

func handleUpdates(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, queries *db.Queries) {
	for update := range updates {
		if update.Message != nil { // If we got a message

			if update.Message.IsCommand() {
				if err := handleCommand(update.Message, bot, queries); err != nil {
					// handleError(update.Message.Chat.ID, err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Извини, ошибка")
					bot.Send(msg)

				}
				continue
			}
			// if err := handleMessage(update.Message); err != nil {
			// 	// 	b.handleError(update.Message.Chat.ID, err)
			// 	// }
			// }
		}

		if update.CallbackQuery != nil {
			if err := handleCallback(update.CallbackQuery, bot, queries); err != nil {
				// handleError(update.Message.Chat.ID, err)
				msg := tgbotapi.NewMessage(update.CallbackQuery.From.ID, "Извини, ошибка")
				bot.Send(msg)
				log.Println(err)
			}
			continue
		}
	}
}
