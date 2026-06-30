package usecase

import (
	"context"
	"errors"
	"strconv"

	"github.com/danisusanto/go-common/hash"
	commonjwt "github.com/danisusanto/go-common/jwt"

	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/log"
	"github.com/golang-jwt/jwt/v5"
)

type LoginUsecase interface {
	Exec(ctx context.Context, i *LoginUsecaseInput) (*LoginUsecaseOutput, error)
}

type loginUsecase struct {
	userRepository        repository.UserRepository
	userSessionRepository repository.UserSessionRepository
	hash                  hash.Hash
	accessTokenJWT        commonjwt.JWT
	refreshTokenJWT       commonjwt.JWT
	log                   log.Log
}

type LoginUsecaseInput struct {
	Email    string
	Password string
}

type LoginUsecaseOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewLoginUsecase(
	userRepository repository.UserRepository,
	userSessionRepository repository.UserSessionRepository,
	hash hash.Hash,
	accessTokenJWT commonjwt.JWT,
	refreshTokenJWT commonjwt.JWT,
	log log.Log,
) LoginUsecase {
	return &loginUsecase{
		userRepository:        userRepository,
		userSessionRepository: userSessionRepository,
		hash:                  hash,
		accessTokenJWT:        accessTokenJWT,
		refreshTokenJWT:       refreshTokenJWT,
		log:                   log,
	}
}

func (u *loginUsecase) Exec(ctx context.Context, i *LoginUsecaseInput) (*LoginUsecaseOutput, error) {
	user, err := u.userRepository.FindByEmail(ctx, i.Email)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if !u.hash.Compare(i.Password, user.Password) {
		return nil, exception.Throw(exception.UNAUTHORIZED, "invalid credentials")
	}

	if !user.IsUserActive() {
		return nil, exception.Throw(exception.FORBIDDEN, "account is inactive")
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
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate access token jwt")
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

	return &LoginUsecaseOutput{
		AccessToken:  accessToken.Token,
		RefreshToken: refreshToken.Token,
	}, nil
}
