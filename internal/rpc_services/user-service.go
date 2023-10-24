package RPCServices

import (
	"net/rpc"
)

type UserService interface {
	CreateUser(email, passwordHash string) (userId uint, err error)
	GetUserByEmail(email string) (user *UserRPCPayload, err error)
	ResetPassword(userId uint, passwordHash string) error
	ActivateUser(userId uint) error
	MatchPassword(email, passwordHash string) (err error)
}

type UserRPC struct {
	rpcClient *rpc.Client
}

func NewUserRPC(rpcClient *rpc.Client) *UserRPC {
	return &UserRPC{
		rpcClient: rpcClient,
	}
}

type UserRPCPayload struct {
	UserId       uint
	IsActive     bool
	Role         string
	Email        string
	PasswordHash string
}

func (u *UserRPC) CreateUser(email, passwordHash string) (userId uint, err error) {
	err = u.rpcClient.Call("UserService.CreateUser", UserRPCPayload{Email: email, PasswordHash: passwordHash}, userId)
	return
}

func (u *UserRPC) GetUserByEmail(email string) (user *UserRPCPayload, err error) {
	err = u.rpcClient.Call("UserService.GetUserByEmail", email, user)
	return
}

func (u *UserRPC) ResetPassword(userId uint, passwordHash string) error {
	err := u.rpcClient.Call("UserService.ResetPassword", UserRPCPayload{UserId: userId, PasswordHash: passwordHash}, nil)
	return err
}

func (u *UserRPC) ActivateUser(userId uint) error {
	err := u.rpcClient.Call("UserService.ActivateUser", userId, nil)
	return err
}

func (u *UserRPC) MatchPassword(email, passwordHash string) (err error) {
	err = u.rpcClient.Call("UserService.MatchPassword", UserRPCPayload{Email: email, PasswordHash: passwordHash}, nil)
	return
}
