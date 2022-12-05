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

package conf

import (
	"github.com/eliona-smart-building-assistant/go-eliona/app"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
	"strings"
	"time"
)

type BitcoinerAppConfigType int8

// Keys to the preconfigured configuration map
const (
	DataCollectionDuration BitcoinerAppConfigType = iota
	DataSourceUrl
	SupportedCurrencies
)

// A list of available currencies to be supported
var availableCurrencies = []string{"USD", "EUR", "GBP"}

// Preconfigured map filled in after configuration initialization
var bitcoinerAppConfigMap map[BitcoinerAppConfigType]interface{}

// BitcoinerAppConfig The configuration type of the bitcoiner application
type BitcoinerAppConfig struct {
	DataCollectionDuration string
	DataSourceURL          string
	SupportedCurrencies    string
}

// GetBitcoinerAppConfigValue gets a configured value from the preconfigured map
func GetBitcoinerAppConfigValue(configKey BitcoinerAppConfigType) interface{} {
	// Finish the application in case of missing configuration because all configuration values should have at least default values.
	if bitcoinerAppConfigMap == nil {
		log.Error("GetBitcoinerAppConfigValue", "%s", "failed to parse config values")
		return nil
	}

	return bitcoinerAppConfigMap[configKey]
}

// parseBitcoinerAppConfiguration parses the bitcoiner application configuration yaml, by default bitcoiner.yaml
func parseBitcoinerAppConfiguration() error {
	// Get a configuration file name path from BITCOINER_CONFIG_PATH env variable or bitcoiner.yaml if not present in the project.
	configPath := common.Getenv("BITCOINER_CONFIG_PATH", "./conf/")
	// Get the viperConfigurer YAML configurer.
	viperConfigurer := configViper(configPath, app.AppName())
	err := viperConfigurer.ReadInConfig()
	if err != nil {
		log.Warn("loadBitcoinerAppConfiguration", "%s - %s", "failed to find config file", err)
	}
	bitcoinerConfig := &BitcoinerAppConfig{
		DataCollectionDuration: "1s",
		DataSourceURL:          "https://api.coindesk.com/v1/bpi/currentprice.json",
		SupportedCurrencies:    "USD",
	}
	// Parse the configuration values.
	err = viperConfigurer.Unmarshal(bitcoinerConfig)
	if err != nil {
		log.Warn("loadBitcoinerAppConfiguration", "%s - %s", "failed to unmarshall config data", err)
	}
	bitcoinerAppConfigMap = map[BitcoinerAppConfigType]interface{}{}
	// Parse a duration for the main loop with default value set to 1 second.
	duration, err := time.ParseDuration(bitcoinerConfig.DataCollectionDuration)
	if err != nil {
		log.Warn("InitEliona", "%s - %s", "error parsing duration, defaulting", err)
	}
	// If the duration is less than 1 second default to 1 second to not exceed CoinDesk rate limit.
	if duration < time.Second {
		log.Warn("InitEliona", "%s", "duration is configured to less than 1 second, defaulting")
		duration = time.Second
	}
	bitcoinerAppConfigMap[DataCollectionDuration] = duration
	// Get the data source URL address with default value in case of error
	bitcoinerAppConfigMap[DataSourceUrl] = bitcoinerConfig.DataSourceURL
	// Get the supported currencies with default value in case of error
	bitcoinerAppConfigMap[SupportedCurrencies] = []string{"USD"}
	if strings.TrimSpace(bitcoinerConfig.SupportedCurrencies) != "" {
		parsedCurrencies := strings.Split(strings.TrimSpace(bitcoinerConfig.SupportedCurrencies), ",")
		for _, currency := range parsedCurrencies {
			c := strings.TrimSpace(currency)
			if c != "" && slices.Contains(availableCurrencies, c) && c != "USD" {
				bitcoinerAppConfigMap[SupportedCurrencies] = append(bitcoinerAppConfigMap[SupportedCurrencies].([]string), c)
			}
		}
	}

	return nil
}

// confiViper configures the viper YAML file parser
func configViper(configPath string, configName string) *viper.Viper {
	err := viper.BindEnv(configPath)
	if err != nil {
		log.Error("configViper", "%s - %s", "failed to setup config reader", err)
		return nil
	}
	v := viper.New()
	// Search for the config file (in case of default path is not set) in the path composed of: configPath + configName + .yaml
	v.SetConfigName(configName)
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	return v
}
