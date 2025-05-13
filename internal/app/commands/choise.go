package commands

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (c *Commander) Choise(inputMessage *tgbotapi.Message) {
	outputMsgText := "Выбери счет для анализа: \n\n"

	msg := tgbotapi.NewMessage(inputMessage.Chat.ID, outputMsgText)

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ИИС-Pension", GetPort(inputMessage, 2)),
		),
	)

	c.bot.Send(msg)
}
