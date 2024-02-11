package usecases

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/helpers"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
	"strings"
)

type UserUsecase struct {
	codeRepo      CodeRepository
	userRepo      UserRepository
	checkmailRepo CheckmailRepo
	mailRepo      MailRepo
	customerRepo  CustomerRepo
}

func NewUserUsecase(codeRepo CodeRepository, userRepo UserRepository, checkmailRepo CheckmailRepo, mailRepo MailRepo, customerRepo CustomerRepo) *UserUsecase {
	return &UserUsecase{
		codeRepo:      codeRepo,
		userRepo:      userRepo,
		checkmailRepo: checkmailRepo,
		mailRepo:      mailRepo,
		customerRepo:  customerRepo,
	}
}

func NewUser(Email, PasswordHash string) *models.User {
	user := models.User{
		Email:        normalizeEmail(Email),
		Uuid:         uuid.New(),
		PasswordHash: PasswordHash,
		IsActive:     false,
	}
	return &user
}

func (us *UserUsecase) GetUserByUuid(uuidStr string) (*models.User, error) {
	uuid, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}
	user, err := us.userRepo.GetByUuid(uuid)
	if err != nil {
		return nil, errors.New("could not get user")
	}
	return user, nil
}

func (us *UserUsecase) RegisterCustomer(email, password, clientIp string) error {
	// hashing password
	passwordHash, _ := us.hashPassword(password)
	// checking email in blacklist via RPC
	if _, err := us.checkmailRepo.IsTrustEmail(email, clientIp); err != nil {
		log.Printf("could not check email in blacklist: %s", err)
	}
	// normalizing email
	email = normalizeEmail(email)
	// getting user by email in local repository
	user, _ := us.userRepo.GetByEmail(email)
	// if user with this email already exists
	if user != nil {
		if user.IsActive {
			return errors.New("user with this email already exists")
		} else {
			// updating password for inactive user
			user.PasswordHash = passwordHash
			if err := us.userRepo.Update(user); err != nil {
				return fmt.Errorf("could not update password for inactive user: %s", err.Error())
			}
			code, _ := us.codeRepo.GetLastIsActiveCode(user.Id, "registration")
			if code == nil {
				// generating confirmation code
				codeObj := NewCode(*user, models.Registration, "")
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
			if err := us.mailRepo.SendEmail(email, "Confirm your email🗯", fmt.Sprintf("Your confirmation code is %s", code.Code)); err != nil {
				return fmt.Errorf("could not send email: %s", err.Error())
			}
			return nil
		}
	}
	// creating new user in local repository
	newUser := NewUser(email, passwordHash)
	newUser.Role = models.CustomerRole
	if err := us.userRepo.Create(newUser); err != nil {
		return fmt.Errorf("could not create new user: %s", err.Error())
	}
	// generating confirmation code
	newCode := NewCode(*newUser, models.Registration, "")
	if err := us.codeRepo.Create(newCode); err != nil {
		return errors.New("could not gen new code")
	}
	// sending confirmation code via RPC
	if err := us.mailRepo.SendEmail(email, "Confirm your email🗯", fmt.Sprintf("Your confirmation code is %s", newCode.Code)); err != nil {
		return fmt.Errorf("could not send email: %s", err.Error())
	}
	return nil
}

func (us *UserUsecase) Confirm(code *models.Code) error {
	switch code.Action {
	case models.Registration:
		uuid, err := us.customerRepo.CreateCustomer()
		if err != nil {
			return fmt.Errorf("could not activate user: %s", err.Error())
		}
		code.IsUsed = true
		code.User.Uuid = uuid
		code.User.IsActive = true
		if err := us.codeRepo.UpdateWithAssociations(code); err != nil {
			return errors.New("could not confirm registration")
		}
	case models.ResetPassword:
		if !code.User.IsActive {
			code.User.IsActive = true
			uuid, err := us.customerRepo.CreateCustomer()
			if err != nil {
				return fmt.Errorf("could not activate user: %s", err.Error())
			}
			code.User.Uuid = uuid
		}
		code.IsUsed = true
		code.User.PasswordHash = code.Data
		if err := us.codeRepo.UpdateWithAssociations(code); err != nil {
			return fmt.Errorf("could not confirm reset password: %s", err.Error())
		}
	}
	return nil
}

func (us *UserUsecase) ResetPassword(email, password string) error {
	// hashing password
	passwordHash, _ := us.hashPassword(password)
	// normalizing email
	email = normalizeEmail(email)
	// getting user by email in local repository
	user, err := us.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("could not get user")
	}
	code, err := us.codeRepo.GetLastIsActiveCode(user.Id, "reset_password")
	if err != nil {
		return errors.New("could not get last active code")
	}
	if code == nil {
		code = NewCode(*user, models.ResetPassword, passwordHash)
		if err := us.codeRepo.Create(code); err != nil {
			return errors.New("could not gen new code")
		}
	}
	// extend expiration code and return previous active code
	code.Data = passwordHash
	if err := us.codeRepo.ExtendExpiration(code); err != nil {
		return errors.New("could not extend expiration code")
	}
	// sending confirmation code via RPC
	if err := us.mailRepo.SendEmail(email, "Reset your password🗯", fmt.Sprintf("Your confirmation code is %s", code.Code)); err != nil {
		return errors.New("could not send email")
	}
	return nil
}

func (us *UserUsecase) CheckPassword(user *models.User, password string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return false, errors.New("invalid credentials")
	}
	return true, nil
}

func (us *UserUsecase) GetActiveUserByEmail(email string) (*models.User, error) {
	// normalizing email
	email = normalizeEmail(email)
	user, err := us.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("could not get user")
	}
	if user.IsActive == false {
		return nil, errors.New("user is not active")
	}
	return user, nil
}

func (us *UserUsecase) hashPassword(password string) (string, error) {
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
