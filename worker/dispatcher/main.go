package dispatcher

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
	"zebrahook/constants"
	"zebrahook/cryptopasta"
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
	AttemptCounter         *int
}

func (app *workerPgGo) CallWebhookEndpointJob(data []byte) error {
	var decodedData EventDeliveryAttempt
	json.Unmarshal(data, &decodedData)

	thisLogger := app.logger.With().Uint("eventDeliveryAttemptId", decodedData.EventDeliveryAttemptId).Logger()

	thisLogger.Info().Msg("started work")

	eventDeliveryAttemptId := decodedData.EventDeliveryAttemptId
	eventId := decodedData.EventId
	endpointId := decodedData.EndpointId

	var eventDeliveryAttempt models.EventDeliveryAttempt
	app.gormDb.Find(&eventDeliveryAttempt, eventDeliveryAttemptId)

	if eventDeliveryAttempt.AttemptMadeAt != nil {
		panic("expected AttemptMadeAt to be empty")
	}

	// get event data
	var eventData models.Event
	app.gormDb.Find(&eventData, eventId)

	// get endpoint url
	var endpointToCall models.Endpoint
	app.gormDb.Where("id = ?", endpointId).First(&endpointToCall)

	// sign event and make http request

	//    load webhook secret and decrypt
	encryptKeyStr := viper.GetString("encryptionKey")
	var encryptionKey [32]byte
	copy(encryptionKey[:], encryptKeyStr)

	webhookSecretStr, err := hex.DecodeString(endpointToCall.SecretEncrypted)
	if err != nil {
		panic(err)
	}

	timestamp := time.Now().Unix()

	decryptedWebhookSecret, _ := cryptopasta.Decrypt(webhookSecretStr, &encryptionKey)
	thisLogger.Debug().Msg("successfully decrypted webhook secret")

	h := hmac.New(sha256.New, decryptedWebhookSecret)

	thisLogger.Debug().Str("signPayload", eventData.EventContent.String()).Int64("sign_timestamp", timestamp).Msg("signing payload")
	payloadToSign := strconv.Itoa(int(timestamp)) + "." + eventData.EventContent.String()

	// Write Data to it
	h.Write([]byte(payloadToSign))

	// Get result and encode as hexadecimal string
	signedPayload := hex.EncodeToString(h.Sum(nil))
	thisLogger.Debug().Str("signature", signedPayload).Msg("sign process complete")

	//     http request
	var jsonStr = []byte(eventData.EventContent.String())
	req, _ := http.NewRequest("POST", endpointToCall.Url, bytes.NewBuffer(jsonStr))
	req.Header.Set("User-Agent", viper.GetString("webhookRequest.userAgent"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(viper.GetString("webhookRequest.signatureHeaderName"), "t="+strconv.Itoa(int(timestamp))+",v1="+signedPayload)

	thisLogger.Debug().Interface("requestHeaders", req.Header).Msg("preparing request")

	client := &http.Client{
		Timeout: time.Duration(viper.GetUint("webhookRequest.timeoutSecs")) * time.Second,
	}

	httpRequestStartTime := time.Now()
	resp, err := client.Do(req)
	httpTimeElapsedSeconds := math.Round(time.Since(httpRequestStartTime).Seconds()*100) / 100

	thisLogger.Info().Float64("httpTimeElapsedSeconds", httpTimeElapsedSeconds).Msg("http request sent")

	var body *string
	var httpStatusCodeResponse *int
	var attemptResultStatus string

	if err == nil {
		defer resp.Body.Close()

		thisLogger.Debug().Msg("no error, reading response body")
		// read only first 200 bytes
		byteBody, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 200))
		stringBody := string(byteBody)
		body = &stringBody

		thisLogger.Debug().Str("responseBody", *body).Int("responseHttpStatusCode", resp.StatusCode).Msg("decoded body")

		httpStatusCodeResponse = &resp.StatusCode

		if *httpStatusCodeResponse >= 200 && *httpStatusCodeResponse <= 299 {
			attemptResultStatus = "success"
		} else {
			attemptResultStatus = "error_response"
		}
	} else {
		thisLogger.Warn().Bool("webhookEndpointError", true).Stack().Err(err).Msg("error while sending request")

		// in case of errors during request (e.g. timeout, connection etc.)
		// we don't have access to response body/status code, set nil
		body = nil
		httpStatusCodeResponse = nil

		// set status based on error type
		if os.IsTimeout(err) {
			attemptResultStatus = "error_timeout"
		} else {
			// using error_network as a generic error
			attemptResultStatus = "error_network"
		}

		// TODO save detailed error message for internal use?
	}

	thisLogger.Debug().Str("attemptResultStatus", attemptResultStatus).Msg("computed status")

	app.gormDb.Transaction(func(tx *gorm.DB) error {
		var eventDelivery models.EventDelivery
		tx.Find(&eventDelivery, eventDeliveryAttempt.EventDeliveryID)

		// update event delivery attempt
		tx.Model(&models.EventDeliveryAttempt{Id: eventDeliveryAttemptId}).Update("http_body_response", body).Update("http_status_code", httpStatusCodeResponse).Update("attempt_made_at", timestamp).Update("http_response_time_secs", httpTimeElapsedSeconds).Update("status", attemptResultStatus)

		// update event delivery
		var attemptsRemaining interface{}
		attemptsRemaining = gorm.Expr("attempts_remaining - ?", 1)

		// check if status code is 2xx, if yes reset attempts remaining
		// as we have successfully delivered the event with this attempt
		if resp != nil && resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			attemptsRemaining = 0
			attemptResultStatus = "success"
		} else {
			// reschedule attempt in the future
			// create new event delivery attempt and queue job

			// update event delivery
			nextAttemptCounter := 1
			// if AttemptCounter is present from the prev event sum it
			if decodedData.AttemptCounter != nil && *decodedData.AttemptCounter > 0 {
				nextAttemptCounter += *decodedData.AttemptCounter
			} else {
				nextAttemptCounter += 1
			}

			// circuit breaker on max attempts, update
			// webhook endpoint state and don't trigger new job
			noMoreAttempts := false
			if eventDelivery.AttemptsRemaining-1 <= 0 {
				noMoreAttempts = true
			}

			// calculate backoff
			if noMoreAttempts == true {
				thisLogger.Info().Str("endpointId", decodedData.EndpointId).Msg("reached maximum attempts, disabling endpoint...")
				tx.Model(models.Endpoint{
					Id: decodedData.EndpointId,
				}).Update("status", constants.StatusDisabled)
			} else {
				backoffSeconds := viper.GetUint("backoffStrategy.baseSecs")
				randomJitter := rand.Float64()
				nextRetrySeconds := math.Pow(float64(backoffSeconds), float64(nextAttemptCounter)+randomJitter)

				nextAttemptScheduled := uint(time.Now().Unix() + int64(nextRetrySeconds))

				thisLogger.Info().Float64("nextRetrySeconds", nextRetrySeconds).Uint("nextAttemptScheduled", nextAttemptScheduled).Msg("rescheduling event delivery with exponential backoff")

				// update event delivery with next attempt scheduled at
				tx.Model(models.EventDelivery{
					Id: eventDeliveryAttempt.EventDeliveryID,
				}).Updates(models.EventDelivery{
					NextAttemptScheduledAt: &nextAttemptScheduled,
				})
				thisLogger.Debug().Msg("event delivery update done")

				// create new event delivery attempt
				nextEventDeliveryAttempt := models.EventDeliveryAttempt{
					EventDeliveryID: eventDeliveryAttempt.EventDeliveryID,
				}
				tx.Create(&nextEventDeliveryAttempt)

				thisLogger.Debug().Uint("nextEventDeliveryAttemptId", nextEventDeliveryAttempt.Id).Msg("new event delivery attempt created")

				// trigger new job for the new attempt
				eventData := EventDeliveryAttempt{
					EndpointId:             decodedData.EndpointId,
					EventDeliveryAttemptId: nextEventDeliveryAttempt.Id,
					EventId:                decodedData.EventId,
					AttemptCounter:         &nextAttemptCounter,
				}
				thisLogger.Debug().Interface("eventDataToEnqueue", eventData).Msg("enqueuing new job")
				encodedEventData, _ := json.Marshal(eventData)
				app.worker.EnqueueJob(constants.QueueWebhookDelivery,
					encodedEventData,
					pgq.After(time.Unix(int64(nextAttemptScheduled), 0)),
					pgq.RetryWaits([]time.Duration{}),
				)
			}
		}

		thisLogger.Debug().Interface("attemptsRemainingUpdate", attemptsRemaining).Msg("updating event delivery with attempts")

		tx.Model(&models.EventDelivery{Id: eventDeliveryAttempt.EventDeliveryID}).Update("attempts_counter", gorm.Expr("attempts_counter + ?", 1))

		// update attempts remaining only if we are at least 1
		if eventDelivery.AttemptsRemaining > 0 {
			tx.Model(&models.EventDelivery{Id: eventDeliveryAttempt.EventDeliveryID}).Update("attempts_remaining", attemptsRemaining)
		}

		return nil
	})

	thisLogger.Info().Msg("job processed")

	return nil
}

func (app *workerPgGo) RegisterWorker() {
	worker := pgq.NewWorker(app.db, pgq.SetLogger(&app.logger))
	err := worker.RegisterQueue(constants.QueueWebhookDelivery, app.CallWebhookEndpointJob)

	if err != nil {
		app.logger.Error().Stack().Err(err).Msg("failed to register queue")
		panic(err)
	}

	app.worker = worker
}

// input: event id - event type
// 1. for every webhook endpoints that
// are subscribed to the same event type
// create new event delivery
// 2. trigger jobs to delivery the webhook
func NewWorker() *workerPgGo {

	// init seed
	rand.Seed(time.Now().UnixNano())

	logger := utils.NewLogger("worker-dispatcher")

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

	workerName := constants.WorkerDispatcher

	utils.RunWorker(
		workerName,
		utils.GenericWorker{
			InternalPgq: worker.worker,
			Logger:      worker.logger,
		},
	)
}
