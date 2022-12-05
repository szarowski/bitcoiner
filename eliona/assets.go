//  This file is part of the eliona project.
//  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package eliona

import (
	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
	"github.com/valyala/fastjson"
	"strconv"
	"time"
)

const (
	// AssetTypeName Bitcoiner asset type name
	AssetTypeName = "bitcoin_currency_rate"
	// A project id for the asset entity
	projectID = "eliona"
)

// BitcoinCurrencyRate data type for the bitcoiner currency rate application
type BitcoinCurrencyRate struct {
	TimestampMilli int64   `json:"timestamp_milli"`
	Currency       string  `json:"currency"`
	Value          float64 `json:"value"`
	Description    string  `json:"description"`
}

// StoreAsset stores and logs collected asset and its data into the Eliona database
func StoreAsset(timestamp int64, dataSourceCurrencies interface{}, jsonModel fastjson.Value) {
	bitcoinCurrencyRateAssetTypeName := AssetTypeName
	assertTypeName := api.NewNullableString(&bitcoinCurrencyRateAssetTypeName)
	currencies := dataSourceCurrencies.([]string)
	// Log and store the data for every currency.
	for _, currency := range currencies {
		timeNow := time.Now()
		currentTime := api.NewNullableTime(&timeNow)
		currencyData := jsonModel.GetStringBytes("bpi", currency, "code")
		descriptionData := jsonModel.GetStringBytes("bpi", currency, "description")
		currencyRateData := jsonModel.GetFloat64("bpi", currency, "rate_float")

		// Log the currency rate data for particular currency
		log.Info("collectCurrencyData", "currency=%s\n", currencyData)
		log.Info("collectCurrencyData", "description=%s\n", descriptionData)
		log.Info("collectCurrencyData", "rate_float=%f\n", currencyRateData)

		a := api.Asset{
			ProjectId: projectID,
			// Generate a unique asset id up to millisecond (better than possible random value).
			GlobalAssetIdentifier: strconv.FormatInt(timeNow.UnixMilli(), 10),
			Name:                  *assertTypeName,
			AssetType:             bitcoinCurrencyRateAssetTypeName,
		}
		// Create the asset in the Eliona database
		assetID, err := asset.UpsertAsset(a)
		if err != nil {
			log.Warn("collectCurrencyData", "%s: %s", "failed to upsert asset for currency:", currencyData)
		}
		// Create the asset data in the Eliona database
		err = asset.UpsertData(getApiData(*assetID, *currentTime, timestamp, currencyData, currencyRateData, descriptionData, *assertTypeName))
		if err != nil {
			log.Warn("collectCurrencyData", "%s: %s", "failed to upsert data for currency:", currencyData)
		}
	}
}

// Gets the Eliona API data model
func getApiData(assetId int32, elionaTimestamp api.NullableTime, timestamp int64, currency []byte, value float64, description []byte, assetTypeName api.NullableString) api.Data {
	return api.Data{AssetId: assetId, Subtype: api.SUBTYPE_INPUT, Timestamp: elionaTimestamp,
		Data: common.StructToMap(
			BitcoinCurrencyRate{
				TimestampMilli: timestamp,
				Currency:       string(currency),
				Value:          value,
				Description:    string(description),
			}), AssetTypeName: assetTypeName}
}
