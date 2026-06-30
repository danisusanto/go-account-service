package provider

import (
	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-account-service/internal/domain/notification/email"
	"github.com/danisusanto/go-common/env"
	"github.com/danisusanto/go-common/hash"
	"github.com/danisusanto/go-common/jwt"
	"github.com/danisusanto/go-common/log"
	"go.opentelemetry.io/otel/trace"
)

type Usecases struct {
	Login                    usecase.LoginUsecase
	Register                 usecase.RegisterUsecase
	EmailVerification        usecase.EmailVerificationUsecase
	GetUser                  usecase.GetUserUsecase
	RefreshToken             usecase.RefreshTokenUsecase
	ChangePassword           usecase.ChangePasswordUsecase
	VerificationRequest      usecase.VerificationRequestUsecase
	EmailVerificationRequest usecase.EmailVerificationRequestUsecase
	EmailVerificationNotify  usecase.EmailVerificationNotifyUsecase
	ResetPasswordRequest     usecase.ResetPasswordRequestUsecase
	ResetPasswordNotify      usecase.ResetPasswordNotifyUsecase
	ResetPasswordVerify      usecase.ResetPasswordVerifyUsecase
	ResetPassword            usecase.ResetPasswordUsecase
	UpdateUser               usecase.UpdateUserUsecase
	ChangeEmailRequest       usecase.ChangeEmailRequestUsecase
	ChangeEmailNotify        usecase.ChangeEmailNotifyUsecase
	ChangeEmailVerify        usecase.ChangeEmailVerifyUsecase

	UserCreated             usecase.UserCreatedUsecase
	UserVerificationCreated usecase.UserVerificationCreatedUsecase
	UserEmailVerified       usecase.UserEmailVerifiedUsecase
	UserResetPassword       usecase.UserResetPasswordUsecase
	UserEmailChanged        usecase.UserEmailChangedUsecase
	UserUpdated             usecase.UserUpdatedUsecase
}

func SetupUsecase(
	repository *repositories,
	event event.Event,
	tracer trace.Tracer,
	accessTokenJWT jwt.JWT,
	refreshTokenJWT jwt.JWT,
	env env.Env,
	hash hash.Hash,
	log log.Log,
) *Usecases {
	login := usecase.NewLoginUsecase(
		repository.userRepository,
		repository.userSessionRepository,
		hash,
		accessTokenJWT,
		refreshTokenJWT,
		log,
	)

	register := usecase.NewRegisterUsecase(
		repository.userRepository,
		repository.userSessionRepository,
		event,
		accessTokenJWT,
		refreshTokenJWT,
		hash,
		log,
	)

	verificationRequest := usecase.NewVerificationRequestUsecase(
		repository.userRepository,
		repository.userVerificationRepository,
		event,
		hash,
		log,
	)

	emailVerificationRequest := usecase.NewEmailVerificationRequestUsecase(
		verificationRequest,
		env,
		log,
	)

	verification := usecase.NewVerificationVerifyUsecase(
		repository.userVerificationRepository,
		env,
		hash,
		log,
	)

	verified := usecase.NewVerificationUsedUsecase(
		repository.userVerificationRepository,
		env,
		hash,
		log,
	)

	emailVerification := usecase.NewEmailVerificationUsecase(
		repository.userRepository,
		verification,
		event,
		log,
	)

	getUser := usecase.NewGetUserUsecase(
		repository.userRepository,
		log,
	)

	refreshToken := usecase.NewRefreshTokenUsecase(
		repository.userRepository,
		repository.userSessionRepository,
		accessTokenJWT,
		refreshTokenJWT,
		log,
	)

	changePassword := usecase.NewChangePasswordUsecase(
		repository.userRepository,
		hash,
		log,
	)

	resetPasswordVerificationRequest := usecase.NewResetPasswordRequestUsecase(
		repository.userRepository,
		verificationRequest,
		env,
		log,
	)

	resetPasswordVerify := usecase.NewResetPasswordVerifyUsecase(
		repository.userRepository,
		verification,
		log,
	)

	resetPassword := usecase.NewResetPasswordUsecase(
		repository.userRepository,
		verified,
		event,
		hash,
		log,
	)

	updateUser := usecase.NewUpdateUserUsecase(
		repository.userRepository,
		log,
	)

	changeEmailVerificationRequest := usecase.NewChangeEmailRequestUsecase(
		repository.userRepository,
		verificationRequest,
		env,
		log,
	)

	changeEmailVerify := usecase.NewChangeEmailVerifyUsecase(
		repository.userRepository,
		verification,
		event,
		hash,
		log,
	)

	return &Usecases{
		Login:                    login,
		Register:                 register,
		EmailVerificationRequest: emailVerificationRequest,
		EmailVerification:        emailVerification,
		GetUser:                  getUser,
		RefreshToken:             refreshToken,
		ChangePassword:           changePassword,
		ResetPasswordRequest:     resetPasswordVerificationRequest,
		ResetPasswordVerify:      resetPasswordVerify,
		ResetPassword:            resetPassword,
		UpdateUser:               updateUser,
		ChangeEmailRequest:       changeEmailVerificationRequest,
		ChangeEmailVerify:        changeEmailVerify,
	}
}

func SetupEventUsecase(
	repository *repositories,
	event event.Event,
	email email.Email,
	env env.Env,
	hash hash.Hash,
	log log.Log,
) *Usecases {
	verificationRequest := usecase.NewVerificationRequestUsecase(
		repository.userRepository,
		repository.userVerificationRepository,
		event,
		hash,
		log,
	)

	emailVerificationRequest := usecase.NewEmailVerificationRequestUsecase(
		verificationRequest,
		env,
		log,
	)

	emailVerificationNotify := usecase.NewEmailVerificationNotifyUsecase(
		email,
		env,
		log,
	)

	resetPasswordVerificationNotify := usecase.NewResetPasswordNotifyUsecase(
		email,
		env,
		log,
	)

	changeEmailVerificationNotify := usecase.NewChangeEmailNotifyUsecase(
		email,
		env,
		log,
	)

	userCreated := usecase.NewUserCreatedUsecase(
		emailVerificationRequest,
		log,
	)

	userVerificationCreated := usecase.NewUserVerificationCreatedUsecase(
		repository.userRepository,
		emailVerificationNotify,
		resetPasswordVerificationNotify,
		changeEmailVerificationNotify,
		log,
	)

	userUpdated := usecase.NewUserUpdatedUsecase(
		repository.userRepository,
		log,
	)

	userEmailVerified := usecase.NewUserEmailVerifiedUsecase(
		repository.userRepository,
		email,
		env,
		hash,
		log,
	)

	userResetPassword := usecase.NewUserResetPasswordUsecase(
		repository.userRepository,
		email,
		env,
		hash,
		log,
	)

	userEmailChanged := usecase.NewUserEmailChangedUsecase(
		repository.userRepository,
		email,
		env,
		hash,
		log,
	)

	return &Usecases{
		UserCreated:             userCreated,
		UserVerificationCreated: userVerificationCreated,
		UserEmailVerified:       userEmailVerified,
		UserResetPassword:       userResetPassword,
		UserEmailChanged:        userEmailChanged,
		UserUpdated:             userUpdated,
	}
}
