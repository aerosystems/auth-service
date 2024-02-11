package rpcRepo

import (
	"errors"
	"fmt"
	RPCClient "github.com/aerosystems/auth-service/pkg/rpc_client"
)

type CheckmailRepo struct {
	rpcClient *RPCClient.ReconnectRPCClient
}

func NewCheckmailRepo(rpcClient *RPCClient.ReconnectRPCClient) *CheckmailRepo {
	return &CheckmailRepo{
		rpcClient: rpcClient,
	}
}

type InspectRPCPayload struct {
	Domain   string
	ClientIp string
}

func (cs *CheckmailRepo) IsTrustEmail(email, clientIp string) (bool, error) {
	var result string
	if err := cs.rpcClient.Call(
		"Server.Inspect",
		InspectRPCPayload{
			Domain:   email,
			ClientIp: clientIp,
		},
		&result); err != nil {
		fmt.Println("could not check email in blacklist: ", err)
		return false, errors.New("email address does not valid")
	}

	if result == "blacklist" {
		return false, errors.New("email address contains in blacklist")
	}

	return true, nil
}
