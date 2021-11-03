package main

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"go.beyondstorage.io/v5/pkg/randbytes"
	"go.beyondstorage.io/v5/services"
	"io"
	"math/rand"
	"os"
	"testing"
)

func getTeeTestService(s string) string {
	if s != "" {
		s += "/"
	}
	return fmt.Sprintf(os.Getenv("BEYOND_CTL_TEST_SERVICE"), s)
}

func setupTee(t *testing.T) (path string) {
	store, err := services.NewStoragerFromString(getTeeTestService(""))
	if err != nil {
		t.Fatal(err)
	}

	path = uuid.NewString()

	// Limit the content under 1MB.
	size := rand.Intn(1024 * 1024)
	bs := make([]byte, size)
	_, err = io.ReadFull(randbytes.NewRand(), bs)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Write(path, bytes.NewReader(bs), int64(size))
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv(
		fmt.Sprintf("BEYOND_CTL_PROFILE_%s", path),
		getTestService(path),
	)
	if err != nil {
		t.Fatal(err)
	}

	return path
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

func TestTee(t *testing.T) {
	if os.Getenv("BEYOND_CTL_INTEGRATION_TEST") != "on" {
		t.Skipf("BEYOND_CTL_INTEGRATION_TEST is not 'on', skipped")
	}

	path := setupTee(t)
	defer tearDownTee(t, path)

	err := app.Run([]string{
		"byctl", "tee",
		fmt.Sprintf("%s:%s", path, path),
	})
	if err != nil {
		t.Error(err)
	}
}