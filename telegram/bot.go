package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"schedule.sqlc.dev/app/conf"
	db "schedule.sqlc.dev/app/db/sqlc"
	telegram "schedule.sqlc.dev/app/telegram/handlers"
)

func StartBot(config conf.Config, queries *db.Queries) {
	bot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	telegram.Scheduler(queries, bot)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	handleUpdates(updates, bot, queries, config)
}

func handleUpdates(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, queries *db.Queries, config conf.Config) {
	for update := range updates {
		if update.Message != nil { // If we got a message

			if update.Message.IsCommand() {
				if err := telegram.HandleCommand(update.Message, bot, queries); err != nil {
					msg := telegram.HandleError(update.Message.Chat.ID, config.AdminID, err)
					msg.SendMsg(bot)
				}
				continue
			}
			if err := telegram.HandleMessage(update.Message, bot, queries); err != nil {
				msg := telegram.HandleError(update.Message.Chat.ID, config.AdminID, err)
				msg.SendMsg(bot)

			}
		}

		if update.CallbackQuery != nil {
			if err := telegram.HandleCallback(update.CallbackQuery, bot, queries); err != nil {
				msg := telegram.HandleError(update.CallbackQuery.From.ID, config.AdminID, err)
				// TO DO: add back button
				msg.UpdateMsg(bot, update.CallbackQuery.Message)
				log.Println(err)
			}
		}

		if update.MyChatMember != nil && update.MyChatMember.NewChatMember.Status == "kicked" {
			if err := telegram.HandleDeleteUser(update.MyChatMember.From.ID, queries); err != nil {
				msg := telegram.HandleError(update.CallbackQuery.From.ID, config.AdminID, err)
				msg.UpdateMsg(bot, update.CallbackQuery.Message)
			}
		}
	}
}
