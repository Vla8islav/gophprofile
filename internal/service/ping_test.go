package service

// honestly, just to increase test coverage
import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Vla8islav/gophprofile/internal/mocks"
)

func TestgophprofileService_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockgophprofileRepository(ctrl)
	svc := gophprofileService{repository: repo}

	repo.EXPECT().Ping(gomock.Any()).Return(nil)
	require.NoError(t, svc.Ping(context.Background()))

	repo.EXPECT().Ping(gomock.Any()).Return(errors.New("db down"))
	require.Error(t, svc.Ping(context.Background()))
}
