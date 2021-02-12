package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"

	"github.com/hedhyw/jsoncjson"
)

// Config of the application.
type Config struct {
	Environment string
	S3          `json:"s3"`
}

// S3 storage client config.
type S3 struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

func defaultConfig() Config {
	return Config{
		Environment: "development",
		S3: S3{
			AccessKeyID:     "",
			SecretAccessKey: "",
		},
	}
}

// Load config from file.
func Load(file string) (cfg Config, err error) {
	cfg = defaultConfig()

	f, err := os.Open(file)
	if err != nil {
		return cfg, fmt.Errorf("opening file: %w", err)
	}

	defer func() { imerrors.ErrorPair(err, f.Close()) }()

	r := jsoncjson.NewReader(f)

	err = json.NewDecoder(r).Decode(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("decoding: %w", err)
	}

	return cfg, nil
}
