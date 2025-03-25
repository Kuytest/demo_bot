package main

import (
	"log"
	"os"
	"telebot/internal/app/commands"
	"telebot/internal/service/product"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// zapConfig := zap.NewDevelopmentConfig()
	// zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	// zapConfig.EncoderConfig.TimeKey = "time"
	// l, err := zapConfig.Build()
	// Logger := l.Sugar()
	// defer func() {
	// 	err := Logger.Sync()
	// 	if err != nil {
	// 		log.Println(err.Error())
	// 	}
	// }()
	// if err != nil {
	// 	log.Fatalf("logger creating error %v", err)
	// }

	godotenv.Load()
	token := os.Getenv("TOKEN")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.UpdateConfig{
		Timeout: 60,
	}

	updates := bot.GetUpdatesChan(u)

	productService := product.NewService()

	commander := commands.NewCommander(bot, productService)

	for update := range updates {
		commander.HandleUpdate(update)
	}
}
