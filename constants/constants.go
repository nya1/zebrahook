package constants

const (
	ZEBRAHOOK_API_KEY_PREFIX = "sk_"

	ZEBRAHOOK_ID_PREFIX = "zh"

	ZEBRAHOOK_ID_WEBHOOK_ENDPOINT_PREFIX = ZEBRAHOOK_ID_PREFIX + "we_"

	ZEBRAHOOK_ID_WEBHOOK_SECRET_PREFIX = ZEBRAHOOK_ID_PREFIX + "whsec_"

	StatusEnabled  string = "enabled"
	StatusDisabled string = "disabled"

	// queue naming
	QueueEventMapping    = "event_mapping"
	QueueWebhookDelivery = "webhook_delivery"

	// worker naming (used for config)
	WorkerEventMapping = "eventMapping"
	WorkerDispatcher   = "dispatcher"
)
