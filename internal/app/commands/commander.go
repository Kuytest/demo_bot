package commands

import (
	"telebot/internal/service/product"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Commander struct {
	bot            *tgbotapi.BotAPI
	productService *product.Service
}

func NewCommander(bot *tgbotapi.BotAPI, productService *product.Service) *Commander {
	return &Commander{
		bot:            bot,
		productService: productService,
	}
}

func (c *Commander) HandleUpdate(update tgbotapi.Update) {

	if update.CallbackQuery != nil {
		msg := tgbotapi.NewMessage(
			update.CallbackQuery.Message.Chat.ID,
			"Data: "+update.CallbackQuery.Data)
		c.bot.Send(msg)
		return
	}

	switch update.Message.Command() {
	// case "choise":
	// 	c.Choise(update.Message)
	case "help":
		c.Help(update.Message)
	case "list":
		c.List(update.Message)
	case "get":
		c.Get(update.Message)
	case "getport":
		c.GetPort(update.Message)
	case "income":
		c.GetIncome(update.Message)
	default: // If we got ordinary message
		c.Default(update.Message)
	}
}
