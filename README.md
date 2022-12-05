# Bitcoiner App

The application for collecting bitcoin rates for USD, EUR or/and GBP currencies.
The project is developed based on this [template](https://docs.github.com/en/repositories/creating-and-managing-repositories/creating-a-repository-from-a-template) can be used to create an app stub for an [Eliona](https://www.eliona.io/) enviroment.
The application periodically (configurable value) collects the bitcoin data from the [CoinDesk](https://api.coindesk.com/v1/bpi/currentprice.json) current rates.
It is possible to specify which of the mentioned currencies are to be collected or some combination of them or all of them.
The Bitcoiner application can also specify another URL for the current rates of the CoinDesk service in case the address changes.
In other words the format of data must be consistent so it is convenient to change it only if the address changed.

## Configuration

There is a configuration file which can sets some project-related settings at the start time.
The name is `bitcoiner.yaml` and it is possible to specify a path to this file in `BITCOINER_CONFIG_PATH` environmental variable.
By default, the following project file is used (with description in comments):

```
DataCollectionDuration: 10s                                       # interval of collection (1 s default)
DataSourceURL: https://api.coindesk.com/v1/bpi/currentprice.json  # url of the data source of rates (https://api.coindesk.com/v1/bpi/currentprice.json default)
SupportedCurrencies: USD,EUR,GBP                                  # comma-separated currencies to collect (USD default)
```

If the `bitcoiner.yaml` file is not present, the default values in comments are used.

### Environment variables

- `BITCOINER_CONFIG_PATH`: can be se to a path from which is read the `bitcoiner.yaml` file with the configuration. If not present, it defaults to `./conf/` and reads the file from the project directory.

- `APPNAME`: set to `bitcoiner`.

- `CONNECTION_STRING`: configures the [Eliona database](https://github.com/eliona-smart-building-assistant/go-eliona/tree/main/db). Otherwise, the app can't be initialized and started. (e.g. `postgres://user:pass@localhost:5432/iot`)

- `API_ENDPOINT`: configures the endpoint to access the [Eliona API v2](https://github.com/eliona-smart-building-assistant/eliona-api). Otherwise, the app can't be initialized and started. (e.g. `http://api-v2:3000/v2` or `http://localhost:3000/v2` by default)

- `API_TOKEN`: defines the secret to authenticate the app and access the API. The default value for the Eliona mock is `secret`. 
  
- `API_SERVER_PORT`(optional): define the port the API server listens. The default value is Port `8082`.

- `DEBUG_LEVEL`(optional): defines the minimum level that should be [logged](https://github.com/eliona-smart-building-assistant/go-eliona/tree/main/log). Not defined the default level is `info`.

### Database tables ###

There are no custom database objects the app needs for configuration. All configuration values are managed by `bitcoiner.yaml` file. 

### App API ###

The app provides its own API to access configuration data and other functions. The full description of the API is defined in the `openapi.yaml` OpenAPI definition file.

- [API Reference](./openapi.yaml) shows Details of the API

### Eliona ###

The data of the application is managed solely by Eliona database tables, the following tables are used in particular:
- `eliona_app` table, where the `bitcoiner` app is stored when the app is started for the first time.

- `asset_type` table, where the app asset type is used. The asset type values are upserted as follows:
```
var (
	AssetTypeName = "bitcoin_currency_rate"
	custom     = true
	vendorName = "ITEC AG"
)
...
asset.UpsertAssetType(api.AssetType{Name: AssetTypeName, Custom: &custom, Vendor: *vendor})
```
- `asset` table, where the app asset is stored on every currency rate collection. To ensure uniqueness of the asset_id keys, which are generated on upsert, `GlobalAssetIdentifier` is set to unix timestamp in milliseconds. Asset is upserted as follows:
```
a := api.Asset{
    ProjectId: projectID,
    GlobalAssetIdentifier: strconv.FormatInt(timeNow.UnixMilli(), 10),
    Name:                  *assertTypeName,
    AssetType:             bitcoinCurrencyRateAssetTypeName,
}
```
ProjectID is `eliona`, Name is `bitcoin_currency_rate` as well as the AssetType to ensure the relation to the heap. 
There is `asset_id` generated on asset upsert which in turn is used for insertion of the data into the following table:
 - `heap` table, where the JSON representation of the following asset model is stored on every data collection:
```
type BitcoinCurrencyRate struct {
	TimestampMilli int64   `json:"timestamp_milli"`
	Currency       string  `json:"currency"`
	Value          float64 `json:"value"`
	Description    string  `json:"description"`
}
```
The meaning of the properties is the following:
- `TimestampMilli` is a UNIX timestamp in milliseconds at UTC time representing the time of generated rate for a particular currency.
    It is equivalent to the `updatedISO` JSON property in the response from the CoinDesk bitcoin rate service.
- `Currency` is one of the following (only possible values limited by application as well as the service):
  - USD (always default)
  - EUR
  - GBP
- `Value` is the actual rate for the currency. It is represented as float in double precision.
- `Description` is just a description of the currency.

## Tools

Docker and docker-compose are the only tools I have in mind, except of the GoLang platform and the Eliona ecosystem.
I use IntelliJ IDEA Ultimate edition with GO plugin for the development.

### Running the Application ###

I tried to minimise the changes done in the app template, so the app name remains `hailo`.
The environmental properties for local start are the following:
```
export CONNECTION_STRING=postgres://postgres:secret@localhost:5432
export API_ENDPOINT=http://localhost:3000/v2
export API_TOKEN=secret
export API_SERVER_PORT=8082
export APPNAME=bitcoiner
export DEBUG_LEVEL=info
```

There are 3 possible ways to start the application:
1. By running `go build ./...` and `go run hailo`. Before that there is required to start `docker-compose up -d ` in `eliona-mock-develop` or using the `docker-compose.yaml` in this project which just extends the set of containers.
2. By creating a Docker container by running:
  - `docker build -t localhost:5001/bitcoiner .`
  - `docker run --network eliona-mock-network localhost:5001/bitcoiner` without `push` the container image into local repository
3. By running `docker-compose` directly after building the local docker image:
- `docker build -t localhost:5001/bitcoiner .`
- `docker-compose up -d` for running the local mocked Eliona ecosystem with the Bitcoiner app

#### Note
Please note I reused the `docker-compose.yaml` and `init.sql` files from the [Eliona Mock](https://github.com/eliona-smart-building-assistant/eliona-mock/) repository in the bitcoiner application sample for simpler use. 

### REST Api of the Bitcoiner Application ###

There is an administration REST API for the Bitcoiner App described above, but it is slightly modified, so that it is more useful.
Here are some examples to use it:

- GET endpoint to retrieve collected data so far:
  - `curl 'http://localhost:8082/v1/examples' -H 'x-api-key secret'`
- POST endpoint to update one of the collected data:
  - `curl 'http://localhost:8082/v1/examples' -H 'x-api-key secret' -d '{
    "id": 1,
    "data": {
    "timestamp_milli": 1670192520000,
    "currency": "CZK",
    "value": 14301.3333,
    "description": "British Pound Sterling"
    }
    }'`
