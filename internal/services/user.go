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
	userRepo        models.UserRepository
	codeRepo        models.CodeRepository
	checkmailRPC    *RPCServices.CheckmailRPC
	mailRPC         *RPCServices.MailRPC
	projectRPC      *RPCServices.ProjectRPC
	subscriptionRPC *RPCServices.SubscriptionRPC
}

func NewUserServiceImpl(userRepo models.UserRepository, codeRepo models.CodeRepository, checkmailRPC *RPCServices.CheckmailRPC, mailRPC *RPCServices.MailRPC, projectRPC *RPCServices.ProjectRPC, subscriptionRPC *RPCServices.SubscriptionRPC) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo:        userRepo,
		codeRepo:        codeRepo,
		checkmailRPC:    checkmailRPC,
		mailRPC:         mailRPC,
		projectRPC:      projectRPC,
		subscriptionRPC: subscriptionRPC,
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

func (us *UserServiceImpl) Confirm(code *models.Code) error {
	switch code.Action {
	case "registration":
		code.User.IsActive = true
		code.IsUsed = true
		if err := us.codeRepo.UpdateWithAssociations(code); err != nil {
			return errors.New("could not confirm registration")
		}
		// create default project via RPC
		if err := us.projectRPC.CreateDefaultProject(code.User.ID); err != nil {
			return errors.New("could not create default project")
		}
		// create default subscription via RPC
		if err := us.subscriptionRPC.CreateFreeTrial(uint(code.User.ID)); err != nil {
			return errors.New("could not create default subscription")
		}
	case "reset_password":
		if !code.User.IsActive {
			code.User.IsActive = true
		}
		code.User.Password = code.Data
		code.IsUsed = true
		err := us.codeRepo.UpdateWithAssociations(code)
		if err != nil {
			return errors.New("could not confirm reset password")
		}
	}
	return nil
}

func (us *UserServiceImpl) ResetPassword(email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return errors.New("could not hash password")
	}
	user, err := us.userRepo.FindByEmail(email)
	if err != nil {
		return errors.New("could not find user by email")
	}
	code, err := us.codeRepo.GetLastIsActiveCode(user.ID, "reset_password")
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("could not get last active code")
	}
	if code == nil || code.IsUsed {
		_, err := us.codeRepo.NewCode(*user, "reset_password", string(hashedPassword))
		if err != nil {
			return errors.New("could not gen new code")
		}
	}
	// extend expiration code and return previous active code
	code.Data = string(hashedPassword)
	if err := us.codeRepo.ExtendExpiration(code); err != nil {
		return errors.New("could not extend expiration code")
	}
	// sending confirmation code via RPC
	if err := us.mailRPC.SendMail(user.Email, "Reset your password🗯", fmt.Sprintf("Your confirmation code is %s", code.Code)); err != nil {
		return errors.New("could not send email")
	}
	return nil
}
