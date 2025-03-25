package commands

import (
	"fmt"
	"log"
	"telebot/internal/service/wrapper"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (c *Commander) GetPort(inputMessage *tgbotapi.Message) {
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
	client, config, cancel, err := wrapper.ActivateInvest(logger)
	if err != nil {
		logger.Fatalf("client creating error %v", err.Error())
	}
	// убрали закрытие клиента в отдельный метод
	// defer func() {
	// 	main.Logger.Infof("closing client connection")
	// 	err := client.Stop()
	// 	if err != nil {
	// 		main.Logger.Errorf("client shutdown error %v", err.Error())
	// 	}
	// }()

	// deferim клиента InvestAPI и отменяем контекст
	defer wrapper.SwitchoffInvest(logger, client, cancel)

	// создаем клиента операций для сбора данных по портфелю
	operationsService := wrapper.CreateOperationsClient(client)

	// создаем клиента операций для сбора данных по портфелю
	// operationsService := client.NewOperationsServiceClient()

	// создаем пустую позицию
	positions := wrapper.NewPosition()

	// загружаем позицию
	positions.Upload(client, operationsService, config, logger)

	// выводим сообщение позиции в ТГ
	text := fmt.Sprintf("Стоимость портфеля: %.2f\n", positions.Amount)
	for i := range positions.Bonds {
		text += fmt.Sprintf("Позиция %d: %s текущая стоимость: %.2f\n",
			i,
			positions.Bonds[i].Title,
			positions.Bonds[i].CurSumm)
	}

	msg := tgbotapi.NewMessage(inputMessage.Chat.ID, text)
	c.bot.Send(msg)

	// запрашиваем портфель в рублях и отдаем стоимость активов (облиги например)
	// portfolioResp, err := operationsService.GetPortfolio(config.AccountId, pb.PortfolioRequest_RUB)
	// if err != nil {
	// 	logger.Errorf(err.Error())
	// } else {
	// 	// юниты - целая часть, нано - дробная. В сообщение собираем стоимость портфеля просто через точку
	// 	units := portfolioResp.TotalAmountPortfolio.Units
	// 	nano := portfolioResp.TotalAmountPortfolio.Nano
	// 	dohunits := portfolioResp.ExpectedYield.Units
	// 	dohnano := portfolioResp.ExpectedYield.Nano
	// 	var positions string
	// 	for i := range portfolioResp.Positions {
	// 		positions += fmt.Sprintf("Позиция %d: %s %d"+"."+"%d\n",
	// 			i,
	// 			wrapper.TakeNameUID(portfolioResp.Positions[i].InstrumentUid, client, logger),
	// 			portfolioResp.Positions[i].Quantity.Units,
	// 			portfolioResp.Positions[i].Quantity.Nano/10000000)
	// 	}

	// 	text := fmt.Sprintf("Стоимость портфеля %s:\n%d"+"."+"%d\n"+"%d"+"."+"%d",
	// 		config.AccountId,
	// 		units,
	// 		nano/10000000,
	// 		dohunits,
	// 		dohnano/10000000)
	// 	msg_2 := tgbotapi.NewMessage(inputMessage.Chat.ID, text)
	// 	msg_3 := tgbotapi.NewMessage(inputMessage.Chat.ID, positions)
	// 	c.bot.Send(msg_2)
	// 	c.bot.Send(msg_3)
	// }

}
