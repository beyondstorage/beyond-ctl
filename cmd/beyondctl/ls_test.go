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

	"github.com/beyondstorage/go-storage/v4/pkg/randbytes"
	"github.com/beyondstorage/go-storage/v4/services"
	"github.com/beyondstorage/go-storage/v4/types"
)

func getTestService(s string) string {
	if s != "" {
		s += "/"
	}
	return fmt.Sprintf(os.Getenv("BEYOND_CTL_TEST_SERVICE"), s)
}

type object struct {
	path string
	size int64
}

func setupLs(t *testing.T) (base string, obs []object) {
	store, err := services.NewStoragerFromString(getTestService(""))
	if err != nil {
		t.Fatal(err)
	}

	base = uuid.NewString()

	obs = make([]object, 0)

	for i := 0; i < rand.Intn(1024); i++ {
		path := fmt.Sprintf("%s/%s", base, uuid.NewString())

		// Limit the content under 1MB.
		size := rand.Intn(1024 * 1024)
		bs := make([]byte, size)
		_, err := io.ReadFull(randbytes.NewRand(), bs)
		if err != nil {
			t.Fatal(err)
		}

		_, err = store.Write(path, bytes.NewReader(bs), int64(size))
		if err != nil {
			t.Fatal(err)
		}

		obs = append(obs, object{
			path: path,
			size: int64(size),
		})
	}

	err = os.Setenv(
		fmt.Sprintf("BEYOND_CTL_PROFILE_%s", base),
		getTestService(base),
	)
	if err != nil {
		t.Fatal(err)
	}
	return base, obs
}

func tearDownLs(t *testing.T, base string) {
	store, err := services.NewStoragerFromString(os.Getenv("BEYOND_CTL_TEST_SERVICE"))
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

func TestLs(t *testing.T) {
	if os.Getenv("BEYOND_CTL_INTEGRATION_TEST") != "on" {
		t.Skipf("BEYOND_CTL_INTEGRATION_TEST is not 'on', skipped")
	}

	base, _ := setupLs(t)
	defer tearDownLs(t, base)

	err := app.Run([]string{"beyondctl", "ls", "-l",
		fmt.Sprintf("%s:", base)})
	if err != nil {
		t.Error(err)
	}
}
