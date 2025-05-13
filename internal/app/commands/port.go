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

func (c *Commander) GetPort(inputMessage *tgbotapi.Message) {
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
	sb.WriteString(fmt.Sprintf("Стоимость портфеля: %.2f\n", positions.Amount))

	for i := range positions.Bonds {

		sb.WriteString(fmt.Sprintf("%d: %s текущая стоимость: %.2f\n",
			i,
			positions.Bonds[i].Title,
			positions.Bonds[i].CurSumm,
		))
	}

	msg := tgbotapi.NewMessage(inputMessage.Chat.ID, sb.String())
	c.bot.Send(msg)

}

func GetPort(inputMessage *tgbotapi.Message, variant int) string {
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
	client, config, cancel, err := wrapper.ActivateInvest(logger, variant)
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
	sb.WriteString(fmt.Sprintf("Стоимость портфеля: %.2f\n", positions.Amount))

	for i := range positions.Bonds {

		sb.WriteString(fmt.Sprintf("%d: %s текущая стоимость: %.2f\n",
			i,
			positions.Bonds[i].Title,
			positions.Bonds[i].CurSumm,
		))
	}

	msg := sb.String()
	return msg

}
