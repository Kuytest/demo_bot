package wrapper

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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
	CurPrice   float64
	Amount     float64
	CurSumm    float64
}

// создает структуру позиций
func NewPosition() *Position {
	return &Position{}
}

// создаем клиента для investAPI, он позволяет создавать нужные сервисы
func ActivateInvest(logger *zap.SugaredLogger, variant int) (*investgo.Client, *investgo.Config, context.CancelFunc, error) {
	var config investgo.Config
	var err error

	if variant == 1 {
		config = investgo.Config{AccountId: os.Getenv("ACC_1"),
			Token:                         os.Getenv("TOKEN_T"),
			EndPoint:                      "invest-public-api.tinkoff.ru:443",
			AppName:                       "invest-api-go-sdk",
			DisableResourceExhaustedRetry: false,
			DisableAllRetry:               false,
			MaxRetries:                    3,
		}
	} else if variant == 2 {
		config = investgo.Config{AccountId: os.Getenv("ACC_2"),
			Token:                         os.Getenv("TOKEN_T"),
			EndPoint:                      "invest-public-api.tinkoff.ru:443",
			AppName:                       "invest-api-go-sdk",
			DisableResourceExhaustedRetry: false,
			DisableAllRetry:               false,
			MaxRetries:                    3,
		}
	} else if variant == 3 {
		config = investgo.Config{AccountId: os.Getenv("ACC_3"),
			Token:                         os.Getenv("TOKEN_T"),
			EndPoint:                      "invest-public-api.tinkoff.ru:443",
			AppName:                       "invest-api-go-sdk",
			DisableResourceExhaustedRetry: false,
			DisableAllRetry:               false,
			MaxRetries:                    3,
		}
	} else {
		log.Fatalf("config loading error %s", "variant is not exist")
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
			if TakeDataUID(portfolioResp.Positions[i].InstrumentUid, client, logger) == "not Bond" {
				continue
			}
			price := MoneyFloat64(portfolioResp.Positions[i].CurrentPrice.Units, portfolioResp.Positions[i].CurrentPrice.Nano)
			p.Bonds = append(p.Bonds, Bonds{
				ExternalId: i,
				UID:        portfolioResp.Positions[i].InstrumentUid,
				Title:      TakeDataUID(portfolioResp.Positions[i].InstrumentUid, client, logger),
				CurPrice:   price,
				Amount:     MoneyFloat64(portfolioResp.Positions[i].Quantity.Units, portfolioResp.Positions[i].Quantity.Nano),
				CurSumm:    MoneyFloat64(portfolioResp.Positions[i].Quantity.Units, portfolioResp.Positions[i].Quantity.Nano) * price,
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

// вытягивает тайтл по UID
func TakeDataUID(uid string, client *investgo.Client, logger *zap.SugaredLogger) string {
	instrumentsService := client.NewInstrumentsServiceClient()
	instService, err := instrumentsService.BondByUid(uid)
	if err != nil {
		logger.Errorf(err.Error())
		return "not Bond"
	}
	return instService.Instrument.Name
}

// получаем будущие купоны по бумаге (все купоны минус НКД) и дней до погашения
func TakeIncomeUID(uid string, client *investgo.Client, logger *zap.SugaredLogger) (float64, int) {
	instrumentsService := client.NewInstrumentsServiceClient()
	// получаем список наших бондов в портфеле
	instService, err := instrumentsService.BondByUid(uid)
	if err != nil {
		logger.Errorf(err.Error())
		return 0.0, 0
	}
	// получаем ивенты по купонам из списка наших бондов по фиги и периоду с текущего до погашения
	bondCoup, err := instrumentsService.GetBondCoupons(instService.Instrument.Figi, time.Now(), instService.Instrument.MaturityDate.AsTime())
	if err != nil {
		logger.Errorf(err.Error())
		return 0.0, 0
	}
	// собираем дни до погашения
	days := int(time.Until(instService.Instrument.MaturityDate.AsTime()).Hours() / 24)
	var all float64
	for i := range bondCoup.Events {
		all += MoneyFloat64(bondCoup.Events[i].PayOneBond.Units, bondCoup.Events[i].PayOneBond.Nano)
	}
	// вычитаем НКД
	all = all - MoneyFloat64(instService.Instrument.AciValue.Units, instService.Instrument.AciValue.Nano)
	return all, days

}

// получаем все купоны на 12 месяцев вперед
func TakeAllIncomeIn12month(uid string, client *investgo.Client, logger *zap.SugaredLogger, amount float64) [12]float64 {
	var coup_12month [12]float64

	instrumentsService := client.NewInstrumentsServiceClient()
	// получаем список наших бондов в портфеле
	instService, err := instrumentsService.BondByUid(uid)
	if err != nil {
		logger.Errorf(err.Error())
		return [12]float64{}
	}

	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	next12month := firstOfMonth.AddDate(0, 12, 0)
	nowMonth := int(now.Month())

	// получаем ивенты по купонам из списка наших бондов по фиги и периоду с начала текущего месяца и плюс 12 мес
	bondCoup, err := instrumentsService.GetBondCoupons(instService.Instrument.Figi, firstOfMonth, next12month)
	if err != nil {
		logger.Errorf(err.Error())
		return [12]float64{}
	}
	for i := range bondCoup.Events {
		vm := int(bondCoup.Events[i].CouponDate.AsTime().Month()) - nowMonth
		if vm < 0 {
			vm = 12 + vm
		}
		coup_12month[vm] += MoneyFloat64(bondCoup.Events[i].PayOneBond.Units, bondCoup.Events[i].PayOneBond.Nano) * amount
	}
	return coup_12month
}

// используется для приображения юнитов и нано в флоат 64 до двух знаков
func MoneyFloat64(units int64, nano int32) float64 {
	unitsF := float64(units)
	nanoF, err := strconv.ParseFloat(fmt.Sprintf("0.%d", nano/10_000_000), 32)
	if err != nil {
		return 0.0
	}
	return unitsF + nanoF
}
