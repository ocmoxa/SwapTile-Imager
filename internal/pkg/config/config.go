package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"

	"github.com/caarlos0/env/v6"
	"github.com/hedhyw/jsoncjson"
)

// UseEnv tells to loads values from environment variables.
const UseEnv = ""

// Config of the application.
type Config struct {
	Environment string `env:"SWAPTILE_ENVIRONMENT" envDefault:"development"`
	S3          `json:"s3"`
	Redis       `json:"redis"`
}

// S3 storage client config.
type S3 struct {
	AccessKeyID     string `json:"access_key_id" env:"SWAPTILE_S3_ACCESS_KEY_ID" envDefault:"minio_key"`
	SecretAccessKey string `json:"secret_access_key" env:"SWAPTILE_S3_SECRET_ACCESS_KEY" envDefault:"minio_secret"`
	Endpoint        string `json:"endpoint" env:"SWAPTILE_S3_ENDPOINT" envDefault:"localhost:9000"`
	Secure          bool   `json:"secure" env:"SWAPTILE_S3_SECURE" envDefault:"false"`
	Bucket          string `json:"bucket" env:"SWAPTILE_S3_BUCKET" envDefault:"swaptile"`
	Location        string `json:"location" env:"SWAPTILE_S3_LOCATION" envDefault:"us-east-1"`
}

// Redis client config.
type Redis struct {
	Endpoint string `json:"endpoint" env:"SWAPTILE_REDIS_ENDPOINT" envDefault:"redis://localhost:6379"`
}

// Load config from file or environment. The file has the highest
// priority.
func Load(file string) (cfg Config, err error) {
	err = env.Parse(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("parsing environment: %w", err)
	}

	if file == UseEnv {
		return cfg, nil
	}

	f, err := os.Open(file)
	if err != nil {
		return cfg, fmt.Errorf("opening file: %w", err)
	}

	defer func() { err = imerrors.ErrorPair(err, f.Close()) }()

	r := jsoncjson.NewReader(f)

	err = json.NewDecoder(r).Decode(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("decoding file: %w", err)
	}

	return cfg, nil
}
