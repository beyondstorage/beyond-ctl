package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"

	"github.com/google/uuid"

	"go.beyondstorage.io/v5/pkg/randbytes"
	"go.beyondstorage.io/v5/services"
	"go.beyondstorage.io/v5/types"
)

func getMvTestService(s string) string {
	if s != "" {
		s += "/"
	}
	return fmt.Sprintf(os.Getenv("BEYOND_CTL_TEST_SERVICE"), s)
}

func setupMvFile(t *testing.T) (base, path string) {
	store, err := services.NewStoragerFromString(getMvTestService(""))
	if err != nil {
		t.Fatal(err)
	}

	base = "workdir"
	path = uuid.NewString()

	// Limit the content under 1MB.
	size := rand.Intn(1024 * 1024)
	bs := make([]byte, size)
	_, err = io.ReadFull(randbytes.NewRand(), bs)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Write(fmt.Sprintf("%s/%s", base, path),
		bytes.NewReader(bs), int64(size))
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv(
		fmt.Sprintf("BEYOND_CTL_PROFILE_%s", base),
		getMvTestService(base),
	)
	if err != nil {
		t.Fatal(err)
	}

	return base, path
}

func setupMvTarget(t *testing.T) (targetService, targetPath string) {
	targetService = "workdir/"
	targetPath = uuid.NewString()

	err := os.Setenv(
		fmt.Sprintf("BEYOND_CTL_PROFILE_%s", targetService),
		getMvTestService(targetService),
	)
	if err != nil {
		t.Fatal(err)
	}
	return targetService, targetPath
}

func tearDownMv(t *testing.T, base string) {
	store, err := services.NewStoragerFromString(getMvTestService(""))
	if err != nil {
		t.Fatal(err)
	}

	it, err := store.List(base)
	if err != nil {
		t.Fatal(err)
	}

	for {
		o, err := it.Next()
		if err != nil && errors.Is(err, types.IterateDone) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}

		err = store.Delete(o.Path)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = os.Unsetenv(fmt.Sprintf("BEYOND_CTL_PROFILE_%s", base))
	if err != nil {
		t.Fatal(err)
	}
}

func TestMvFile(t *testing.T) {
	if os.Getenv("BEYOND_CTL_INTEGRATION_TEST") != "on" {
		t.Skipf("BEYOND_CTL_INTEGRATION_TEST is not 'on', skipped")
	}

	base, path := setupMvFile(t)
	defer tearDownMv(t, base)

	targetService, targetPath := setupMvTarget(t)
	defer tearDownMv(t, targetService)

	err := app.Run([]string{
		"byctl", "mv",
		fmt.Sprintf("%s:%s", base, path),
		fmt.Sprintf("%s:%s", targetService, targetPath),
	})
	if err != nil {
		t.Error(err)
	}
}
