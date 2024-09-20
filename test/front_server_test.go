//go:build integration

package test

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/itimofeev/yas3/internal/provider/front"
)

func TestFrontServer(t *testing.T) {
	storeClient, err := front.New(front.Config{BasePath: "http://localhost:8080"})
	require.NoError(t, err)
	for range 10 {
		checkFileUpload(t, 100, storeClient)
	}
}

func checkFileUpload(t *testing.T, fileSize int, storeClient *front.Client) {
	ctx := context.Background()

	fileName := uuid.New().String()

	originalString := generateStringOfSize(fileSize)
	err := storeClient.UploadFile(ctx, fileName, []byte(originalString))
	require.NoError(t, err)

	data, err := storeClient.GetFile(ctx, fileName)
	require.NoError(t, err)

	require.Equal(t, originalString, string(data))
}

func generateStringOfSize(size int) string {
	const alfa = `abcdefghijklmnopqrstuvwxyz0123456789`
	b := &bytes.Buffer{}
	for i := range size {
		b.WriteByte(alfa[i%len(alfa)])
	}
	return b.String()
}
