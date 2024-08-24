package config

import (
	"fmt"
	"log/slog"

	"github.com/BurntSushi/toml"
)

type HttpClientPoolConfiguration struct {
	NumHttpClients int
	UseH2C         bool
}

type Configuration struct {
	URL                         string
	Workers                     int
	IterationsPerWorker         int
	HttpClientPoolConfiguration HttpClientPoolConfiguration
}

func ReadConfiguration(configFile string) (*Configuration, error) {

	logger := slog.Default().With("configFile", configFile)

	logger.Info("begin ReadConfiguration")

	var configuration Configuration
	_, err := toml.DecodeFile(configFile, &configuration)
	if err != nil {
		logger.Error("toml.DecodeFile error",
			"error", err,
		)
		return nil, fmt.Errorf("ReadConfiguration error: %w", err)
	}

	logger.Info("end ReadConfiguration",
		"configuration", &configuration,
	)

	return &configuration, nil
}
