package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aaronriekenberg/go-bench/config"
)

func makeHTTPCall(
	ctx context.Context,
	httpClient *http.Client,
	url string,
) (statusCode int, err error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		err = fmt.Errorf("http.NewRequestWithContext error %w", err)
		return
	}

	response, err := httpClient.Do(request)
	if err != nil {
		err = fmt.Errorf("httpClient.Do error %w", err)
		return
	}

	slog.Debug("got response",
		"statusCode", response.StatusCode,
	)

	defer response.Body.Close()
	io.Copy(io.Discard, response.Body)

	statusCode = response.StatusCode

	return
}

type workerResult struct {
	statusCodeCounts map[int]int
	callsPerSecond   float64
}

func runWorker(
	workerNumber int,
	config config.Configuration,
	wg *sync.WaitGroup,
	workerResultChannel chan<- workerResult,
) {
	defer wg.Done()

	slog.Info("begin runWorker",
		"workerNumber", workerNumber,
	)

	httpClient := &http.Client{}

	ctx := context.TODO()

	statusCodeCounts := make(map[int]int)

	numCalls := 0

	startTime := time.Now()

	for i := 0; i < config.IterationsPerWorker; i++ {

		statusCode, err := makeHTTPCall(
			ctx,
			httpClient,
			config.URL,
		)

		if err != nil {
			slog.Warn("makeHTTPCall error",
				"error", err,
			)
		}
		statusCodeCounts[statusCode]++

		numCalls++
	}

	callDuration := time.Since(startTime)

	callsPerSecond := float64(numCalls) / callDuration.Seconds()

	slog.Info("end runWorker",
		"workerNumber", workerNumber,
		"numCalls", numCalls,
		"callDuration", callDuration.String(),
		"callsPerSecond", callsPerSecond,
	)

	workerResultChannel <- workerResult{
		statusCodeCounts: statusCodeCounts,
		callsPerSecond:   callsPerSecond,
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic in main",
				"error", err,
			)
			os.Exit(1)
		}
	}()

	setupSlog()

	if len(os.Args) != 2 {
		panic("config file required as command line arument")
	}

	configFile := os.Args[1]

	configuration, err := config.ReadConfiguration(configFile)
	if err != nil {
		panic(fmt.Errorf("main: config.ReadConfiguration error: %w", err))
	}

	slog.Info("begin main",
		"configuration", configuration,
	)

	workerResultsChannel := make(chan workerResult, configuration.Workers)

	var wg sync.WaitGroup

	for i := 0; i < configuration.Workers; i++ {
		wg.Add(1)
		go runWorker(
			i,
			*configuration,
			&wg,
			workerResultsChannel,
		)
	}

	wg.Wait()

	close(workerResultsChannel)

	totalCallsPerSecond := float64(0)
	mergedStatusCodeCount := make(map[int]int)
	for workerResult := range workerResultsChannel {
		for statusCode, count := range workerResult.statusCodeCounts {
			mergedStatusCodeCount[statusCode] += count
			totalCallsPerSecond += workerResult.callsPerSecond
		}
	}

	slog.Info(
		"end main",
		"mergedStatusCodeCount", mergedStatusCodeCount,
		"totalCallsPerSecond", totalCallsPerSecond,
	)
}

func setupSlog() {
	level := slog.LevelInfo

	if levelString, ok := os.LookupEnv("LOG_LEVEL"); ok {
		err := level.UnmarshalText([]byte(levelString))
		if err != nil {
			panic(fmt.Errorf("level.UnmarshalText error %w", err))
		}
	}

	slog.SetDefault(
		slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{
					Level: level,
				},
			),
		),
	)

	slog.Info("setupSlog",
		"configuredLevel", level,
	)
}
