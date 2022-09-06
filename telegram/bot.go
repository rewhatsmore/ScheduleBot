package telegram

import (
	"context"
	"database/sql"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	db "schedule.sqlc.dev/app/db/sqlc"
	telegram "schedule.sqlc.dev/app/telegram/handlers"
)

func StartBot(token string, queries *db.Queries) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	telegram.Scheduler(queries, bot)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	handleUpdates(updates, bot, queries)
}

func handleUpdates(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, queries *db.Queries) {
	for update := range updates {
		if update.Message != nil { // If we got a message

			if update.Message.IsCommand() {
				if err := telegram.HandleCommand(update.Message, bot, queries); err != nil {
					// handleError(update.Message.Chat.ID, err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Извини, ошибка")
					bot.Send(msg)

				}
				continue
			}
			if err := telegram.HandleMessage(update.Message, bot, queries); err != nil {
				// b.handleError(update.Message.Chat.ID, err)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Извини, ошибка")
				bot.Send(msg)

			}
		}

		if update.CallbackQuery != nil {
			if err := telegram.HandleCallback(update.CallbackQuery, bot, queries); err != nil {
				// handleError(update.Message.Chat.ID, err)
				msg := tgbotapi.NewMessage(update.CallbackQuery.From.ID, "Извини, ошибка")
				bot.Send(msg)
				log.Println(err)
			}
		}

		if update.MyChatMember.NewChatMember.Status == "kicked" {
			if err := HandleDeleteUser(update.MyChatMember.From.ID, queries); err != nil {
				// handleError(update.Message.Chat.ID, err)
				log.Println(err)
			}
		}
	}
}

func HandleDeleteUser(userID int64, queries *db.Queries) error {
	err := queries.DeleteUser(context.Background(), userID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	return nil
}
