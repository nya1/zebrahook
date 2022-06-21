package zebrahook

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	goa "goa.design/goa/v3/pkg"
	"goa.design/goa/v3/security"
	gormLogger "gorm.io/gorm/logger"

	"zebrahook/constants"
	cryptopasta "zebrahook/cryptopasta"
	"zebrahook/database"
	"zebrahook/models"
	"zebrahook/utils"

	front "zebrahook/gen/zebrahook"

	"github.com/nya1/pgq"
	xid "github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/wagslane/go-rabbitmq"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type EventMapping struct {
	EventType string
	EventId   uint
}

type frontsrvc struct {
	logger    *zerolog.Logger
	db        *gorm.DB
	publisher *rabbitmq.Publisher
}

// NewFront returns the front service implementation.
// note: this is called before http initialization
func NewFront() front.Service {
	utils.LoadConfig()

	// init seed
	rand.Seed(time.Now().UnixNano())

	// connect to db
	databaseInfo := database.Open()

	logger := utils.NewLogger("server")

	gormLoggerWithZerolog := gormLogger.New(
		&logger, // IO.writer
		gormLogger.Config{
			SlowThreshold:             time.Second,             // Slow SQL threshold
			LogLevel:                  utils.GetGormLogLevel(), // Log level
			IgnoreRecordNotFoundError: true,                    // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,                   // Disable color
		},
	)

	db, err := gorm.Open(postgres.Open(databaseInfo.Dsn), &gorm.Config{
		Logger: gormLoggerWithZerolog,
	})
	if err != nil {
		panic("failed to connect to database")
	}

	return &frontsrvc{&logger, db, nil}
}

func (s *frontsrvc) findApiKey(clearTextApiKey string) (matchFound *models.ApiKey, err error) {
	hashedApiKey := utils.HashApiKey(clearTextApiKey)

	var apiKeyFound models.ApiKey
	if result := s.db.First(&apiKeyFound, models.ApiKey{Hash: hashedApiKey}); result.Error != nil {
		return nil, errors.New("api key not found")
	}

	return &apiKeyFound, nil
}

func (s *frontsrvc) CreateAPIKey(ctx context.Context, p *front.CreateAPIKeyPayload) (res *front.CreateAPIKeyResult, err error) {
	s.logger.Debug().Interface("payload", p).Msg("front.CreateApiKey")

	// generate api key
	clearTextApiKey := constants.ZEBRAHOOK_API_KEY_PREFIX + utils.GenerateRandomString(36)

	hashedApiKey := utils.HashApiKey(clearTextApiKey)

	// insert to db
	// create new endpoint instance
	newApiKeyToCreate := &models.ApiKey{
		Hash:        hashedApiKey,
		Description: *p.Description,
		Status:      constants.StatusEnabled,
	}

	s.db.Create(newApiKeyToCreate)

	return &front.CreateAPIKeyResult{
		APIKey: clearTextApiKey,
	}, nil
}

func (s *frontsrvc) SubmitNewEvents(ctx context.Context, p *front.SubmitNewEventsPayload) (res *front.SubmitNewEventsResult, err error) {
	var eventsToInsert []models.Event

	for _, eventReq := range p.Events {
		// default priority
		priority := 0

		if eventReq.Priority != nil {
			priority = *eventReq.Priority
		}
		eventContentJson, _ := json.Marshal(eventReq.EventContent)

		newEvent := models.Event{
			EventType:    eventReq.EventType,
			EventContent: datatypes.JSON(eventContentJson),
			Priority:     priority,
		}
		eventsToInsert = append(eventsToInsert, newEvent)
	}

	// sort by priority
	sort.SliceStable(eventsToInsert, func(i, j int) bool {
		return eventsToInsert[i].Priority > eventsToInsert[j].Priority
	})

	rawDb, _ := s.db.DB()
	worker := pgq.NewWorker(rawDb)

	// create event delivery, first event attempt entry and trigger job
	// TODO pack everything in one tranasaction

	s.db.Create(&eventsToInsert)

	var basicEventMapping []EventMapping
	for _, event := range eventsToInsert {
		eventMapping := EventMapping{
			EventType: event.EventType,
			EventId:   event.Id,
		}
		s.logger.Debug().Interface("eventMapping", eventMapping).Msg("going to enqueue new event")
		basicEventMapping = append(basicEventMapping, eventMapping)
		encodedEventMapping, _ := json.Marshal(eventMapping)

		jobID, _ := worker.EnqueueJob(constants.QueueEventMapping, encodedEventMapping)
		s.logger.Debug().Int("jobId", jobID).Msg("enqueue successfully")
	}

	s.logger.Info().Interface("events", basicEventMapping).Msg(fmt.Sprintf("created %d event(s)", len(eventsToInsert)))

	return &front.SubmitNewEventsResult{
		Success: utils.BoolPointer(true),
	}, nil
}

