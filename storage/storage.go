package storage

import (
	"context"
	"errors"
)

var ErrNotConfigured = errors.New("storage is not configured")

type Store interface {
	Configured() bool
	ObjectKey(userID, uploadID, filename string) string
	PresignPut(context.Context, string, string) (string, error)
	PresignGet(context.Context, string) (string, error)
	Exists(context.Context, string) error
	Delete(context.Context, string) error
}
