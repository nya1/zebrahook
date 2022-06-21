package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	zebrahook "zebrahook"
	"zebrahook/constants"
	"zebrahook/database"
	front "zebrahook/gen/zebrahook"
	"zebrahook/models"
	"zebrahook/utils"
	"zebrahook/worker/dispatcher"
	"zebrahook/worker/eventMapping"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	gormLogger "gorm.io/gorm/logger"

	"gorm.io/gorm"
)

func main() {
	utils.LoadConfig()

	// Define command line flags, add any other flag required to configure the
	// service.
	var (
		serverMode        = flag.Bool("server", false, "Start in server mode")
		workerToRun       = flag.String("worker", "", "Start only the provided worker, available: "+constants.WorkerEventMapping+", "+constants.WorkerDispatcher)
		_                 = flag.String("host", "localhost", "Server host (valid values: localhost)")
		_                 = flag.String("log-level", "info", "log level, allowed: "+strings.Join(utils.AllowedLogLevels[:], ","))
		_                 = flag.Bool("log-json", false, "flag to output logs in json format")
		domainF           = flag.String("domain", "", "Server host domain name")
		_                 = flag.String("http-port", "3000", "Server HTTP port")
		secureF           = flag.Bool("secure", false, "Server, use secure scheme (https or grpcs)")
		dbgF              = flag.Bool("http-verbose", false, "Server, log request and response bodies")
		generateNewApiKey = flag.String("new-api-key", "", "Allows to generate a new API key")
		setupDb           = flag.Bool("setup", false, "Setup database with required sql tables")
	)

	flag.Parse()

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	viper.BindPFlag("http-port", pflag.CommandLine.Lookup("http-port"))
	viper.BindPFlag("host", pflag.CommandLine.Lookup("host"))
	viper.BindPFlag("logger.level", pflag.CommandLine.Lookup("log-level"))
	viper.BindPFlag("logger.output.json", pflag.CommandLine.Lookup("log-json"))

	zerologInstance := utils.NewLogger("")

	// check if setup db flag was provided
	// if yes run auto migrate and execute sql queries
	if setupDb != nil && *setupDb {
		databaseInfo := database.Open()

		gormLoggerWithZerolog := gormLogger.New(
			&zerologInstance, // IO.writer
			gormLogger.Config{
				SlowThreshold:             time.Second,             // Slow SQL threshold
				LogLevel:                  utils.GetGormLogLevel(), // Log level
				IgnoreRecordNotFoundError: true,                    // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,                   // Disable color
			},
		)

		zerologInstance.Info().Msg("starting with database setup")
		db, err := gorm.Open(postgres.Open(databaseInfo.Dsn), &gorm.Config{
			Logger: gormLoggerWithZerolog,
		})
		if err != nil {
			panic("failed to connect to database")
		}
		rawDb, err := db.DB()
		if err != nil {
			panic(err)
		}
		defer func() {
			rawDb.Close()
		}()

		// auto migrate tables
		err = db.AutoMigrate(
			&models.Endpoint{},
			&models.ApiKey{},
			&models.Event{},
			&models.EventDelivery{},
			models.EventDeliveryAttempt{},
		)
		if err != nil {
			panic(err)
		}

		// create pgql table
		// ref https://github.com/btubbs/pgq/blob/0a3335913e86a402013ee81a9e45ffbe502bbffe/sql/create_table.sql
		db.Exec(`BEGIN;
		CREATE TABLE IF NOT EXISTS pgq_jobs (
		  id SERIAL PRIMARY KEY,
		  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
		  queue_name TEXT NOT NULL,
		  data BYTEA NOT NULL,
		  run_after TIMESTAMP WITH TIME ZONE NOT NULL,
		  retry_waits TEXT[] NOT NULL,
		  ran_at TIMESTAMP WITH TIME ZONE,
		  error TEXT
		);
		
		-- Add an index for fast fetching of jobs by queue_name, sorted by run_after.  But only
		-- index jobs that haven't been done yet, in case the user is keeping the job history around.
		CREATE INDEX IF NOT EXISTS idx_pgq_jobs_fetch
			ON pgq_jobs (queue_name, run_after)
			WHERE ran_at IS NULL;
		COMMIT;`)

		zerologInstance.Info().Msg("database setup completed")

		return
	}

	// check if we need to start a worker or the server
	if workerToRun != nil && *workerToRun != "" {
		zerologInstance.Info().Msg(fmt.Sprintf("going to start %s worker", *workerToRun))

		if *workerToRun == constants.WorkerEventMapping {
			eventMapping.Start()
		} else if *workerToRun == constants.WorkerDispatcher {
			dispatcher.Start()
		} else {
			panic("invalid worker name provided, expected " + constants.WorkerEventMapping + " or " + constants.WorkerDispatcher)
		}
	} else if *serverMode == true {
		zerologInstance.Info().Msg("going to start server...")
		// Initialize the services.
		var (
			frontSvc front.Service
		)
		{
			frontSvc = zebrahook.NewFront()
		}

		// Wrap the services in endpoints that can be invoked from other services
		// potentially running in different processes.
		var (
			frontEndpoints *front.Endpoints
		)
		{
			frontEndpoints = front.NewEndpoints(frontSvc)
		}

		// Create channel used by both the signal handler and server goroutines
		// to notify the main goroutine when to stop the server.
		errc := make(chan error)

		// Setup interrupt handler. This optional step configures the process so
		// that SIGINT and SIGTERM signals cause the services to stop gracefully.
		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			errc <- fmt.Errorf("%s", <-c)
		}()

		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())

		// check if generate api key command is provided
		if generateNewApiKey != nil && *generateNewApiKey != "" {
			log.Print("generate api key!! ", *generateNewApiKey)
			apiKeyCreationResult, _ := frontSvc.CreateAPIKey(ctx, &front.CreateAPIKeyPayload{Description: generateNewApiKey})

			// output cleartext api key
			log.Print("created new API Key: " + apiKeyCreationResult.APIKey)

			// all done, exit
			cancel()

			wg.Wait()
			return
		}

		// Start the servers and send errors (if any) to the error channel.
		host := viper.GetString("host")
		switch host {
		case "localhost":
			{
				httpPort := viper.GetString("http-port")
				addr := "http://localhost:" + fmt.Sprint(httpPort)
				u, err := url.Parse(addr)
				if err != nil {
					fmt.Fprintf(os.Stderr, "invalid URL %#v: %s\n", addr, err)
					os.Exit(1)
				}
				if *secureF {
					u.Scheme = "https"
				}
				if *domainF != "" {
					u.Host = *domainF
				}

				h, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					fmt.Fprintf(os.Stderr, "invalid URL %#v: %s\n", u.Host, err)
					os.Exit(1)
				}
				u.Host = net.JoinHostPort(h, httpPort)

				handleHTTPServer(ctx, u, frontEndpoints, &wg, errc, &zerologInstance, *dbgF)
			}

		default:
			fmt.Fprintf(os.Stderr, "invalid host argument: %q (	: localhost)\n", host)
			os.Exit(1)
		}

		// Wait for signal.
		zerologInstance.Info().Msg(fmt.Sprintf("exiting (%v)", <-errc))

		// Send cancellation signal to the goroutines.
		cancel()

		wg.Wait()
		zerologInstance.Info().Msg("exited")
	} else {
		zerologInstance.Panic().Msg("Expected --server or --worker <worker type>")
	}
}
