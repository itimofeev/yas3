package entity

import (
	"context"
	"io"
)

type AvailableSpace struct {
	Total int64
	Used  int64
}

type StoreClient interface {
	GetID() string
	UploadFile(ctx context.Context, fileName string, content io.Reader) error
	GetFile(ctx context.Context, fileName string) (io.ReadCloser, error)
	GetAvailableSpace(ctx context.Context) (AvailableSpace, error)
}
