package RPCServices

import (
	"errors"
	RPCClient "github.com/aerosystems/auth-service/pkg/rpc_client"
)

type CheckmailService interface {
	IsTrustEmail(email, clientIp string) (bool, error)
}

type CheckmailRPC struct {
	rpcClient *RPCClient.ReconnectRPCClient
}

type InspectRPCPayload struct {
	Domain   string
	ClientIp string
}

func NewCheckmailRPC(rpcClient *RPCClient.ReconnectRPCClient) *CheckmailRPC {
	return &CheckmailRPC{
		rpcClient: rpcClient,
	}
}

func (cs *CheckmailRPC) IsTrustEmail(email, clientIp string) (bool, error) {
	var result string
	if err := cs.rpcClient.Call(
		"CheckmailServer.Inspect",
		InspectRPCPayload{
			Domain:   email,
			ClientIp: clientIp,
		},
		&result); err != nil {
		return false, errors.New("email address does not valid")
	}

	if result == "blacklist" {
		return false, errors.New("email address contains in blacklist")
	}

	return true, nil
}
