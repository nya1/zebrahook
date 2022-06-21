// Code generated by goa v3.7.2, DO NOT EDIT.
//
// Zebrahook service
//
// Command:
// $ goa gen zebrahook/design

package zebrahook

import (
	"context"

	"goa.design/goa/v3/security"
)

// Exposes API for Zebrahook
type Service interface {
	// Create a new API key that allows youo to access all the APIs
	CreateAPIKey(context.Context, *CreateAPIKeyPayload) (res *CreateAPIKeyResult, err error)
	// Submit new events, all events will be asynchronously dispatched to all
	// endpoints that are subscribed to the provided event type (`enabled_events`)
	SubmitNewEvents(context.Context, *SubmitNewEventsPayload) (res *SubmitNewEventsResult, err error)
	// Allows to register a new webhook URL with the specified enabled events
	Register(context.Context, *RegisterPayload) (res *WebhookIDAndSecret, err error)
	// Allows to update a webhook created before
	Update(context.Context, *UpdatePayload) (res *UpdateResult, err error)
	// Allows to list and query registered webhook
	ListWebhookEndpoint(context.Context, *ListWebhookEndpointPayload) (res *ListWebhookEndpointResult, err error)
	// Allows to get info about a registered webhook URL via the identifier
	GetWebhookEndpointByID(context.Context, *GetWebhookEndpointByIDPayload) (res *WebhookEndpoint, err error)
}

// Auther defines the authorization functions to be implemented by the service.
type Auther interface {
	// JWTAuth implements the authorization logic for the JWT security scheme.
	JWTAuth(ctx context.Context, token string, schema *security.JWTScheme) (context.Context, error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "Zebrahook"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [6]string{"createApiKey", "submitNewEvents", "register", "update", "listWebhookEndpoint", "getWebhookEndpointById"}

// CreateAPIKeyPayload is the payload type of the Zebrahook service
// createApiKey method.
type CreateAPIKeyPayload struct {
	// description for internal use
	Description *string
}

// CreateAPIKeyResult is the result type of the Zebrahook service createApiKey
// method.
type CreateAPIKeyResult struct {
	// Api Key created (cleartext), keep it in a safe place
	APIKey string
}

type EventRequest struct {
	// Event type of the `event_content`
	EventType string
	// event content that will be dispatched (any json object)
	EventContent map[string]interface{}
	// Optional priority for this event, an higher number will make this event
	// delivered before other ones
	Priority *int
}

// GetWebhookEndpointByIDPayload is the payload type of the Zebrahook service
// getWebhookEndpointById method.
type GetWebhookEndpointByIDPayload struct {
	Token string
	// webhook identifier returned in creation
	ID string
}

// ListWebhookEndpointPayload is the payload type of the Zebrahook service
// listWebhookEndpoint method.
type ListWebhookEndpointPayload struct {
	Token string
	// limit how many results to return, use -1 to return all results
	Limit int32
	// pagination, must be used in combination with limit
	Offset uint32
	// Search by metadata (key->value)
	Metadata map[string]string
	// filter by createdAt unix (greater than or equal)
	CreatedAtGte *uint64
	// filter by updatedAt unix (less than)
	UpdatedAtLt *uint64
}

// ListWebhookEndpointResult is the result type of the Zebrahook service
// listWebhookEndpoint method.
type ListWebhookEndpointResult struct {
	Result []*WebhookEndpointWithoutSecret
}

// RegisterPayload is the payload type of the Zebrahook service register method.
type RegisterPayload struct {
	Token string
	// URL of the webhook that will be called on each `enabled_events`
	URL string
	// Enabled events for this webhook URL, regex supported - use `["*"]` to listen
	// to all events
	EnabledEvents []string
	// Optionally pass any custom metadata (key->value)
	Metadata map[string]string
}

// SubmitNewEventsPayload is the payload type of the Zebrahook service
// submitNewEvents method.
type SubmitNewEventsPayload struct {
	Token  string
	Events []*EventRequest
}

// SubmitNewEventsResult is the result type of the Zebrahook service
// submitNewEvents method.
type SubmitNewEventsResult struct {
	Success *bool
}

// UpdatePayload is the payload type of the Zebrahook service update method.
type UpdatePayload struct {
	Token string
	// URL of the webhook that will be called on each `enabled_events`
	URL *string
	// If true this webhook endpoint won't receive any events, set to false to
	// re-enable it
	Disabled *bool
	// Enabled events for this webhook URL, regex supported - use `["*"]` to listen
	// to all events
	EnabledEvents []string
	// Optionally pass any custom metadata (key->value)
	Metadata map[string]string
	// identifier of the webhook
	ID string
}

// UpdateResult is the result type of the Zebrahook service update method.
type UpdateResult struct {
	Success *bool
}

// WebhookEndpoint is the result type of the Zebrahook service
// getWebhookEndpointById method.
type WebhookEndpoint struct {
	// when this item was created (unix timestamp seconds)
	CreatedAt int64
	// when this item was last updated (unix timestamp seconds)
	UpdatedAt int64
	// status of current endpoint, enabled means that the webhook endpoint is
	// eligible for receiving webhook events
	Status *string
	// URL of the webhook that will be called on each `enabled_events`
	URL string
	// Enabled events for this webhook URL, regex supported - use `["*"]` to listen
	// to all events
	EnabledEvents []string
	// Optionally pass any custom metadata (key->value)
	Metadata map[string]string
	// identifier of the webhook
	ID string
	// secret to be used by the webhook to verify the events
	Secret string
}

type WebhookEndpointWithoutSecret struct {
	// when this item was created (unix timestamp seconds)
	CreatedAt int64
	// when this item was last updated (unix timestamp seconds)
	UpdatedAt int64
	// status of current endpoint, enabled means that the webhook endpoint is
	// eligible for receiving webhook events
	Status *string
	// URL of the webhook that will be called on each `enabled_events`
	URL string
	// Enabled events for this webhook URL, regex supported - use `["*"]` to listen
	// to all events
	EnabledEvents []string
	// Optionally pass any custom metadata (key->value)
	Metadata map[string]string
	// identifier of the webhook
	ID string
}

// WebhookIDAndSecret is the result type of the Zebrahook service register
// method.
type WebhookIDAndSecret struct {
	// identifier of the webhook
	ID string
	// secret to be used by the webhook to verify the events
	Secret string
}
