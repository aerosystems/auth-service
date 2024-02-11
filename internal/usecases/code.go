package usecases

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/models"
	"math/rand"
	"time"
)

type CodeUsecase struct {
	codeRepo CodeRepository
}

func NewCodeUsecase(codeRepo CodeRepository) *CodeUsecase {
	return &CodeUsecase{
		codeRepo: codeRepo,
	}
}

func (cs *CodeUsecase) GetCode(code string) (*models.Code, error) {
	codeObj, err := cs.codeRepo.GetByCode(code)
	if err != nil {
		return nil, errors.New("could not get data by code")
	}
	if codeObj == nil {
		return nil, errors.New("code does not exist")
	}
	if codeObj.ExpireAt.Before(time.Now()) {
		return nil, errors.New("code is expired")
	}
	if codeObj.IsUsed {
		return nil, errors.New("code is already used")
	}
	return codeObj, nil
}

func NewCode(user models.User, Action models.KindCode, Data string) *models.Code {
	return &models.Code{
		Code:   genCode(),
		User:   user,
		Action: Action,
		Data:   Data,
		IsUsed: false,
	}
}

func genCode() string {
	rand.Seed(time.Now().UnixNano())
	var availableNumbers [3]int
	for i := 0; i < 3; i++ {
		availableNumbers[i] = rand.Intn(9)
	}
	var code string
	for i := 0; i < 6; i++ {
		randNum := availableNumbers[rand.Intn(len(availableNumbers))]

		code = fmt.Sprintf("%s%d", code, randNum)
	}
	return code
}
