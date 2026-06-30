package usecase

import (
	"context"
	"fmt"

	"github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	"github.com/danisusanto/go-common/log"
)

type UserUpdatedUsecase interface {
	Exec(ctx context.Context, i *event.UserUpdated) error
}

type userUpdatedUsecase struct {
	userRepository repository.UserRepository
	log            log.Log
}

func NewUserUpdatedUsecase(
	userRepository repository.UserRepository,
	log log.Log,
) UserUpdatedUsecase {
	return &userUpdatedUsecase{
		userRepository: userRepository,
		log:            log,
	}
}

func (u *userUpdatedUsecase) Exec(ctx context.Context, i *event.UserUpdated) error {
	fmt.Println(fmt.Printf("inside user updated usecase %v:", i))
	return nil
}
