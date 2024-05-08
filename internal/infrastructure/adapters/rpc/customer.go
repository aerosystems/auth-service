package RpcRepo

import (
	RpcClient "github.com/aerosystems/auth-service/pkg/rpc_client"
	"github.com/google/uuid"
)

type CustomerAdapter struct {
	rpcClient *RpcClient.ReconnectRpcClient
}

func NewCustomerAdapter(rpcClient *RpcClient.ReconnectRpcClient) *CustomerAdapter {
	return &CustomerAdapter{
		rpcClient: rpcClient,
	}
}

type CustomerRPCPayload struct {
	Uuid uuid.UUID
}

func (c *CustomerAdapter) CreateCustomer() (uuid.UUID, error) {
	result := CustomerRPCPayload{}
	if err := c.rpcClient.Call("Server.CreateCustomer", "", &result); err != nil {
		return uuid.UUID{}, err
	}
	return result.Uuid, nil
}
