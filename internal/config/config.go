package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Env            string
	GrpcHost       string
	GrpcPort       string
	FlushFrequency time.Duration
}

func Load() (*Config, error) {
	env := os.Getenv("ENV")
	if env == "" {
		return nil, errors.New("env environment variable is required")
	}
	grpcHost := os.Getenv("GRPC_HOST")
	if grpcHost == "" {
		return nil, errors.New("grpc host environment variable is required")
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		return nil, errors.New("grpc port environment variable is required")
	}

	flushFrequencyRaw := os.Getenv("FLUSH_FREQUENCY")
	if flushFrequencyRaw == "" {
		return nil, errors.New("flush frequency environment variable is required")
	}
	flushFrequency, err := time.ParseDuration(flushFrequencyRaw)
	if err != nil {
		return nil, errors.New("flush frequency must be of type time.Duration")
	}

	return &Config{
		Env:            env,
		GrpcHost:       grpcHost,
		GrpcPort:       grpcPort,
		FlushFrequency: flushFrequency,
	}, nil
}
