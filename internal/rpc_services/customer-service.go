package RPCServices

import (
	RPCClient "github.com/aerosystems/auth-service/pkg/rpc_client"
	"github.com/google/uuid"
)

type CustomerService interface {
	CreateCustomer() (uuid.UUID, error)
}

type CustomerRPC struct {
	rpcClient *RPCClient.ReconnectRPCClient
}

func NewCustomerRPC(rpcClient *RPCClient.ReconnectRPCClient) *CustomerRPC {
	return &CustomerRPC{
		rpcClient: rpcClient,
	}
}

type CustomerRPCPayload struct {
	Uuid uuid.UUID
}

func (c *CustomerRPC) CreateCustomer() (uuid.UUID, error) {
	result := CustomerRPCPayload{}
	if err := c.rpcClient.Call("CustomerServer.CreateCustomer", "", &result); err != nil {
		return uuid.UUID{}, err
	}
	return result.Uuid, nil
}
