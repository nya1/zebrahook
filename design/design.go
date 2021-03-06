package design

import . "goa.design/goa/v3/dsl"

var ApiKeyOrJWTviaToken = JWTSecurity("jwt", func() {
	Description("Provide a JWT token or an API key")
})

// API describes the global properties of the API server.
var _ = API("Zebrahook", func() {
	Title("Zebrahook API")
	Description("Zebrahook API allows to delegate the entire webhook stack.")

	Security(ApiKeyOrJWTviaToken)

	HTTP(func() {
		Path("/v1")
		Consumes("application/json") // Media types supported by the API
		Produces("application/json") // Media types generated by the API
	})
})

var NewWebhookEndpointRegistered = ResultType("application/json", func() {
	Description("New webhook endpointregistered response")
	TypeName("ClientManagement")
})

var EventRequest = Type("EventRequest", func() {
	Attribute("event_type", String, "Event type of the `event_content`", func() {
		Example("merchant-93842.order.shipped")
	})

	Attribute("event_content", MapOf(String, Any), "event content that will be dispatched (any json object)", func() {
		Example(map[string]interface{}{
			"sku": "002432800",
			"customer": map[string]interface{}{
				"country": "NL",
				"address": "Lorem Ipsum 123",
			},
		})
	})

	Attribute("priority", Int, "Optional priority for this event, an higher number will make this event delivered before other ones", func() {
		Example(1000)
	})

	Required("event_type", "event_content")
})

var NewWebhookEndpointRequest = Type("NewWebhookEndpointRequest", func() {
	Attribute("url", String, "URL of the webhook that will be called on each `enabled_events`", func() {
		Format(FormatURI)
		Example("https://example.com/notifications")
	})

	Attribute("enabled_events", ArrayOf(String, func() {
		// NoExample
		// must set single item example otherwise a lorem-ipsum example is used instead
		Example("your.event_name")
	}), "Enabled events for this webhook URL, regex supported - use `[\"*\"]` to listen to all events", func() {
		MinLength(1)
		Example([]string{"merchant-93842.order.*", "my.custom.event"})
	})

	Attribute("metadata", MapOf(String, String), "Optionally pass any custom metadata (key->value)", func() {
		// Default(map[string]string{})
		Example(map[string]string{"anyKeyHere": "any value here"})

		Key(func() {
			MinLength(1) // min length of map key
		})
		Elem(func() {
			MinLength(1) // min length of map value
		})
	})

	Required("url", "enabled_events")
})

var WebhookSecret = Type("WebhookSecret", func() {
	Attribute("secret", String, "secret to be used by the webhook to verify the events", func() {
		Example("zhwhsec_EkXBAkjQZLCtTMtTCoaNatyyiNKARe")
	})
})

var WebhookId = Type("WebhookId", func() {
	Attribute("id", String, "identifier of the webhook", func() {
		Example("zhwe_c9ddsgbei1cst46tglh0")
	})
})

var UpdateWebhookEndpointRequest = Type("UpdateWebhookEndpointRequest", func() {
	Attribute("url", String, "URL of the webhook that will be called on each `enabled_events`", func() {
		Format(FormatURI)
		Example("https://example.com/notifications")
	})

	Attribute("disabled", Boolean, "If true this webhook endpoint won't receive any events, set to false to re-enable it")

	Attribute("enabled_events", ArrayOf(String, func() {
		// NoExample
		// must set single item example otherwise a lorem-ipsum example is used instead
		Example("your.event_name")
	}), "Enabled events for this webhook URL, regex supported - use `[\"*\"]` to listen to all events", func() {
		Example([]string{"your.event_name", "custom.event.*"})
	})

	Attribute("metadata", MapOf(String, String), "Optionally pass any custom metadata (key->value)", func() {
		Example(map[string]string{"anyKeyHere": "any value here"})

		Key(func() {
			MinLength(1) // min length of map key
		})
		Elem(func() {
			MinLength(1) // min length of map value
		})
	})

	Extend(WebhookId)

	Required("id")
})

var WebhookIdAndSecret = Type("WebhookIdAndSecret", func() {
	Required("id", "secret")
	Extend(WebhookId)
	Extend(WebhookSecret)
})

var WebhookEndpointWithoutSecret = Type("WebhookEndpointWithoutSecret", func() {
	Attribute("createdAt", Int64, "when this item was created (unix timestamp seconds)", func() {
		Example(1646278413)
	})
	Attribute("updatedAt", Int64, "when this item was last updated (unix timestamp seconds)", func() {
		Example(1646369084)
	})
	Attribute("status", String, "status of current endpoint, enabled means that the webhook endpoint is eligible for receiving webhook events", func() {
		Enum("enabled", "disabled")
	})

	Required("id", "url", "enabled_events", "createdAt", "updatedAt")

	Extend(NewWebhookEndpointRequest)
	Extend(WebhookId)
})

var WebhookEndpoint = Type("WebhookEndpoint", func() {
	Required("id", "secret", "url", "enabled_events", "createdAt", "updatedAt")

	Extend(WebhookEndpointWithoutSecret)
	Extend(WebhookSecret)
})

