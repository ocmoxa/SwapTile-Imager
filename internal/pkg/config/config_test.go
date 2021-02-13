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

func TestLoad_Default(t *testing.T) {
	const expEnvironment = "development"

	cfg, err := config.Load(config.UseEnv)
	test.AssertErrNil(t, err)

	if expEnvironment != cfg.Environment {
		t.Fatal("exp", expEnvironment, "got", cfg.Environment)
	}
}

func TestLoad_Environment(t *testing.T) {
	const envVar = "SWAPTILE_ENVIRONMENT"
	const expValue = "test"

	test.AssertErrNil(t, os.Setenv(envVar, expValue))
	defer func() { test.AssertErrNil(t, os.Unsetenv(envVar)) }()

	cfg, err := config.Load(config.UseEnv)
	test.AssertErrNil(t, err)

	if expValue != cfg.Environment {
		t.Fatal("exp", expValue, "got", cfg.Environment)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load(uuid.New().String())

	pathErr := &os.PathError{}
	if !errors.As(err, &pathErr) {
		t.Fatal(err)
	}
}
