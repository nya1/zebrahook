package eventMapping

import (
	"database/sql"
	"encoding/json"
	"math/rand"
	"strings"
	"time"
	"zebrahook/constants"
	"zebrahook/database"
	"zebrahook/models"
	"zebrahook/utils"

	gormLogger "gorm.io/gorm/logger"

	"github.com/nya1/pgq"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type workerPgGo struct {
	db     *sql.DB
	gormDb *gorm.DB
	worker *pgq.Worker
	logger zerolog.Logger
}

type EventDeliveryAttempt struct {
	EventDeliveryAttemptId uint
	EventId                uint
	EndpointId             string
}

func buildEventSearchRegex(eventType string) string {
	// split string by dot (separator)
	eventTypeSplit := strings.Split(eventType, ".")

	eventSearchRegex := eventType // add full event initially

	// build regex like string
	for i := 0; i < len(eventTypeSplit); i++ {
		// add portion of event type with wildcard
		// e.g. charge.* charge.payment_failed*
		eventSearchRegex += "|" + strings.Join(eventTypeSplit[0:i+1], ".") + ".*"
	}

	return "(" + eventSearchRegex + ")"
}

type EventMapping struct {
	EventType string
	EventId   uint
}

// job that by the event_type it builds a list of endpoints
// that are subscribed to the same event type, result is
// 1. will create rows: event delivery, (initial) event delivery attempt
// 2. trigger dispatch job to perform actual http request (event delivery attempt)
func (app *workerPgGo) EventToEndpointsMappingJob(data []byte) error {

	var decodedData EventMapping
	json.Unmarshal(data, &decodedData)

	eventType := decodedData.EventType
	eventId := decodedData.EventId

	// create new logger with correlation id for this job
	thisLogger := app.logger.With().Uint("eventId", decodedData.EventId).Logger()

	thisLogger.Info().Msg("started work")

	thisLogger.Debug().Interface("decodedEvent", decodedData).Msg("")

	eventRegex := buildEventSearchRegex(eventType)

	thisLogger.Debug().Str("regex", eventRegex).Msg("regex to use for query")

	var endpointsToCall []models.Endpoint
	app.gormDb.Raw(`SELECT id, url, enabled_event
	FROM (
		SELECT id, url, unnest(enabled_events) enabled_event
		FROM endpoints WHERE status = ?) x
	WHERE enabled_event ~ ? OR enabled_event = '*'
	`, constants.StatusEnabled, eventRegex).Scan(&endpointsToCall)

	thisLogger.Debug().Interface("endpointsToCall", endpointsToCall).Msg("")

	eventsToDelivery := []models.EventDelivery{}
	nextAttempt := uint(time.Now().Unix())
	for _, endpoint := range endpointsToCall {

		thisLogger.Debug().Interface("endpoint", endpoint).Msg("processing endpoint")

		eventsToDelivery = append(eventsToDelivery, models.EventDelivery{
			NextAttemptScheduledAt: &nextAttempt,
			// on creation set the max attempt from the config
			AttemptsRemaining: viper.GetUint("backoffStrategy.maxAttempts"),
			EndpointID:        endpoint.Id,
			EventID:           eventId,
		})
	}

	thisLogger.Debug().Interface("eventsToDelivery", eventsToDelivery).Msg("events to delivery")

	app.gormDb.Create(&eventsToDelivery)

	// create also first attempts records
	attemptsToCreate := []models.EventDeliveryAttempt{}
	for _, eventDelivery := range eventsToDelivery {
		attemptsToCreate = append(attemptsToCreate, models.EventDeliveryAttempt{
			EventDeliveryID: eventDelivery.Id,
		})
	}
	thisLogger.Debug().Interface("attemptsToCreate", attemptsToCreate).Msg("initial attempts list")

	app.gormDb.Create(&attemptsToCreate)

	// trigger jobs
	for i, eventDeliveryAttempt := range attemptsToCreate {
		eventDeliveryAttempt := EventDeliveryAttempt{
			EventDeliveryAttemptId: eventDeliveryAttempt.Id,
			EndpointId:             eventsToDelivery[i].EndpointID,
			EventId:                eventsToDelivery[i].EventID,
		}

		thisLogger.Debug().Interface("eventDeliveryAttemptJob", eventDeliveryAttempt).Msg("triggering dispatcher job")

		encodedData, _ := json.Marshal(eventDeliveryAttempt)
		app.worker.EnqueueJob(constants.QueueWebhookDelivery, encodedData, pgq.RetryWaits([]time.Duration{}))
	}

	thisLogger.Info().Msg("job processed")

	return nil
}

func (app *workerPgGo) RegisterWorker() {
	worker := pgq.NewWorker(app.db, pgq.SetLogger(&app.logger))
	err := worker.RegisterQueue(constants.QueueEventMapping, app.EventToEndpointsMappingJob)

	if err != nil {
		panic(err)
	}

	app.worker = worker
}

func NewWorker() *workerPgGo {

	// init seed
	rand.Seed(time.Now().UnixNano())

	logger := utils.NewLogger("worker-eventMapping")

	// connect to db
	databaseInfo := database.Open()

	gormLoggerWithZerolog := gormLogger.New(
		&logger, // IO.writer
		gormLogger.Config{
			SlowThreshold:             time.Second,             // Slow SQL threshold
			LogLevel:                  utils.GetGormLogLevel(), // Log level
			IgnoreRecordNotFoundError: true,                    // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,                   // Disable color
		},
	)

	gormDb, err := gorm.Open(postgres.Open(databaseInfo.Dsn), &gorm.Config{
		Logger: gormLoggerWithZerolog,
	})
	if err != nil {
		panic(err)
	}

	return &workerPgGo{databaseInfo.OpenedDatabase, gormDb, nil, logger}
}

func Start() {
	utils.LoadConfig()

	worker := NewWorker()

	worker.RegisterWorker()

	workerName := constants.WorkerEventMapping

	utils.RunWorker(
		workerName,
		utils.GenericWorker{
			InternalPgq: worker.worker,
			Logger:      worker.logger,
		},
	)
}
