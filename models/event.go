package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Original event informations
type Event struct {
	gorm.Model
	Id           uint           `gorm:"primaryKey;not null;<-:create"`
	EventType    string         `gorm:"not null;<-:create"`
	EventContent datatypes.JSON `gorm:"type:json;null;<-:create"`

	Priority int `gorm:"not null;default:0"`

	// TODO metadata, content type

	UpdatedAt int64 `gorm:"autoUpdateTime;not null"` // remove / disable
	CreatedAt int64 `gorm:"autoCreateTime;not null"`
}

// Event delivery, what event to delivery whom
type EventDelivery struct {
	gorm.Model
	Id uint `gorm:"primaryKey;not null;<-:create"`

	// when the next attempt is scheduled (timestamp unix seconds)
	NextAttemptScheduledAt *uint `gorm:"default:null"`

	// how many attempts we have done so far
	AttemptsCounter uint `gorm:"not null;default:0"`

	// how many attempts are remaining, initially set based on current configuration
	// and decremented
	AttemptsRemaining uint `gorm:"not null"`

	// reference to Endpoint
	EndpointID string
	Endpoint   Endpoint

	// reference to Event table
	EventID uint `gorm:"index;not null"`
	Event   Event

	// reference to event delivery attempts
	Attempts []EventDeliveryAttempt

	// unix timestamp (seconds)
	UpdatedAt int64 `gorm:"autoUpdateTime;not null"`
	CreatedAt int64 `gorm:"autoCreateTime;not null"`
}

// data about each attempt
// this will be a write only table
type EventDeliveryAttempt struct {
	gorm.Model
	Id uint `gorm:"primaryKey;not null;<-:create"`

	// outcome for this attempt
	// pending ->
	//			 success (http response status code 2xx)
	//			 error_timeout (http request in timeout error)
	//			 error_response (http response status code != 2xx)
	//			 error_network (generic error, unable to make a connection to the endpoint)
	//
	Status string `gorm:"not null;default:'pending'"`

	// when we have sent the request
	AttemptMadeAt *int64

	// http status code of the response (if any)
	HttpStatusCode *int

	// response body returned from the endpoint (if any)
	HttpBodyResponse *string

	// response time in seconds
	HttpResponseTimeSecs *float32

	// reference to event delivery
	EventDeliveryID uint
	EventDelivery   EventDelivery

	CreatedAt int64 `gorm:"autoCreateTime;not null"`
	UpdatedAt int64 `gorm:"autoUpdateTime;not null"`
}
