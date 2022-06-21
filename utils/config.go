package utils

import (
	"fmt"
	"zebrahook/constants"

	"github.com/spf13/viper"
)

type PollingIntervalConfig struct {
	Min float64
	Max float64
}

type WorkerConfig struct {
	ParallelJobs        uint
	PollingIntervalSecs PollingIntervalConfig
}

// helper function to init viper
func LoadConfig() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	viper.SetDefault("logger.output.json", false)
	viper.SetDefault("logger.output.type", "console")

	viper.SetDefault("database.type", "postgres")

	viper.SetDefault("worker.pollingIntervalSecs.min", 0.5)
	viper.SetDefault("worker.pollingIntervalSecs.max", 2)

	viper.SetDefault("worker.dispatcher.pollingIntervalSecs.min", 0.5)
	viper.SetDefault("worker.dispatcher.pollingIntervalSecs.max", 2)

	viper.SetDefault("worker.eventMapping.pollingIntervalSecs.min", 0.5)
	viper.SetDefault("worker.eventMapping.pollingIntervalSecs.max", 2)

	// how many parallel jobs to process per worker
	viper.SetDefault("worker.dispatcher.parallelJobs", 3)
	viper.SetDefault("worker.eventMapping.parallelJobs", 1)

	// backoff and max attempt
	viper.SetDefault("backoffStrategy.type", "exponential")
	viper.SetDefault("backoffStrategy.maxAttempts", 3)
	viper.SetDefault("backoffStrategy.baseSecs", 60)

	// webhook request options
	viper.SetDefault("webhookRequest.timeoutSecs", 30)
	// TODO by default set to `Zebrahook/<current version> (+https://github...)`
	viper.SetDefault("webhookRequest.userAgent", "Zebrahook")
	viper.SetDefault("webhookRequest.signatureHeaderName", "Zebrahook-Signature")

	err := viper.ReadInConfig()

	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w ", err))
	}

	// sanity checks
	if viper.GetUint("worker.dispatcher.parallelJobs") <= 0 {
		panic(fmt.Errorf("expected configuration %s to have a value above 0", "worker.dispatcher.parallelJobs"))
	}
	if viper.GetUint("worker.eventMapping.parallelJobs") <= 0 {
		panic(fmt.Errorf("expected configuration %s to have a value above 0", "worker.eventMapping.parallelJobs"))
	}

	if !viper.IsSet("encryptionKey") || len(viper.GetString("encryptionKey")) <= 8 {
		panic(fmt.Errorf("expected configuration %s to have a length above 8", "encryptionKey"))
	}

	if viper.GetString("webhookRequest.signatureHeaderName") == "" {
		panic(fmt.Errorf("expected configuration %s to be set", "webhookRequest.signatureHeaderName"))
	}
}

func getPollingIntervalConfig(workerName *string) PollingIntervalConfig {
	globalPollingMin := viper.GetFloat64("worker.pollingIntervalSecs.min")
	globalPollingMax := viper.GetFloat64("worker.pollingIntervalSecs.max")

	if workerName != nil {
		// dispatcher
		if viper.IsSet(fmt.Sprintf("worker.%s.pollingIntervalSecs", constants.WorkerDispatcher)) && *workerName == constants.WorkerDispatcher {
			return PollingIntervalConfig{
				Min: viper.GetFloat64(fmt.Sprintf("worker.%s.pollingIntervalSecs.min", constants.WorkerDispatcher)),
				Max: viper.GetFloat64(fmt.Sprintf("worker.%s.pollingIntervalSecs.max", constants.WorkerDispatcher)),
			}
		}

		// event mapping
		if viper.IsSet(fmt.Sprintf("worker.%s.pollingIntervalSecs", constants.WorkerEventMapping)) && *workerName == constants.WorkerEventMapping {
			return PollingIntervalConfig{
				Min: viper.GetFloat64(fmt.Sprintf("worker.%s.pollingIntervalSecs.min", constants.WorkerEventMapping)),
				Max: viper.GetFloat64(fmt.Sprintf("worker.%s.pollingIntervalSecs.max", constants.WorkerEventMapping)),
			}
		}
	}

	// fallback to global config
	return PollingIntervalConfig{
		Min: globalPollingMin,
		Max: globalPollingMax,
	}
}

func GetWorkerConfig(workerName string) WorkerConfig {
	parallelJobs := viper.GetUint(fmt.Sprintf("worker.%s.parallelJobs", workerName))

	pollingIntervalSeconds := getPollingIntervalConfig(&workerName)

	return WorkerConfig{
		ParallelJobs:        parallelJobs,
		PollingIntervalSecs: pollingIntervalSeconds,
	}
}
