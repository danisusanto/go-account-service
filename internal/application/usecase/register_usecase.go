package usecase

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/danisusanto/go-common/hash"
	commonjwt "github.com/danisusanto/go-common/jwt"

	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/log"
	"github.com/golang-jwt/jwt/v5"
)

type RegisterUsecase interface {
	Exec(ctx context.Context, i *RegisterUsecaseInput) (*RegisterUsecaseOutput, error)
}

type registerUsecase struct {
	userRepository        repository.UserRepository
	userSessionRepository repository.UserSessionRepository
	event                 event.Event
	accessTokenJWT        commonjwt.JWT
	refreshTokenJWT       commonjwt.JWT
	hash                  hash.Hash
	log                   log.Log
}

type RegisterUsecaseInput struct {
	Name        string
	Email       string
	Password    string
	CountryCode string
	PhoneNumber string
}

type RegisterUsecaseOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewRegisterUsecase(
	userRepository repository.UserRepository,
	userSessionRepository repository.UserSessionRepository,
	event event.Event,
	accessTokenJWT commonjwt.JWT,
	refreshTokenJWT commonjwt.JWT,
	hash hash.Hash,
	log log.Log,
) RegisterUsecase {
	return &registerUsecase{
		userRepository:        userRepository,
		userSessionRepository: userSessionRepository,
		event:                 event,
		accessTokenJWT:        accessTokenJWT,
		refreshTokenJWT:       refreshTokenJWT,
		hash:                  hash,
		log:                   log,
	}
}

func (u *registerUsecase) Exec(ctx context.Context, i *RegisterUsecaseInput) (*RegisterUsecaseOutput, error) {
	user, err := u.userRepository.FindByEmail(ctx, i.Email)

	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}
	if user != nil {
		return nil, exception.Throw(exception.CONFLICT, "email already exists")
	}

	password, err := u.hash.Hash(i.Password)
	if err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate password hashed")
	}

	user, err = u.userRepository.Create(ctx, entity.NewUser(i.Name, i.Email, password, i.CountryCode, i.PhoneNumber))
	if err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to create user")
	}

	if err := u.event.Publish(ctx, &event.Payload{
		EventName: event.UserCreatedEvent,
		CreatedAt: time.Now(),
		Data: &event.UserCreated{
			UserID: user.ID,
			Email:  user.Email,
		},
	}); err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish user create event")
	}

	accessToken, err := u.accessTokenJWT.Generate(
		&commonjwt.ClaimsInput{
			UserID: strconv.Itoa(int(user.ID)),
			Name:   user.Name,
			Email:  user.Email,
			Role:   string(user.Role),
		},
	)
	if err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate jwt token")
	}

	refreshToken, err := u.refreshTokenJWT.Generate(
		&commonjwt.ClaimsInput{
			UserID: strconv.Itoa(int(user.ID)),
			RegisteredClaims: jwt.RegisteredClaims{
				ID: accessToken.ID,
			},
		},
	)
	if err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate refresh token jwt")
	}

	_, err = u.userSessionRepository.CreateBulk(
		ctx,
		[]*entity.UserSession{
			entity.NewSession(
				user.ID,
				accessToken.ID,
				entity.AccessToken,
				&accessToken.ExpiresAt.Time,
			),
			entity.NewSession(
				user.ID,
				refreshToken.ID,
				entity.RefreshToken,
				&refreshToken.ExpiresAt.Time,
			),
		},
	)
	if err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to create session cache")
	}

	return &RegisterUsecaseOutput{
		AccessToken:  accessToken.Token,
		RefreshToken: refreshToken.Token,
	}, nil
}