func (s *frontsrvc) ListWebhookEndpoint(ctx context.Context, p *front.ListWebhookEndpointPayload) (res *front.ListWebhookEndpointResult, err error) {
	s.logger.Debug().Interface("payload", p).Msg("front.ListWebhookEndpoint")

	var webhookEndpointFoundList []models.Endpoint

	query := s.db.Limit(int(p.Limit)).Offset(int(p.Offset))

	// add additional where based on provided input
	// metadata, for each key->value provided add to where query
	// e.g. metadata[key]=value will look for metadata.key == value
	if p.Metadata != nil {
		for metadataKey, metadataValue := range p.Metadata {
			if metadataKey == "" {
				continue
			}
			query = query.Where(datatypes.JSONQuery("metadata").Equals(metadataValue, metadataKey))
		}
	}
	// date related
	if p.CreatedAtGte != nil && uint64(*p.CreatedAtGte) >= uint64(0) {
		query = query.Where("created_at >= ?", uint64(*p.CreatedAtGte))
	}
	if p.UpdatedAtLt != nil && uint64(*p.UpdatedAtLt) >= uint64(0) {
		query = query.Where("updated_at < ?", uint64(*p.UpdatedAtLt))
	}

	// execute query
	result := query.Find(&webhookEndpointFoundList)

	if result.Error != nil {
		err := errors.New("error while querying webhook endpoints")
		return nil, err
	}

	s.logger.Debug().Msg(fmt.Sprintf("results found %d", len(webhookEndpointFoundList)))

	formattedResult := []*front.WebhookEndpointWithoutSecret{}
	for _, element := range webhookEndpointFoundList {
		// json to map
		var metadataMapping map[string]string
		json.Unmarshal([]byte(element.Metadata), &metadataMapping)

		formattedResult = append(formattedResult, &front.WebhookEndpointWithoutSecret{
			ID:            element.Id,
			URL:           element.Url,
			EnabledEvents: element.EnabledEvents,
			Metadata:      metadataMapping,
			Status:        &element.Status,
			CreatedAt:     element.CreatedAt,
			UpdatedAt:     element.UpdatedAt,
		})
	}

	res = &front.ListWebhookEndpointResult{
		Result: formattedResult,
	}

	return res, nil
}

func (s *frontsrvc) GetWebhookEndpointByID(ctx context.Context, p *front.GetWebhookEndpointByIDPayload) (res *front.WebhookEndpoint, err error) {
	s.logger.Debug().Interface("payload", p).Msg("front.getWebhookEndpointById")

	var webhookEndpointFound models.Endpoint
	result := s.db.First(&webhookEndpointFound, "id = ?", p.ID)

	if result.Error != nil {
		err := errors.New("Webhook endpoint identifier " + p.ID + " not found")
		return nil, err
	}

	s.logger.Debug().Interface("webhookEndpointFound", webhookEndpointFound).Msg("")

	// decrypt secret
	encryptKeyStr := viper.GetString("encryptionKey")
	var encryptionKey [32]byte
	copy(encryptionKey[:], encryptKeyStr)

	webhookSecretStr, err := hex.DecodeString(webhookEndpointFound.SecretEncrypted)
	if err != nil {
		panic(err)
	}

	decryptedSecret, _ := cryptopasta.Decrypt(webhookSecretStr, &encryptionKey)

	// json to map
	var metadataMapping map[string]string
	json.Unmarshal([]byte(webhookEndpointFound.Metadata), &metadataMapping)

	res = &front.WebhookEndpoint{
		ID:            webhookEndpointFound.Id,
		Secret:        string(decryptedSecret),
		URL:           webhookEndpointFound.Url,
		EnabledEvents: webhookEndpointFound.EnabledEvents,
		Metadata:      metadataMapping,
		Status:        &webhookEndpointFound.Status,
		CreatedAt:     webhookEndpointFound.CreatedAt,
		UpdatedAt:     webhookEndpointFound.UpdatedAt,
	}
	return res, nil
}

