package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"

	"github.com/caarlos0/env/v6"
	"github.com/hedhyw/jsoncjson"
)

// UseEnv tells to loads values from environment variables.
const UseEnv = ""

// Config of the application.
type Config struct {
	Environment string `json:"environment" env:"SWAPTILE_ENVIRONMENT" envDefault:"development"`
	LogLevel    string `json:"loglevel" env:"SWAPTILE_LOGLEVEL" envDefault:"debug"`

	S3     `json:"s3"`
	Redis  `json:"redis"`
	Core   `json:"core"`
	Server `json:"Server"`
}

type Server struct {
	Name               string   `json:"name" env:"SWAPTILE_SERVER_NAME" envDefault:"SwapTile/Imager"`
	Address            string   `json:"address" env:"SWAPTILE_SERVER_ADDRESS" envDefault:":8080"`
	ExposeErrors       bool     `json:"expose_errors" env:"SWAPTILE_SERVER_EXPOSE_ERRORS" envDefault:"false"`
	ReadTimeout        Duration `json:"read_timeout" env:"SWAPTILE_SERVER_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout       Duration `json:"write_timeout" env:"SWAPTILE_SERVER_WRITE_TIMEOUT" envDefault:"15s"`
	ShutdownTimeout    Duration `json:"shutdown_timeout" env:"SWAPTILE_SERVER_SHUTDOWN_TIMEOUT" envDefault:"5s"`
	CacheControlMaxAge Duration `json:"cache_control_max_age" env:"SWAPTILE_SERVER_CACHE_CONTROL_MAX_AGE" envDefault:"0"`
}

type Core struct {
	ImageContentTypes   []string           `json:"image_content_types" env:"SWAPTILE_CORE_IMAGE_CONTENT_TYPE" envDefault:"image/jpeg"`
	SupportedImageSizes []imager.ImageSize `json:"supported_image_sizes" env:"SWAPTILE_CORE_SUPPORTED_IMAGE_SIZES" envDefault:"1920x1080,480x360"`
	// MaxImageSize is in bytes.
	MaxImageSize int64 `json:"max_image_size" env:"SWAPTILE_CORE_MAX_IMAGE_SIZE" envDefault:"12582912"`
}

// S3 storage client config.
type S3 struct {
	AccessKeyID     string `json:"access_key_id" env:"SWAPTILE_S3_ACCESS_KEY_ID" envDefault:"minio_key"`
	SecretAccessKey string `json:"secret_access_key" env:"SWAPTILE_S3_SECRET_ACCESS_KEY" envDefault:"minio_secret"`
	Endpoint        string `json:"endpoint" env:"SWAPTILE_S3_ENDPOINT" envDefault:"localhost:9001"`
	Secure          bool   `json:"secure" env:"SWAPTILE_S3_SECURE" envDefault:"false"`
	Bucket          string `json:"bucket" env:"SWAPTILE_S3_BUCKET" envDefault:"swaptile"`
	Location        string `json:"location" env:"SWAPTILE_S3_LOCATION" envDefault:"us-east-1"`
}

// Redis client config.
type Redis struct {
	Endpoint string `json:"endpoint" env:"SWAPTILE_REDIS_ENDPOINT" envDefault:"redis://localhost:6380"`
}

// Load config from the file or the environment. The file has the
// highest priority.
func Load(file string) (cfg Config, err error) {
	err = env.ParseWithFuncs(&cfg, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(Duration(time.Nanosecond)): func(v string) (interface{}, error) {
			d, err := time.ParseDuration(v)
			return Duration(d), err
		},
	})
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

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))

		return nil
	case string:
		duration, err := time.ParseDuration(value)
		if err != nil {
			return err
		}

		*d = Duration(duration)

		return nil
	default:
		return imerrors.Error("invalid duration type")
	}
}
