/*
 * App template API
 *
 * API to access and configure the app template
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package apiserver

// CurrencyRate REST interface model for the Bitcoiner application
type CurrencyRate struct {
	AssetID        int32   `json:"asset_id,omitempty"`
	TimestampMilli int64   `json:"timestamp_milli"`
	Currency       string  `json:"currency"`
	Value          float64 `json:"value"`
	Description    string  `json:"description"`
	Timestamp      string  `json:"timestamp,omitempty"`
}

type Example struct {

	// A id identifying the example configuration
	Id *int64 `json:"id,omitempty"`

	// Configuration data for example
	Data CurrencyRate `json:"data,omitempty"`
}

// AssertExampleRequired checks if the required fields are not zero-ed
func AssertExampleRequired(obj Example) error {
	return nil
}

// AssertRecurseExampleRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of Example (e.g. [][]Example), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseExampleRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aExample, ok := obj.(Example)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertExampleRequired(aExample)
	})
}