// the "jwt" security scheme.
func (s *frontsrvc) JWTAuth(ctx context.Context, token string, scheme *security.JWTScheme) (context.Context, error) {
	s.logger.Debug().Str("token", token).Interface("scheme", scheme).Msg("called jwt auth")

	apiKeyFound, err := s.findApiKey(token)

	if err != nil {
		return ctx, goa.PermanentError("unauthorized", "invalid token")
	}

	ctx = contextWithAuthInfo(ctx, authInfo{
		userId: "utestinguser_" + fmt.Sprint(apiKeyFound.ID),
	})

	return ctx, nil

	//
	// TBD: add authorization logic.
	//
	// In case of authorization failure this function should return
	// one of the generated error structs, e.g.:
	//
	//    return ctx, myservice.MakeUnauthorizedError("invalid token")
	//
	// Alternatively this function may return an instance of
	// goa.ServiceError with a Name field value that matches one of
	// the design error names, e.g:
	//
	//    return ctx, goa.PermanentError("unauthorized", "invalid token")
	//
	// return ctx, goa.PermanentError("unauthorized", "invalid token")
}

// Allows to register a new webhook URL with the specified enabled events
func (s *frontsrvc) Register(ctx context.Context, p *front.RegisterPayload) (res *front.WebhookIDAndSecret, err error) {
	s.logger.Debug().Interface("payload", p).Msg("front.register")

	authInfo := contextAuthInfo(ctx)

	s.logger.Debug().Interface("authInfo", authInfo).Msg("")

	// generate new webhook identifier
	guid := xid.New()
	newWebhookId := constants.ZEBRAHOOK_ID_WEBHOOK_ENDPOINT_PREFIX + guid.String()

	// generate new webhook secret token, this is returned to the
	// client in clear text
	webhookSecret := constants.ZEBRAHOOK_ID_WEBHOOK_SECRET_PREFIX + utils.GenerateRandomString(30)

	// encrypt secret, this is stored on our side
	encryptKeyStr := viper.GetString("encryptionKey")
	var encryptionKey [32]byte
	copy(encryptionKey[:], encryptKeyStr)
	webhookSecretEncrypted, _ := cryptopasta.Encrypt([]byte(webhookSecret), &encryptionKey)
	webhookSecretEncryptedHex := hex.EncodeToString(webhookSecretEncrypted)

	// create new endpoint instance
	newEndpointToCreate := &models.Endpoint{
		Id:              newWebhookId,
		Url:             p.URL,
		SecretEncrypted: webhookSecretEncryptedHex,
		EnabledEvents:   p.EnabledEvents,
		Status:          constants.StatusEnabled,
	}

	// encode map to json
	if p.Metadata != nil && len(p.Metadata) > 0 {
		jsonMetadata, _ := json.Marshal(p.Metadata)
		newEndpointToCreate.Metadata = datatypes.JSON(jsonMetadata)
	}

	s.db.Create(newEndpointToCreate)

	s.logger.Info().Msg("registered new webhook endpoint " + newWebhookId)

	res = &front.WebhookIDAndSecret{
		ID:     newWebhookId,
		Secret: webhookSecret,
	}

	return res, nil
}

func (s *frontsrvc) Update(ctx context.Context, p *front.UpdatePayload) (res *front.UpdateResult, err error) {
	s.logger.Debug().Interface("payload", p).Msg("front.update")

	authInfo := contextAuthInfo(ctx)

	s.logger.Debug().Interface("authInfo", authInfo).Msg("")

	// add update based on input provided
	endpointContent := models.Endpoint{}

	if p.EnabledEvents != nil && len(p.EnabledEvents) > 0 {
		endpointContent.EnabledEvents = p.EnabledEvents
	}

	if p.Metadata != nil && len(p.Metadata) > 0 {
		jsonMetadata, _ := json.Marshal(p.Metadata)
		endpointContent.Metadata = jsonMetadata
	}

	if p.Disabled != nil {
		var newStatusToUse string
		if *p.Disabled {
			newStatusToUse = constants.StatusDisabled
		} else {
			newStatusToUse = constants.StatusEnabled
		}
		endpointContent.Status = newStatusToUse
	}

	s.db.Model(models.Endpoint{}).Where("id = ?", p.ID).Updates(endpointContent)

	success := true
	res = &front.UpdateResult{
		Success: &success,
	}

	return res, nil
}
