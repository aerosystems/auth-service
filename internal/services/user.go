package services

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/models"
	RPCServices "github.com/aerosystems/auth-service/internal/rpc_services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
)

type UserServiceImpl struct {
	userRepo     models.UserRepository
	codeRepo     models.CodeRepository
	checkmailRPC *RPCServices.CheckmailRPC
	mailRPC      *RPCServices.MailRPC
}

func NewUserServiceImpl(userRepo models.UserRepository, codeRepo models.CodeRepository, checkmailRPC *RPCServices.CheckmailRPC, mailRPC *RPCServices.MailRPC) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo:     userRepo,
		codeRepo:     codeRepo,
		checkmailRPC: checkmailRPC,
		mailRPC:      mailRPC,
	}
}

func (us *UserServiceImpl) Register(email, password, clientIp string) error {
	if _, err := us.checkmailRPC.IsTrustEmail(email, clientIp); err != nil {
		log.Println(err)
	}
	user, err := us.userRepo.FindByEmail(email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("could not find user by email")
	}
	if user != nil {
		if user.IsActive {
			return errors.New("user with this email already exists")
		} else {
			// updating password for inactive user
			if err := us.userRepo.ResetPassword(user, password); err != nil {
				return errors.New("could not update password for inactive user")
			}
			code, _ := us.codeRepo.GetLastIsActiveCode(user.ID, "registration")
			if code == nil {
				// generating confirmation code
				if _, err = us.codeRepo.NewCode(*user, "registration", ""); err != nil {
					return errors.New("could not gen new code")
				}
			} else {
				// extend expiration code and return previous active code
				if err = us.codeRepo.ExtendExpiration(code); err != nil {
					return errors.New("could not extend expiration code")
				}
			}
			// sending confirmation code via RPC
			if err := us.mailRPC.SendMail(user.Email, "Confirm your email🗯", fmt.Sprintf("Your confirmation code is %s", code.Code)); err != nil {
				return errors.New("could not send email")
			}
			return nil
		}
	}
	// hashing password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return errors.New("could not hash password")
	}
	// creating new inactive user
	newUser := models.User{
		Email:    email,
		Password: string(hashedPassword),
		Role:     "user",
	}
	if err = us.userRepo.Create(&newUser); err != nil {
		return errors.New("could not create new user")
	}
	// generating confirmation code
	code, err := us.codeRepo.NewCode(newUser, "registration", "")
	if err != nil {
		return errors.New("could not gen new code")
	}
	// sending confirmation code via RPC
	if err := us.mailRPC.SendMail(newUser.Email, "Confirm your email🗯", fmt.Sprintf("Your confirmation code is %s", code.Code)); err != nil {
		return errors.New("could not send email")
	}
	return nil
}
