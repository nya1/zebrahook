// Code generated by goa v3.7.2, DO NOT EDIT.
//
// HTTP request path constructors for the Zebrahook service.
//
// Command:
// $ goa gen zebrahook/design

package server

import (
	"fmt"
)

// SubmitNewEventsZebrahookPath returns the URL path to the Zebrahook service submitNewEvents HTTP endpoint.
func SubmitNewEventsZebrahookPath() string {
	return "/v1/webhook/events"
}

// RegisterZebrahookPath returns the URL path to the Zebrahook service register HTTP endpoint.
func RegisterZebrahookPath() string {
	return "/v1/webhook/endpoints"
}

// UpdateZebrahookPath returns the URL path to the Zebrahook service update HTTP endpoint.
func UpdateZebrahookPath(id string) string {
	return fmt.Sprintf("/v1/webhook/endpoints/%v", id)
}

// ListWebhookEndpointZebrahookPath returns the URL path to the Zebrahook service listWebhookEndpoint HTTP endpoint.
func ListWebhookEndpointZebrahookPath() string {
	return "/v1/webhook/endpoints/"
}

// GetWebhookEndpointByIDZebrahookPath returns the URL path to the Zebrahook service getWebhookEndpointById HTTP endpoint.
func GetWebhookEndpointByIDZebrahookPath(id string) string {
	return fmt.Sprintf("/v1/webhook/endpoints/%v", id)
}
