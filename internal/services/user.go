package services

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/models"
	RPCServices "github.com/aerosystems/auth-service/internal/rpc_services"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type UserService interface {
	Register(email, password, clientIp string) error
	Confirm(code *models.Code) error
	ResetPassword(email, password string) error
	MatchPassword(email, password string) (*RPCServices.UserRPCPayload, error)
}

type UserServiceImpl struct {
	codeRepo        models.CodeRepository
	checkmailRPC    *RPCServices.CheckmailRPC
	mailRPC         *RPCServices.MailRPC
	projectRPC      *RPCServices.ProjectRPC
	subscriptionRPC *RPCServices.SubscriptionRPC
	userRPC         *RPCServices.UserRPC
}

func NewUserServiceImpl(codeRepo models.CodeRepository, checkmailRPC *RPCServices.CheckmailRPC, mailRPC *RPCServices.MailRPC, projectRPC *RPCServices.ProjectRPC, subscriptionRPC *RPCServices.SubscriptionRPC, userRPC *RPCServices.UserRPC) *UserServiceImpl {
	return &UserServiceImpl{
		codeRepo:        codeRepo,
		checkmailRPC:    checkmailRPC,
		mailRPC:         mailRPC,
		projectRPC:      projectRPC,
		subscriptionRPC: subscriptionRPC,
		userRPC:         userRPC,
	}
}

func (us *UserServiceImpl) Register(email, password, clientIp string) error {
	// hashing password
	passwordHash, _ := us.hashPassword(password)
	// checking email in blacklist via RPC
	if _, err := us.checkmailRPC.IsTrustEmail(email, clientIp); err != nil {
		log.Println(err)
	}
	// getting user by email via RPC
	user, err := us.userRPC.GetUserByEmail(email)
	if err != nil {
		return errors.New("could not get user")
	}
	// if user with this email already exists
	if user != nil {
		if user.IsActive {
			return errors.New("user with this email already exists")
		} else {
			// updating password for inactive user
			if err := us.userRPC.ResetPassword(user.UserId, passwordHash); err != nil {
				return errors.New("could not update password")
			}
			code, _ := us.codeRepo.GetLastIsActiveCode(user.UserId, "registration")
			if code == nil {
				// generating confirmation code
				if _, err = us.codeRepo.NewCode(user.UserId, "registration", ""); err != nil {
					return errors.New("could not gen new code")
				}
			} else {
				// extend expiration code and return previous active code
				if err = us.codeRepo.ExtendExpiration(code); err != nil {
					return errors.New("could not extend expiration code")
				}
			}
			// sending confirmation code via RPC
			if err := us.mailRPC.SendMail(email, "Confirm your email🗯", fmt.Sprintf("Your confirmation code is %s", code.Code)); err != nil {
				return errors.New("could not send email")
			}
			return nil
		}
	}
	// creating new user via RPC
	userId, err := us.userRPC.CreateUser(email, passwordHash)
	if err != nil {
		return errors.New("could not create new user")
	}
	// generating confirmation code
	code, err := us.codeRepo.NewCode(userId, "registration", "")
	if err != nil {
		return errors.New("could not gen new code")
	}
	// sending confirmation code via RPC
	if err := us.mailRPC.SendMail(email, "Confirm your email🗯", fmt.Sprintf("Your confirmation code is %s", code.Code)); err != nil {
		return errors.New("could not send email")
	}
	return nil
}

func (us *UserServiceImpl) Confirm(code *models.Code) error {
	switch code.Action {
	case "registration":
		// activate user via RPC
		if err := us.userRPC.ActivateUser(code.UserId); err != nil {
			return errors.New("could not activate user")
		}
		code.IsUsed = true
		if err := us.codeRepo.Update(code); err != nil {
			return errors.New("could not confirm registration")
		}
		// create default project via RPC
		if err := us.projectRPC.CreateDefaultProject(code.UserId); err != nil {
			return errors.New("could not create default project")
		}
		// create default subscription via RPC
		if err := us.subscriptionRPC.CreateFreeTrial(code.UserId); err != nil {
			return errors.New("could not create default subscription")
		}
	case "reset_password":
		// activate user via RPC
		if err := us.userRPC.ActivateUser(code.UserId); err != nil {
			return errors.New("could not activate user")
		}
		// reset password via RPC
		if err := us.userRPC.ResetPassword(code.UserId, code.Data); err != nil {
			return errors.New("could not reset password")
		}
		code.IsUsed = true
		err := us.codeRepo.Update(code)
		if err != nil {
			return errors.New("could not confirm reset password")
		}
	}
	return nil
}

func (us *UserServiceImpl) ResetPassword(email, password string) error {
	// hashing password
	passwordHash, _ := us.hashPassword(password)
	// get user by email via RPC
	user, err := us.userRPC.GetUserByEmail(email)
	if err != nil {
		return errors.New("could not get user")
	}
	code, err := us.codeRepo.GetLastIsActiveCode(user.UserId, "reset_password")
	if err != nil {
		return errors.New("could not get last active code")
	}
	if code == nil || code.IsUsed {
		_, err := us.codeRepo.NewCode(user.UserId, "reset_password", passwordHash)
		if err != nil {
			return errors.New("could not gen new code")
		}
	}
	// extend expiration code and return previous active code
	code.Data = passwordHash
	if err := us.codeRepo.ExtendExpiration(code); err != nil {
		return errors.New("could not extend expiration code")
	}
	// sending confirmation code via RPC
	if err := us.mailRPC.SendMail(email, "Reset your password🗯", fmt.Sprintf("Your confirmation code is %s", code.Code)); err != nil {
		return errors.New("could not send email")
	}
	return nil
}

func (us *UserServiceImpl) MatchPassword(email, password string) (*RPCServices.UserRPCPayload, error) {
	// get user by email via RPC
	user, err := us.userRPC.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("could not get user")
	}
	if user.IsActive == false {
		return nil, errors.New("user is not active")
	}
	// match password via RPC
	if err := us.userRPC.MatchPassword(email, password); err != nil {
		return nil, errors.New("password does not match")
	}
	return user, nil
}

func (us *UserServiceImpl) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", errors.New("could not hash password")
	}
	return string(hash), nil
}
