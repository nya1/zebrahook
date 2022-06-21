package utils

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/nya1/pgq"
	"github.com/rs/zerolog"
)

type GenericWorker struct {
	InternalPgq *pgq.Worker
	Logger      zerolog.Logger
}

// run provided worker in parallel (based on config)
func RunWorker(workerName string, worker GenericWorker) {
	workerConfig := GetWorkerConfig(workerName)

	worker.Logger.Debug().Interface("workerConfig", workerConfig).Msg("parallel jobs to execute: " + fmt.Sprint(workerConfig.ParallelJobs))

	for i := uint(0); i < workerConfig.ParallelJobs; i++ {
		go func(n uint) {
			randomPollingNumber := GetRandomFloatRange(workerConfig.PollingIntervalSecs.Min, workerConfig.PollingIntervalSecs.Max)
			randomDuration := time.Duration(randomPollingNumber * float64(time.Second))

			worker.Logger.Debug().Float64("pollingIntervalSeconds", randomDuration.Seconds()).Msg("starting worker number " + fmt.Sprint(n))

			err := worker.InternalPgq.Run(&randomDuration)
			if err != nil {
				worker.Logger.Error().Stack().Err(err).Msg("run error")
			}
		}(i)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	worker.InternalPgq.StopChan <- true
}
