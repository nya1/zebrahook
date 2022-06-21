// Code generated by goa v3.7.2, DO NOT EDIT.
//
// Zebrahook endpoints
//
// Command:
// $ goa gen zebrahook/design

package zebrahook

import (
	"context"

	goa "goa.design/goa/v3/pkg"
	"goa.design/goa/v3/security"
)

// Endpoints wraps the "Zebrahook" service endpoints.
type Endpoints struct {
	CreateAPIKey           goa.Endpoint
	SubmitNewEvents        goa.Endpoint
	Register               goa.Endpoint
	Update                 goa.Endpoint
	ListWebhookEndpoint    goa.Endpoint
	GetWebhookEndpointByID goa.Endpoint
}

// NewEndpoints wraps the methods of the "Zebrahook" service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	// Casting service to Auther interface
	a := s.(Auther)
	return &Endpoints{
		CreateAPIKey:           NewCreateAPIKeyEndpoint(s),
		SubmitNewEvents:        NewSubmitNewEventsEndpoint(s, a.JWTAuth),
		Register:               NewRegisterEndpoint(s, a.JWTAuth),
		Update:                 NewUpdateEndpoint(s, a.JWTAuth),
		ListWebhookEndpoint:    NewListWebhookEndpointEndpoint(s, a.JWTAuth),
		GetWebhookEndpointByID: NewGetWebhookEndpointByIDEndpoint(s, a.JWTAuth),
	}
}

// Use applies the given middleware to all the "Zebrahook" service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.CreateAPIKey = m(e.CreateAPIKey)
	e.SubmitNewEvents = m(e.SubmitNewEvents)
	e.Register = m(e.Register)
	e.Update = m(e.Update)
	e.ListWebhookEndpoint = m(e.ListWebhookEndpoint)
	e.GetWebhookEndpointByID = m(e.GetWebhookEndpointByID)
}

// NewCreateAPIKeyEndpoint returns an endpoint function that calls the method
// "createApiKey" of service "Zebrahook".
func NewCreateAPIKeyEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*CreateAPIKeyPayload)
		return s.CreateAPIKey(ctx, p)
	}
}

// NewSubmitNewEventsEndpoint returns an endpoint function that calls the
// method "submitNewEvents" of service "Zebrahook".
func NewSubmitNewEventsEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*SubmitNewEventsPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{},
			RequiredScopes: []string{},
		}
		ctx, err = authJWTFn(ctx, p.Token, &sc)
		if err != nil {
			return nil, err
		}
		return s.SubmitNewEvents(ctx, p)
	}
}

// NewRegisterEndpoint returns an endpoint function that calls the method
// "register" of service "Zebrahook".
func NewRegisterEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*RegisterPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{},
			RequiredScopes: []string{},
		}
		ctx, err = authJWTFn(ctx, p.Token, &sc)
		if err != nil {
			return nil, err
		}
		return s.Register(ctx, p)
	}
}

// NewUpdateEndpoint returns an endpoint function that calls the method
// "update" of service "Zebrahook".
func NewUpdateEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*UpdatePayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{},
			RequiredScopes: []string{},
		}
		ctx, err = authJWTFn(ctx, p.Token, &sc)
		if err != nil {
			return nil, err
		}
		return s.Update(ctx, p)
	}
}

// NewListWebhookEndpointEndpoint returns an endpoint function that calls the
// method "listWebhookEndpoint" of service "Zebrahook".
func NewListWebhookEndpointEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*ListWebhookEndpointPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{},
			RequiredScopes: []string{},
		}
		ctx, err = authJWTFn(ctx, p.Token, &sc)
		if err != nil {
			return nil, err
		}
		return s.ListWebhookEndpoint(ctx, p)
	}
}

// NewGetWebhookEndpointByIDEndpoint returns an endpoint function that calls
// the method "getWebhookEndpointById" of service "Zebrahook".
func NewGetWebhookEndpointByIDEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*GetWebhookEndpointByIDPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{},
			RequiredScopes: []string{},
		}
		ctx, err = authJWTFn(ctx, p.Token, &sc)
		if err != nil {
			return nil, err
		}
		return s.GetWebhookEndpointByID(ctx, p)
	}
}
