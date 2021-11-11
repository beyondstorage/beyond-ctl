package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"go.beyondstorage.io/v5/pkg/randbytes"
	"go.beyondstorage.io/v5/services"
	"go.beyondstorage.io/v5/types"
)

func getSignTestService(s string) string {
	if s != "" {
		s += "/"
	}
	return fmt.Sprintf(os.Getenv("BEYOND_CTL_TEST_SERVICE"), s)
}

func setupSign(t *testing.T) (base, path string) {
	store, err := services.NewStoragerFromString(getSignTestService(""))
	if err != nil {
		t.Fatal(err)
	}

	base = uuid.NewString()
	path = uuid.NewString()

	rand.Seed(time.Now().Unix())
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

func tearDownSign(t *testing.T, base string) {
	store, err := services.NewStoragerFromString(getSignTestService(""))
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

func TestSign(t *testing.T) {
	if os.Getenv("BEYOND_CTL_INTEGRATION_TEST") != "on" {
		t.Skipf("BEYOND_CTL_INTEGRATION_TEST is not 'on', skipped")
	}

	base, path := setupSign(t)
	defer tearDownSign(t, base)

	err := app.Run([]string{
		"byctl", "sign",
		fmt.Sprintf("%s:%s", base, path),
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSignViaExpire(t *testing.T) {
	if os.Getenv("BEYOND_CTL_INTEGRATION_TEST") != "on" {
		t.Skipf("BEYOND_CTL_INTEGRATION_TEST is not 'on', skipped")
	}

	base, path := setupSign(t)
	defer tearDownSign(t, base)

	// Set the expire time to 150 seconds.
	err := app.Run([]string{
		"byctl", "sign",
		fmt.Sprintf("--expire=%d", 150),
		fmt.Sprintf("%s:%s", base, path),
	})
	if err != nil {
		t.Error(err)
	}
}
