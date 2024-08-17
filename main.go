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

	"github.com/jamiealquiza/tachymeter"
)

type configuration struct {
	URL                 string
	NumWorkers          int
	IterationsPerWorker int
}

func buildConfiguration() configuration {
	return configuration{
		URL:                 "http://macmini.local/",
		NumWorkers:          64,
		IterationsPerWorker: 250,
	}
}

func makeHTTPCall(
	ctx context.Context,
	httpClient *http.Client,
	url string,
) error {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext error %w", err)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("httpClient.Do error %w", err)
	}

	slog.Debug("got response",
		"statusCode", response.StatusCode,
	)

	defer response.Body.Close()
	io.Copy(io.Discard, response.Body)

	return nil
}

func runWorker(
	workerNumber int,
	t *tachymeter.Tachymeter,
	config configuration,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	slog.Info("begin runWorker",
		"workerNumber", workerNumber,
	)
	httpClient := &http.Client{}

	ctx := context.TODO()

	for i := 0; i < config.IterationsPerWorker; i++ {
		start := time.Now()

		makeHTTPCall(
			ctx,
			httpClient,
			config.URL,
		)

		// Task we're timing added here.
		t.AddTime(time.Since(start))
	}

	slog.Info("end runWorker",
		"workerNumber", workerNumber,
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

	configuration := buildConfiguration()

	slog.Info("begin main",
		"configuration", &configuration,
	)

	t := tachymeter.New(&tachymeter.Config{
		Size: 1000,
	})

	wallTimeStart := time.Now()

	var wg sync.WaitGroup

	for i := 0; i < configuration.NumWorkers; i++ {
		wg.Add(1)
		go runWorker(
			i,
			t,
			configuration,
			&wg,
		)
	}

	wg.Wait()

	t.SetWallTime(time.Since(wallTimeStart))

	slog.Info(
		"end main",
		"metrics", t.Calc(),
	)
}
