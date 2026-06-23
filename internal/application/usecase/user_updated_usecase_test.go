package usecase_test

import (
	"context"
	"testing"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/event"
	mocklog "github.com/dani-susanto/go-account-service/mocks/log"
	mockrepository "github.com/dani-susanto/go-account-service/mocks/repository"
	"github.com/stretchr/testify/assert"
)

func TestUserUpdatedUsecase(t *testing.T) {
	ctx := context.Background()
	input := &event.UserUpdated{UserID: 1}
	userRepository := mockrepository.NewMockUserRepository(t)
	log := mocklog.NewMockLog(t)

	usecase := usecase.NewUserUpdatedUsecase(userRepository, log)

	err := usecase.Exec(ctx, input)

	assert.Nil(t, err)
}
