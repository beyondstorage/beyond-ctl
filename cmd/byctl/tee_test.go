package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"

	"github.com/docker/go-units"
	"github.com/google/uuid"

	"go.beyondstorage.io/v5/pkg/randbytes"
	"go.beyondstorage.io/v5/services"
)

func getTeeTestService(s string) string {
	if s != "" {
		s += "/"
	}
	return fmt.Sprintf(os.Getenv("BEYOND_CTL_TEST_SERVICE"), s)
}

func setupTee(t *testing.T) (base, path string) {
	base = uuid.NewString()
	path = base + "/" + uuid.NewString()

	err := os.Setenv(
		fmt.Sprintf("BEYOND_CTL_PROFILE_%s", base),
		getTeeTestService(base),
	)
	if err != nil {
		t.Fatal(err)
	}

	return base, path
}

func tearDownTee(t *testing.T, path string) {
	store, err := services.NewStoragerFromString(os.Getenv("BEYOND_CTL_TEST_SERVICE"))
	if err != nil {
		t.Fatal(err)
	}

	err = store.Delete(path)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Unsetenv(fmt.Sprintf("BEYOND_CTL_PROFILE_%s", path))
	if err != nil {
		t.Fatal(err)
	}
}

func checkResult(t *testing.T, path string) int64 {
	store, err := services.NewStoragerFromString(os.Getenv("BEYOND_CTL_TEST_SERVICE"))
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	n, err := store.Read(path, &buf)
	if err != nil {
		t.Fatal(err)
	}

	return n
}

func TestTee(t *testing.T) {
	if os.Getenv("BEYOND_CTL_INTEGRATION_TEST") != "on" {
		t.Skipf("BEYOND_CTL_INTEGRATION_TEST is not 'on', skipped")
	}

	base, path := setupTee(t)
	defer tearDownTee(t, path)

	// Limit the content under 1MB.
	size := rand.Intn(1024 * 1024)
	app.Reader = io.LimitReader(randbytes.NewRand(), int64(size))

	err := app.Run([]string{
		"byctl", "tee",
		fmt.Sprintf("%s:%s", base, path),
	})
	if err != nil {
		t.Error(err)
	}

	n := checkResult(t, path)
	if n != int64(size) {
		t.Error("tee failed")
	}
}

func TestTeeViaExpectedSize(t *testing.T) {
	if os.Getenv("BEYOND_CTL_INTEGRATION_TEST") != "on" {
		t.Skipf("BEYOND_CTL_INTEGRATION_TEST is not 'on', skipped")
	}

	base, path := setupTee(t)
	defer tearDownTee(t, path)

	// Limit the content under 1MB.
	size := rand.Intn(1024 * 1024)
	app.Reader = io.LimitReader(randbytes.NewRand(), int64(size))

	floatSize := float64(size)
	err := app.Run([]string{
		"byctl", "tee",
		fmt.Sprintf("--expected-size=%s", units.BytesSize(floatSize)),
		fmt.Sprintf("%s:%s", base, path),
	})
	if err != nil {
		t.Error(err)
	}

	n := checkResult(t, path)
	if n != int64(size) {
		t.Error("tee failed")
	}
}
