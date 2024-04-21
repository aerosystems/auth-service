package usecases

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/helpers"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

type AuthUsecase struct {
	codeRepo         CodeRepository
	userRepo         UserRepository
	checkmailAdapter CheckmailAdapter
	mailAdapter      MailAdapter
	customerAdapter  CustomerAdapter
	codeExpMinutes   time.Duration
}

func NewAuthUsecase(codeRepo CodeRepository, userRepo UserRepository, checkmailAdapter CheckmailAdapter, mailAdapter MailAdapter, customerAdapter CustomerAdapter, codeExpMinutes int) *AuthUsecase {
	return &AuthUsecase{
		codeRepo:         codeRepo,
		userRepo:         userRepo,
		checkmailAdapter: checkmailAdapter,
		mailAdapter:      mailAdapter,
		customerAdapter:  customerAdapter,
		codeExpMinutes:   time.Duration(codeExpMinutes) * time.Minute,
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

func NewCode(user models.User, Action models.KindCode, expireAt time.Time, Data string) *models.Code {
	return &models.Code{
		Code:     genCode(),
		User:     user,
		Action:   Action,
		Data:     Data,
		ExpireAt: expireAt,
		IsUsed:   false,
	}
}

func (as AuthUsecase) GetUserByUuid(uuidStr string) (*models.User, error) {
	uuid, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}
	user, err := as.userRepo.GetByUuid(uuid)
	if err != nil {
		return nil, errors.New("could not get user")
	}
	return user, nil
}

func (as AuthUsecase) RegisterCustomer(email, password, clientIp string) error {
	// hashing password
	passwordHash, _ := as.hashPassword(password)
	// checking email in blacklist via RPC
	if _, err := as.checkmailAdapter.IsTrustEmail(email, clientIp); err != nil {
		log.Printf("could not check email in blacklist: %s", err)
	}
	// normalizing email
	email = normalizeEmail(email)
	// getting user by email in local repository
	user, _ := as.userRepo.GetByEmail(email)
	// if user with this email already exists
	if user != nil {
		if user.IsActive {
			return errors.New("user with this email already exists")
		} else {
			// updating password for inactive user
			user.PasswordHash = passwordHash
			if err := as.userRepo.Update(user); err != nil {
				return fmt.Errorf("could not update password for inactive user: %s", err.Error())
			}
			code, _ := as.codeRepo.GetLastIsActiveCode(user.Id, "registration")
			if code == nil {
				// generating confirmation code
				expTime := time.Now().Add(as.codeExpMinutes)
				codeObj := NewCode(*user, models.RegistrationCode, expTime, "")
				if err := as.codeRepo.Create(codeObj); err != nil {
					return errors.New("could not gen new code")
				}
			} else {
				// extend expiration code and return previous active code
				code.ExpireAt = time.Now().Add(as.codeExpMinutes)
				if err := as.codeRepo.Update(code); err != nil {
					return fmt.Errorf("could not extend expiration code: %s", err.Error())
				}
			}
			// sending confirmation code via RPC
			if err := as.mailAdapter.SendEmail(email, "Confirm your email🗯", fmt.Sprintf("Your confirmation code is %s", code.Code)); err != nil {
				return fmt.Errorf("could not send email: %s", err.Error())
			}
			return nil
		}
	}
	// creating new user in local repository
	newUser := NewUser(email, passwordHash)
	newUser.Role = models.CustomerRole
	if err := as.userRepo.Create(newUser); err != nil {
		return fmt.Errorf("could not create new user: %s", err.Error())
	}
	// generating confirmation code
	expTime := time.Now().Add(as.codeExpMinutes)
	newCode := NewCode(*newUser, models.RegistrationCode, expTime, "")
	if err := as.codeRepo.Create(newCode); err != nil {
		return errors.New("could not gen new code")
	}
	// sending confirmation code via RPC
	if err := as.mailAdapter.SendEmail(email, "Confirm your email🗯", fmt.Sprintf("Your confirmation code is %s", newCode.Code)); err != nil {
		return fmt.Errorf("could not send email: %s", err.Error())
	}
	return nil
}

func (as AuthUsecase) Confirm(code *models.Code) error {
	switch code.Action {
	case models.RegistrationCode:
		uuid, err := as.customerAdapter.CreateCustomer()
		if err != nil {
			return fmt.Errorf("could not activate user: %s", err.Error())
		}
		code.IsUsed = true
		code.User.Uuid = uuid
		code.User.IsActive = true
		if err := as.codeRepo.UpdateWithAssociations(code); err != nil {
			return errors.New("could not confirm registration")
		}
	case models.ResetPasswordCode:
		if !code.User.IsActive {
			code.User.IsActive = true
			uuid, err := as.customerAdapter.CreateCustomer()
			if err != nil {
				return fmt.Errorf("could not activate user: %s", err.Error())
			}
			code.User.Uuid = uuid
		}
		code.IsUsed = true
		code.User.PasswordHash = code.Data
		if err := as.codeRepo.UpdateWithAssociations(code); err != nil {
			return fmt.Errorf("could not confirm reset password: %s", err.Error())
		}
	}
	return nil
}

func (as AuthUsecase) ResetPassword(email, password string) error {
	// hashing password
	passwordHash, _ := as.hashPassword(password)
	// normalizing email
	email = normalizeEmail(email)
	// getting user by email in local repository
	user, err := as.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("could not get user")
	}
	code, err := as.codeRepo.GetLastIsActiveCode(user.Id, "reset_password")
	if err != nil {
		return errors.New("could not get last active code")
	}
	if code == nil {
		expTime := time.Now().Add(as.codeExpMinutes)
		code = NewCode(*user, models.ResetPasswordCode, expTime, passwordHash)
		if err := as.codeRepo.Create(code); err != nil {
			return errors.New("could not gen new code")
		}
	}
	// extend expiration code and return previous active code
	code.Data = passwordHash
	code.ExpireAt = time.Now().Add(as.codeExpMinutes)
	if err := as.codeRepo.Update(code); err != nil {
		return errors.New("could not extend expiration code")
	}
	// sending confirmation code via RPC
	if err := as.mailAdapter.SendEmail(email, "Reset your password🗯", fmt.Sprintf("Your confirmation code is %s", code.Code)); err != nil {
		return errors.New("could not send email")
	}
	return nil
}

func (as AuthUsecase) CheckPassword(user *models.User, password string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return false, errors.New("invalid credentials")
	}
	return true, nil
}

func (as AuthUsecase) GetActiveUserByEmail(email string) (*models.User, error) {
	// normalizing email
	email = normalizeEmail(email)
	user, err := as.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("could not get user")
	}
	if user.IsActive == false {
		return nil, errors.New("user is not active")
	}
	return user, nil
}

func (as AuthUsecase) hashPassword(password string) (string, error) {
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

func (as AuthUsecase) GetCode(code string) (*models.Code, error) {
	codeObj, err := as.codeRepo.GetByCode(code)
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
