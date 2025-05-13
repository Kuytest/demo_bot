package commands

import (
	"fmt"
	"log"
	"strings"
	"telebot/internal/service/wrapper"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (c *Commander) GetIncome(inputMessage *tgbotapi.Message) {
	// унести логгер в мейн
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	zapConfig.EncoderConfig.TimeKey = "time"
	l, err := zapConfig.Build()
	logger := l.Sugar()
	defer func() {
		err := logger.Sync()
		if err != nil {
			log.Println(err.Error())
		}
	}()
	if err != nil {
		log.Fatalf("logger creating error %v", err)
	}

	// активируем клиента InvestAPI
	client, config, cancel, err := wrapper.ActivateInvest(logger, 1)
	if err != nil {
		logger.Fatalf("client creating error %v", err.Error())
	}

	// deferim клиента InvestAPI и отменяем контекст
	defer wrapper.SwitchoffInvest(logger, client, cancel)

	// создаем клиента операций для сбора данных по портфелю
	operationsService := wrapper.CreateOperationsClient(client)

	// создаем пустую позицию
	positions := wrapper.NewPosition()

	// загружаем позицию
	positions.Upload(client, operationsService, config, logger)

	// выводим сообщение позиции в ТГ
	sb := strings.Builder{}
	var n12m [12]float64

	for i := range positions.Bonds {
		v12m := wrapper.TakeAllIncomeIn12month(positions.Bonds[i].UID, client, logger, positions.Bonds[i].Amount)
		for j := range n12m {
			n12m[j] += v12m[j]
		}
	}

	for i := range n12m {
		sb.WriteString(fmt.Sprintf("Пассив за %s: %.2f\n",
			time.Now().AddDate(0, i, 0).Month().String(),
			n12m[i],
		))
	}

	msg := tgbotapi.NewMessage(inputMessage.Chat.ID, sb.String())
	c.bot.Send(msg)

}
