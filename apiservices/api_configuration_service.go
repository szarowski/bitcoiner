/*
 * App template API
 *
 * API to access and configure the app template
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package apiservices

import (
	"context"
	"encoding/json"
	"errors"
	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-eliona/client"
	"github.com/eliona-smart-building-assistant/go-utils/log"
	"hailo/apiserver"
	"hailo/eliona"
	"net/http"
	"sort"
	"time"
)

// ConfigurationApiService is a service that implements the logic for the ConfigurationApiServicer
// This service should implement the business logic for every endpoint for the ConfigurationApi API.
// Include any external packages or services that will be required by this service.
type ConfigurationApiService struct {
	// The data api service to PUT and GET the data values
	dataApiService *api.DataApiService
}

// NewConfigurationApiService creates a default api service
func NewConfigurationApiService() apiserver.ConfigurationApiServicer {
	return &ConfigurationApiService{dataApiService: client.NewClient().DataApi}
}

// GetExamples - Get example configuration
func (s *ConfigurationApiService) GetExamples(context.Context) (apiserver.ImplResponse, error) {
	// Get the data request with authentication context
	request := s.dataApiService.GetData(client.AuthenticationContext())
	// Configure only the specific asset type and subtype to be retrieved
	request = request.
		AssetTypeName(eliona.AssetTypeName).
		DataSubtype(string(api.SUBTYPE_INPUT))
	// Get the asset data
	assetData, response, err := request.Execute()
	// Set the result to the empty list by default in any case
	result := []apiserver.CurrencyRate{}
	if err != nil {
		log.Warn("GetExamples", "%s - %s", "failed to retrieve data", err)
		return apiserver.Response(response.StatusCode, result), errors.New("GetExamples method failed to retrieve data")
	}

	for _, data := range assetData {
		// Get the data for every asset data item in response
		d := data.GetData()
		timestamp := int64(d["timestamp_milli"].(float64))
		// Create an array of the retrieved currency rates in the REST model
		result = append(result, apiserver.CurrencyRate{
			AssetID: data.AssetId,
			// Specific time just for better understanding by a user
			TimestampMilli: timestamp,
			Timestamp:      time.UnixMilli(timestamp).Format(time.RFC3339),
			Currency:       d["currency"].(string),
			Value:          d["value"].(float64),
			Description:    d["description"].(string),
		})
	}
	// Sort the data in result based on the timestamp_milli in ascending order
	sort.Slice(result, func(i, j int) bool {
		return result[i].TimestampMilli < result[i].TimestampMilli
	})

	// Return the response with a status code and data
	return apiserver.Response(response.StatusCode, result), nil
}

// PostExample - Creates an example configuration
func (s *ConfigurationApiService) PostExample(_ context.Context, example apiserver.Example) (apiserver.ImplResponse, error) {
	// Get the data request with authentication context
	request := s.dataApiService.PutData(client.AuthenticationContext())
	// Set the input body and the result
	result := apiserver.CurrencyRate{}
	// Marshal the input data to the model and return in case of error
	data, err := json.Marshal(example.Data)
	if err != nil {
		return apiserver.Response(http.StatusUnprocessableEntity, result), errors.New("error processing input body")
	}
	var dataMap map[string]interface{}
	// Unmarshal the input data from the REST interface model to the map of values and return in case of error
	err = json.Unmarshal(data, &dataMap)
	if err != nil {
		return apiserver.Response(http.StatusUnprocessableEntity, result), errors.New("error processing input body")
	}

	// Validate the asset id to be able to PUT the data (only collected data by the application can be updated)
	if example.Id == nil {
		return apiserver.Response(http.StatusUnprocessableEntity, result), errors.New("error processing asset id")
	}
	id := *example.Id
	assetId := int32(id)
	// Check that asset with id assetId exists
	if ok, err := asset.ExistAsset(assetId); !ok && err != nil {
		return apiserver.Response(http.StatusNotFound, result), errors.New("asset with id not found")
	}

	timestamp := time.Now()
	assetType := eliona.AssetTypeName
	// Create a REST request for PUT
	request = request.Data(api.Data{
		AssetId:       assetId,
		Subtype:       api.SUBTYPE_INPUT,
		Timestamp:     *api.NewNullableTime(&timestamp),
		Data:          dataMap,
		AssetTypeName: *api.NewNullableString(&assetType),
	})

	// Execute the PUT REST request
	_, err = request.Execute()
	if err != nil {
		return apiserver.Response(http.StatusBadRequest, result), errors.New("error processing request")
	}

	// Return the response with updated data in case of success
	return apiserver.Response(http.StatusCreated, apiserver.CurrencyRate{
		AssetID:        assetId,
		Timestamp:      timestamp.Format(time.RFC3339),
		TimestampMilli: example.Data.TimestampMilli,
		Currency:       example.Data.Currency,
		Value:          example.Data.Value,
		Description:    example.Data.Description,
	}), nil
}
