package memsql

import(
	log "github.com/sirupsen/logrus"
	"factors/model/model"
	C "factors/config"
	"github.com/jinzhu/gorm"
	U "factors/util"
	"errors"
	"fmt"
)
// GetFactorsTrackedEvent - Get details of tracked event
func (store *MemSQL) GetCurrencyDetails(currency string, date int64) ([]model.Currency, error) {

	db := C.GetServices().Db

	var currencyDetails []model.Currency
	if err := db.Table("currency").Where("currency = ? AND date = ?", currency, date).First(&currencyDetails).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return []model.Currency{}, nil
		}
		return nil, err
	}
	return currencyDetails, nil
}

func (store *MemSQL) CreateCurrencyDetails(currency string, date int64, value float64) (error){
	db := C.GetServices().Db

	currencyObj := model.Currency{
			Currency   : currency,
			InrValue   : value,
			Date       : date,
			CreatedAt  : U.TimeNowZ(),
			UpdatedAt  : U.TimeNowZ(),
	}

	existingCurrency, err := store.GetCurrencyDetails(currency, date)
	if(err != nil){
		return err
	} 
	if (len(existingCurrency) > 0){
		return errors.New(fmt.Sprintf("Value already exist for the currency %v, date  %v combination", currency, date))
	}
	if err := db.Table("currency").Create(&currencyObj).Error; err != nil {
		log.WithError(err).Error("Failure to insert currency details")
		return err
	}
	return nil
}