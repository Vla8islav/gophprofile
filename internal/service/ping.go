package service

import "context"

func (m gophprofileService) Ping(ctx context.Context) error {
	return m.repository.Ping(ctx)
}
