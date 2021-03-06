// Code generated by goa v3.7.2, DO NOT EDIT.
//
// Zebrahook client
//
// Command:
// $ goa gen zebrahook/design

package zebrahook

import (
	"context"

	goa "goa.design/goa/v3/pkg"
)

// Client is the "Zebrahook" service client.
type Client struct {
	CreateAPIKeyEndpoint           goa.Endpoint
	SubmitNewEventsEndpoint        goa.Endpoint
	RegisterEndpoint               goa.Endpoint
	UpdateEndpoint                 goa.Endpoint
	ListWebhookEndpointEndpoint    goa.Endpoint
	GetWebhookEndpointByIDEndpoint goa.Endpoint
}

// NewClient initializes a "Zebrahook" service client given the endpoints.
func NewClient(createAPIKey, submitNewEvents, register, update, listWebhookEndpoint, getWebhookEndpointByID goa.Endpoint) *Client {
	return &Client{
		CreateAPIKeyEndpoint:           createAPIKey,
		SubmitNewEventsEndpoint:        submitNewEvents,
		RegisterEndpoint:               register,
		UpdateEndpoint:                 update,
		ListWebhookEndpointEndpoint:    listWebhookEndpoint,
		GetWebhookEndpointByIDEndpoint: getWebhookEndpointByID,
	}
}

// CreateAPIKey calls the "createApiKey" endpoint of the "Zebrahook" service.
func (c *Client) CreateAPIKey(ctx context.Context, p *CreateAPIKeyPayload) (res *CreateAPIKeyResult, err error) {
	var ires interface{}
	ires, err = c.CreateAPIKeyEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*CreateAPIKeyResult), nil
}

// SubmitNewEvents calls the "submitNewEvents" endpoint of the "Zebrahook"
// service.
func (c *Client) SubmitNewEvents(ctx context.Context, p *SubmitNewEventsPayload) (res *SubmitNewEventsResult, err error) {
	var ires interface{}
	ires, err = c.SubmitNewEventsEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*SubmitNewEventsResult), nil
}

// Register calls the "register" endpoint of the "Zebrahook" service.
func (c *Client) Register(ctx context.Context, p *RegisterPayload) (res *WebhookIDAndSecret, err error) {
	var ires interface{}
	ires, err = c.RegisterEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*WebhookIDAndSecret), nil
}

// Update calls the "update" endpoint of the "Zebrahook" service.
func (c *Client) Update(ctx context.Context, p *UpdatePayload) (res *UpdateResult, err error) {
	var ires interface{}
	ires, err = c.UpdateEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*UpdateResult), nil
}

// ListWebhookEndpoint calls the "listWebhookEndpoint" endpoint of the
// "Zebrahook" service.
func (c *Client) ListWebhookEndpoint(ctx context.Context, p *ListWebhookEndpointPayload) (res *ListWebhookEndpointResult, err error) {
	var ires interface{}
	ires, err = c.ListWebhookEndpointEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*ListWebhookEndpointResult), nil
}

// GetWebhookEndpointByID calls the "getWebhookEndpointById" endpoint of the
// "Zebrahook" service.
func (c *Client) GetWebhookEndpointByID(ctx context.Context, p *GetWebhookEndpointByIDPayload) (res *WebhookEndpoint, err error) {
	var ires interface{}
	ires, err = c.GetWebhookEndpointByIDEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*WebhookEndpoint), nil
}
