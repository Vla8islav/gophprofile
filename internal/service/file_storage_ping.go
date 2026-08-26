package service

import (
	"context"
)

func (m gophprofileService) FileStoragePing(ctx context.Context) error {
	return m.fileStorage.Ping(ctx)
}
