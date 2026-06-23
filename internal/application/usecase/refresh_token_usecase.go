package usecase

import (
	"context"
	"errors"
	"strconv"

	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	"github.com/dani-susanto/go-common/exception"
	commonjwt "github.com/dani-susanto/go-common/jwt"
	"github.com/dani-susanto/go-common/log"
	"github.com/golang-jwt/jwt/v5"
)

type RefreshTokenUsecase interface {
	Exec(ctx context.Context, token string) (*RefreshTokenOutput, error)
}

type refreshTokenUsecase struct {
	userRepository        repository.UserRepository
	userSessionRepository repository.UserSessionRepository
	accessTokenJWT        commonjwt.JWT
	refreshTokenJWT       commonjwt.JWT
	log                   log.Log
}

type RefreshTokenOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewRefreshTokenUsecase(
	userRepository repository.UserRepository,
	userSessionRepository repository.UserSessionRepository,
	accessTokenJWT commonjwt.JWT,
	refreshTokenJWT commonjwt.JWT,
	log log.Log,
) RefreshTokenUsecase {
	return &refreshTokenUsecase{
		userRepository:        userRepository,
		userSessionRepository: userSessionRepository,
		accessTokenJWT:        accessTokenJWT,
		refreshTokenJWT:       refreshTokenJWT,
		log:                   log,
	}
}

func (u *refreshTokenUsecase) Exec(ctx context.Context, token string) (*RefreshTokenOutput, error) {
	claims, err := u.refreshTokenJWT.Validate(token)
	if err != nil {
		return nil, exception.Throw(exception.BAD_REQUEST, err.Error())
	}

	_, err = u.userSessionRepository.FindByKey(ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, exception.Throw(exception.BAD_REQUEST, "refresh token session expired")
		}
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if err := u.userSessionRepository.DeleteByKey(ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)); err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	userID, err := strconv.Atoi(claims.UserID)
	if err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}
	user, err := u.userRepository.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
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
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate accesst token jwt")
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

	return &RefreshTokenOutput{
		AccessToken:  accessToken.Token,
		RefreshToken: refreshToken.Token,
	}, nil
}