// Service describes a service
var _ = Service("Zebrahook", func() {
	Description("Exposes API for Zebrahook")
	// Security(ApiKeyOrJWTviaToken)

	HTTP(func() {
		Path("/webhook")
	})

	Method("createApiKey", func() {
		Description("Create a new API key that allows youo to access all the APIs")

		// not exposed via http
		NoSecurity()

		Payload(func() {
			Attribute("description", String, "description for internal use")
		})

		Result(func() {
			Attribute("apiKey", String, "Api Key created (cleartext), keep it in a safe place")

			Required("apiKey")
		})

		// this method should not be exposed via HTTP
	})

	Method("submitNewEvents", func() {
		Description("Submit new events, all events will be asynchronously dispatched to all endpoints that are subscribed to the provided event type (`enabled_events`)")

		Payload(func() {
			Token("token", String)

			Attribute("events", ArrayOf(EventRequest), func() {
				Example([]map[string]interface{}{
					{
						"event_type": "merchant-93842.charge.succeeded",
						"event_data": map[string]interface{}{
							"id":       372853,
							"amount":   8000,
							"currency": "eur",
							"payment_method_details": map[string]interface{}{
								"card": map[string]interface{}{
									"brand": "visa",
								},
							},
						},
					},
					{
						"event_type": "merchant-93842.order.shipped",
						"event_data": map[string]interface{}{
							"order_id": 12643,
							"sku":      "9001-2",
							"type":     "A01",
							"customer": map[string]interface{}{
								"country": "NL",
								"address": "Lorem Ipsum 33",
							},
						},
					},
				})
			})

			Required("token", "events")
		})

		Result(func() {
			Attribute("success", Boolean, func() {
				Example(true)
			})
		})

		HTTP(func() {
			// Requests to the service consist of HTTP GET requests
			// The payload fields are encoded as path parameters
			POST("/events")
			// Responses use a "200 OK" HTTP status
			// The result is encoded in the response body
			Response(StatusOK)
		})
	})

	Method("register", func() {

		Description("Allows to register a new webhook URL with the specified enabled events")
		// Payload describes the method payload
		// Here the payload is an object that consists of two fields
		Payload(func() {
			Token("token", String)

			Extend(NewWebhookEndpointRequest)
			Required("token")
		})

		// Result describes the method result
		Result(WebhookIdAndSecret)

		// HTTP describes the HTTP transport mapping
		HTTP(func() {
			// Requests to the service consist of HTTP GET requests
			// The payload fields are encoded as path parameters
			POST("/endpoints")
			// Responses use a "200 OK" HTTP status
			// The result is encoded in the response body
			Response(StatusOK)
		})
	})

	Method("update", func() {

		Description("Allows to update a webhook created before")
		// Payload describes the method payload
		Payload(func() {
			Token("token", String)

			Extend(UpdateWebhookEndpointRequest)
			Required("token")
		})

		// Result describes the method result
		Result(func() {
			Attribute("success", Boolean)
		})

		// HTTP describes the HTTP transport mapping
		HTTP(func() {
			// Requests to the service consist of HTTP GET requests
			// The payload fields are encoded as path parameters
			PUT("/endpoints/{id}")
			// Responses use a "200 OK" HTTP status
			// The result is encoded in the response body
			Response(StatusOK)
		})
	})

	Method("listWebhookEndpoint", func() {
		Description("Allows to list and query registered webhook")
		// Payload describes the method payload
		// Here the payload is an object that consists of two fields
		Payload(func() {
			Token("token", String)

			Attribute("limit", Int32, "limit how many results to return, use -1 to return all results", func() {
				Default(50)
				Example(50)
			})
			Attribute("offset", UInt32, "pagination, must be used in combination with limit", func() {
				Default(0)
				Example(0)
			})

			Attribute("metadata", MapOf(String, String), "Search by metadata (key->value)", func() {
				// Example(map[string]string{"anyKeyHere": "any value here"})
				Example(map[string]string{"metadata": "valuehere"})
			})

			Attribute("createdAt.gte", UInt64, "filter by createdAt unix (greater than or equal)", func() {
				Example(1646278413)
				Minimum(0)
			})
			Attribute("updatedAt.lt", UInt64, "filter by updatedAt unix (less than)", func() {
				Example(1646369084)
				Minimum(0)
			})

			Required("token")
		})

		// Result describes the method result
		Result(func() {
			// TODO define ResultType https://github.com/goadesign/goa/issues/2934
			Attribute("result", ArrayOf(WebhookEndpointWithoutSecret))

			Required("result")
		})

		// HTTP describes the HTTP transport mapping
		HTTP(func() {
			// query params
			Param("limit")
			Param("offset")

			MapParams("metadata")

			Param("createdAt.gte")
			Param("updatedAt.lt")

			// Requests to the service consist of HTTP GET requests
			// The payload fields are encoded as path parameters
			GET("/endpoints/")
			// Responses use a "200 OK" HTTP status
			// The result is encoded in the response body
			Response(StatusOK)
		})
	})

	Method("getWebhookEndpointById", func() {
		Description("Allows to get info about a registered webhook URL via the identifier")
		// Payload describes the method payload
		// Here the payload is an object that consists of two fields
		Payload(func() {
			Token("token", String)

			Attribute("id", String, "webhook identifier returned in creation", func() {
				Example("zhwe_c9ddsgbei1cst46tglh0")
			})

			Required("token", "id")
		})

		// Result describes the method result
		Result(WebhookEndpoint)

		// HTTP describes the HTTP transport mapping
		HTTP(func() {
			// Requests to the service consist of HTTP GET requests
			// The payload fields are encoded as path parameters
			GET("/endpoints/{id}")
			// Responses use a "200 OK" HTTP status
			// The result is encoded in the response body
			Response(StatusOK)
		})
	})
})
