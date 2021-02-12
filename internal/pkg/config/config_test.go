// +build integration

package config_test

import (
	"errors"
	"os"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/config"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"

	"github.com/google/uuid"
)

func TestLoad(t *testing.T) {
	const testFile = "../../../config.example.jsonc"

	cfg, err := config.Load(testFile)
	test.AssertErrNil(t, err)

	if cfg.S3.AccessKeyID == "" {
		t.Fatal(cfg.S3.AccessKeyID)
	}
}

func TestLoadDefault(t *testing.T) {
	const expEnvironment = "development"

	cfg, err := config.Load(uuid.New().String())

	pathErr := &os.PathError{}
	switch {
	case !errors.As(err, &pathErr):
		t.Fatal(err)
	case expEnvironment != cfg.Environment:
		t.Fatal("exp", expEnvironment, "got", cfg.Environment)
	}
}
