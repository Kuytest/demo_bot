package wrapper

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"strconv"
	"syscall"

	pb "github.com/tinkoff/invest-api-go-sdk/proto"

	"github.com/tinkoff/invest-api-go-sdk/investgo"
	"go.uber.org/zap"
)

type Position struct {
	Amount float64
	Bonds  []Bonds
}

type Bonds struct {
	ExternalId int
	UID        string
	Title      string
	Count      float64
	CurPrice   float64
	CurSumm    float64
}

// создает структуру позиций
func NewPosition() *Position {
	return &Position{}
}

// создаем клиента для investAPI, он позволяет создавать нужные сервисы
func ActivateInvest(logger *zap.SugaredLogger) (*investgo.Client, *investgo.Config, context.CancelFunc, error) {
	config, err := investgo.LoadConfig("cmd/bot/config.yaml")
	if err != nil {
		log.Fatalf("config loading error %v", err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	client, err := investgo.NewClient(ctx, config, logger)
	if err != nil {
		logger.Fatalf("client creating error %v", err.Error())
	}

	return client, &config, cancel, err
}

// закрывает клиенты для InvestAPI командой
func SwitchoffInvest(logger *zap.SugaredLogger, client *investgo.Client, cancel context.CancelFunc) {
	logger.Infof("closing client connection")
	err := client.Stop()
	if err != nil {
		logger.Errorf("client shutdown error %v", err.Error())
	}

	defer cancel()
}

// создаем клиента для работы с портфелем
func CreateOperationsClient(client *investgo.Client) *investgo.OperationsServiceClient {
	operationsService := client.NewOperationsServiceClient()
	return operationsService
}

// базово заливает позиции по счету
func (p *Position) Upload(client *investgo.Client, opClient *investgo.OperationsServiceClient, config *investgo.Config, logger *zap.SugaredLogger) {
	portfolioResp, err := opClient.GetPortfolio(config.AccountId, pb.PortfolioRequest_RUB)
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		amountP := MoneyFloat64(portfolioResp.TotalAmountPortfolio.Units, portfolioResp.TotalAmountPortfolio.Nano)
		p.Amount = amountP
		for i := range portfolioResp.Positions {
			p.Bonds = append(p.Bonds, Bonds{
				ExternalId: i,
				UID:        portfolioResp.Positions[i].InstrumentUid,
				Title:      TakeNameUID(portfolioResp.Positions[i].InstrumentUid, client, logger),
				Count:      MoneyFloat64(portfolioResp.Positions[i].Quantity.Units, portfolioResp.Positions[i].Quantity.Nano),
				CurPrice:   MoneyFloat64(portfolioResp.Positions[i].CurrentPrice.Units, portfolioResp.Positions[i].CurrentPrice.Nano),
				CurSumm: MoneyFloat64(portfolioResp.Positions[i].Quantity.Units, portfolioResp.Positions[i].Quantity.Nano) *
					MoneyFloat64(portfolioResp.Positions[i].CurrentPrice.Units, portfolioResp.Positions[i].CurrentPrice.Nano),
			})
		}
	}
}

// чистит позиции структуры
func (p *Position) Clear() {
	p = &Position{}
}

// чистит позицию и снова загружает (использует оба доступных метода)
func (p *Position) Update(client *investgo.Client, opClient *investgo.OperationsServiceClient, config *investgo.Config, logger *zap.SugaredLogger) {
	p.Clear()
	p.Upload(client, opClient, config, logger)
}

func TakeNameUID(uid string, client *investgo.Client, logger *zap.SugaredLogger) string {
	instrumentsService := client.NewInstrumentsServiceClient()
	name, err := instrumentsService.BondByUid(uid)
	if err != nil {
		logger.Errorf(err.Error())
		return "not Bond"
	}
	return name.Instrument.Name
}

func MoneyFloat64(units int64, nano int32) float64 {
	unitsF := float64(units)
	nanoF, err := strconv.ParseFloat(fmt.Sprintf("0.%d", nano/10_000_000), 32)
	if err != nil {
		return 0.0
	}
	return unitsF + nanoF
}
