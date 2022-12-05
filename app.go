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

package main

import (
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
	"github.com/valyala/fastjson"
	"hailo/apiserver"
	"hailo/apiservices"
	"hailo/conf"
	"hailo/eliona"
	"io"
	"net/http"
	"time"
)

// collectCurrencyData is the main app function which is called periodically
func collectCurrencyData() {
	// Get the data source URL of the CoinDesk bitcoin rate provider and collect the current data
	dataSourceUrl := conf.GetBitcoinerAppConfigValue(conf.DataSourceUrl)
	resp, err := http.Get(dataSourceUrl.(string))
	if err != nil {
		log.Error("collectCurrencyData", "%s - %s", "Data are not accessible on DataSourceUrl", err)
		return
	}
	// Close correctly the response body reader
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Warn("collectCurrencyData", "%s - %s", "error closing http read closer", err)
		}
	}(resp.Body)

	// Get the response body as the byte array
	body, err := io.ReadAll(resp.Body)

	// Parse the JSON data in the body
	var parser fastjson.Parser
	value, err := parser.ParseBytes(body)
	if err != nil {
		log.Fatal("collectCurrencyData", "%s - %s", "error parsing currency data", err)
	}
	// Get the timestamp in the ISO format (in RFC3339 format)
	updatedBytes := string(value.GetStringBytes("time", "updatedISO"))
	updated, err := time.Parse(time.RFC3339, updatedBytes)
	if err != nil {
		log.Warn("collectCurrencyData", "%s - %s", "error parsing currency timestamp")
	}
	timestamp := updated.UnixMilli()
	log.Info("collectCurrencyData", "updated=%d\n", timestamp)

	// Get the preconfigured currencies to be collected
	currencies := conf.GetBitcoinerAppConfigValue(conf.SupportedCurrencies)
	// Store the asset and data
	eliona.StoreAsset(timestamp, currencies, *value)
}

// listenApiRequests starts an API server and listen for API requests
// The API endpoints are defined in the openapi.yaml file
func listenApiRequests() {
	err := http.ListenAndServe(":"+common.Getenv("API_SERVER_PORT", "3000"), apiserver.NewRouter(
		apiserver.NewConfigurationApiController(apiservices.NewConfigurationApiService()),
	))
	log.Fatal("Hailo", "Error in API Server: %v", err)
}
