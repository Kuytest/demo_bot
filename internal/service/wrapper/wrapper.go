package wrapper

import (
	"github.com/tinkoff/invest-api-go-sdk/investgo"
	"go.uber.org/zap"
)

// type Position []struct {
// 	Title    string
// 	Count    int
// 	CurPrice int
// }

func TakeNameUID(uid string, client *investgo.Client, logger *zap.SugaredLogger) string {
	instrumentsService := client.NewInstrumentsServiceClient()
	name, err := instrumentsService.BondByUid(uid)
	if err != nil {
		logger.Errorf(err.Error())
		return "not Bond"
	}
	return name.Instrument.Name
}

func MoneyString(units, nano int) {

}
