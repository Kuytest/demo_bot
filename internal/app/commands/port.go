package commands

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"telebot/internal/service/wrapper"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	pb "github.com/tinkoff/invest-api-go-sdk/proto"

	"github.com/tinkoff/invest-api-go-sdk/investgo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (c *Commander) GetPort(inputMessage *tgbotapi.Message) {
	config, err := investgo.LoadConfig("cmd/bot/config.yaml")
	if err != nil {
		log.Fatalf("config loading error %v", err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()
	// сдк использует для внутреннего логирования investgo.Logger
	// для примера передадим uber.zap
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
	// создаем клиента для investAPI, он позволяет создавать нужные сервисы и уже
	// через них вызывать нужные методы
	client, err := investgo.NewClient(ctx, config, logger)
	if err != nil {
		logger.Fatalf("client creating error %v", err.Error())
	}
	// убрали закрытие клиента в отдельный метод
	defer func() {
		logger.Infof("closing client connection")
		err := client.Stop()
		if err != nil {
			logger.Errorf("client shutdown error %v", err.Error())
		}
	}()

	// создаем клиента операций для сбора данных по портфелю
	operationsService := client.NewOperationsServiceClient()

	// запрашиваем портфель в рублях и отдаем стоимость активов (облиги например)
	portfolioResp, err := operationsService.GetPortfolio(config.AccountId, pb.PortfolioRequest_RUB)
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		// юниты - целая часть, нано - дробная. В сообщение собираем стоимость портфеля просто через точку
		units := portfolioResp.TotalAmountPortfolio.Units
		nano := portfolioResp.TotalAmountPortfolio.Nano
		dohunits := portfolioResp.ExpectedYield.Units
		dohnano := portfolioResp.ExpectedYield.Nano
		var positions string
		for i := range portfolioResp.Positions {
			positions += fmt.Sprintf("Позиция %d: %s %d"+"."+"%d\n",
				i,
				wrapper.TakeNameUID(portfolioResp.Positions[i].InstrumentUid, client, logger),
				portfolioResp.Positions[i].Quantity.Units,
				portfolioResp.Positions[i].Quantity.Nano/10000000)
		}

		text := fmt.Sprintf("Стоимость портфеля %s:\n%d"+"."+"%d\n"+"%d"+"."+"%d",
			config.AccountId,
			units,
			nano/10000000,
			dohunits,
			dohnano/10000000)
		msg_2 := tgbotapi.NewMessage(inputMessage.Chat.ID, text)
		msg_3 := tgbotapi.NewMessage(inputMessage.Chat.ID, positions)
		c.bot.Send(msg_2)
		c.bot.Send(msg_3)
	}
}
