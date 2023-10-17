package usecase

import "github.com/aerosystems/auth-service/internal/models"

type UserService interface {
	Register() error
}

type UserServiceImpl struct {
	userRepo models.UserRepository
	codeRepo models.CodeRepository
}

func NewUserServiceImpl(userRepo models.UserRepository, codeRepo models.CodeRepository) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo: userRepo,
		codeRepo: codeRepo,
	}
}

func (us *UserServiceImpl) Register() error {

	return nil
}
