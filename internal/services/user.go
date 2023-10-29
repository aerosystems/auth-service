package services

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/helpers"
	"github.com/aerosystems/auth-service/internal/models"
	RPCServices "github.com/aerosystems/auth-service/internal/rpc_services"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
	"strings"
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
		log.Printf("could not check email in blacklist: %s", err)
	}
	// getting user by email via RPC
	user, _ := us.userRPC.GetUserByEmail(email)
	// if user with this email already exists
	if user != nil {
		if user.IsActive {
			return errors.New("user with this email already exists")
		} else {
			// updating password for inactive user
			if err := us.userRPC.ResetPassword(user.UserId, passwordHash); err != nil {
				return fmt.Errorf("could not update password for inactive user: %s", err.Error())
			}
			code, _ := us.codeRepo.GetLastIsActiveCode(user.UserId, "registration")
			if code == nil {
				// generating confirmation code
				codeObj := NewCode(user.UserId, "registration", "")
				if err := us.codeRepo.Create(codeObj); err != nil {
					return errors.New("could not gen new code")
				}
			} else {
				// extend expiration code and return previous active code
				if err := us.codeRepo.ExtendExpiration(code); err != nil {
					return fmt.Errorf("could not extend expiration code: %s", err.Error())
				}
			}
			// sending confirmation code via RPC
			if err := us.mailRPC.SendEmail(email, "Confirm your email🗯", fmt.Sprintf("Your confirmation code is %s", code.Code)); err != nil {
				return fmt.Errorf("could not send email: %s", err.Error())
			}
			return nil
		}
	}
	// creating new user via RPC
	userId, err := us.userRPC.CreateUser(email, passwordHash)
	if err != nil {
		return fmt.Errorf("could not create new user: %s", err.Error())
	}
	// generating confirmation code
	codeObj := NewCode(userId, "registration", "")
	if err := us.codeRepo.Create(codeObj); err != nil {
		return errors.New("could not gen new code")
	}
	// sending confirmation code via RPC
	if err := us.mailRPC.SendEmail(email, "Confirm your email🗯", fmt.Sprintf("Your confirmation code is %s", codeObj.Code)); err != nil {
		return fmt.Errorf("could not send email: %s", err.Error())
	}
	return nil
}

func (us *UserServiceImpl) Confirm(code *models.Code) error {
	switch code.Action {
	case "registration":
		// activate user via RPC
		if err := us.userRPC.ActivateUser(code.UserId); err != nil {
			return fmt.Errorf("could not activate user: %s", err.Error())
		}
		code.IsUsed = true
		if err := us.codeRepo.Update(code); err != nil {
			return errors.New("could not confirm registration")
		}
		// create default project via RPC
		if err := us.projectRPC.CreateDefaultProject(code.UserId); err != nil {
			return fmt.Errorf("could not create default project: %s", err.Error())
		}
		// create default subscription via RPC
		if err := us.subscriptionRPC.CreateFreeTrial(code.UserId); err != nil {
			return fmt.Errorf("could not create default subscription: %s", err.Error())
		}
	case "reset_password":
		// activate user via RPC
		if err := us.userRPC.ActivateUser(code.UserId); err != nil {
			return fmt.Errorf("could not activate user: %s", err.Error())
		}
		// reset password via RPC
		if err := us.userRPC.ResetPassword(code.UserId, code.Data); err != nil {
			return fmt.Errorf("could not reset password: %s", err.Error())
		}
		code.IsUsed = true
		if err := us.codeRepo.Update(code); err != nil {
			return fmt.Errorf("could not confirm reset password: %s", err.Error())
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
		codeObj := NewCode(user.UserId, "reset_password", passwordHash)
		if err := us.codeRepo.Create(codeObj); err != nil {
			return errors.New("could not gen new code")
		}
	}
	// extend expiration code and return previous active code
	code.Data = passwordHash
	if err := us.codeRepo.ExtendExpiration(code); err != nil {
		return errors.New("could not extend expiration code")
	}
	// sending confirmation code via RPC
	if err := us.mailRPC.SendEmail(email, "Reset your password🗯", fmt.Sprintf("Your confirmation code is %s", code.Code)); err != nil {
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

func normalizeEmail(data string) string {
	addr := strings.ToLower(data)

	arrAddr := strings.Split(addr, "@")
	username := arrAddr[0]
	domain := arrAddr[1]

	googleDomains := strings.Split(os.Getenv("GOOGLEMAIL_DOMAINS"), ",")

	//checking Google mail aliases
	if helpers.Contains(googleDomains, domain) {
		//removing all dots from username mail
		username = strings.ReplaceAll(username, ".", "")
		//removing all characters after +
		if strings.Contains(username, "+") {
			res := strings.Split(username, "+")
			username = res[0]
		}
		addr = username + "@gmail.com"
	}

	return addr
}
