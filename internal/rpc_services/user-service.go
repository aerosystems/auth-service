package RPCServices

import (
	"net/rpc"
)

type UserService interface {
	CreateUser(email, password string) (userId uint, err error)
	GetUserByEmail(email string) (user *UserRPCPayload, err error)
	ResetPassword(userId uint, password string) error
	ActivateUser(userId uint) error
}

type UserRPC struct {
	rpcClient *rpc.Client
}

func NewUserRPC(rpcClient *rpc.Client) *UserRPC {
	return &UserRPC{
		rpcClient: rpcClient,
	}
}

type CreateUserRPCPayload struct {
	Email    string
	Password string
}

type ResetPasswordRPCPayload struct {
	UserId   uint
	Password string
}

type UserRPCPayload struct {
	UserId   uint
	IsActive bool
}

func (u *UserRPC) CreateUser(email, password string) (userId uint, err error) {
	err = u.rpcClient.Call("UserService.CreateUser", CreateUserRPCPayload{email, password}, userId)
	return
}

func (u *UserRPC) GetUserByEmail(email string) (user *UserRPCPayload, err error) {
	err = u.rpcClient.Call("UserService.GetUserByEmail", email, user)
	return
}

func (u *UserRPC) ResetPassword(userId uint, password string) error {
	err := u.rpcClient.Call("UserService.ResetPassword", ResetPasswordRPCPayload{userId, password}, nil)
	return err
}

func (u *UserRPC) ActivateUser(userId uint) error {
	err := u.rpcClient.Call("UserService.ActivateUser", userId, nil)
	return err
}
