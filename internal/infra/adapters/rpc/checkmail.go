package RpcRepo

import (
	"errors"
	RpcClient "github.com/aerosystems/auth-service/pkg/rpc_client"
)

type CheckmailAdapter struct {
	rpcClient *RpcClient.ReconnectRpcClient
}

func NewCheckmailAdapter(rpcClient *RpcClient.ReconnectRpcClient) *CheckmailAdapter {
	return &CheckmailAdapter{
		rpcClient: rpcClient,
	}
}

type InspectRPCPayload struct {
	Domain   string
	ClientIp string
}

func (ca *CheckmailAdapter) IsTrustEmail(email, clientIp string) (bool, error) {
	var result string
	if err := ca.rpcClient.Call(
		"Server.Inspect",
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
