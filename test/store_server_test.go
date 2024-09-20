//go:build integration

package test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itimofeev/yas3/internal/provider/store"
)

func TestClientConnect(t *testing.T) {
	ctx := context.Background()
	storeClient, err := store.New(store.Config{StoreAddr: "https://localhost:9090"})
	require.NoError(t, err)

	fileName := "someFileName.txt"

	err = storeClient.UploadFile(ctx, fileName, strings.NewReader("hello, there!"))
	require.NoError(t, err)

	resp, err := storeClient.GetFile(ctx, fileName)
	if err != nil {
		panic(err)
	}
	defer resp.Close()

	body := &bytes.Buffer{}
	_, err = io.Copy(body, resp)
	if err != nil {
		panic(err)
	}

	log.Printf("Body length: %d bytes \n", body.Len())
	log.Printf("Response body %s \n", body.Bytes())

	available, err := storeClient.GetAvailableSpace(ctx)
	require.NoError(t, err)
	fmt.Println(available)
}
