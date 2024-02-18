package RpcRepo

import (
	RPCClient "github.com/aerosystems/auth-service/pkg/rpc_client"
	"github.com/google/uuid"
)

type CustomerRepo struct {
	rpcClient *RPCClient.ReconnectRPCClient
}

func NewCustomerRepo(rpcClient *RPCClient.ReconnectRPCClient) *CustomerRepo {
	return &CustomerRepo{
		rpcClient: rpcClient,
	}
}

type CustomerRPCPayload struct {
	Uuid uuid.UUID
}

func (c *CustomerRepo) CreateCustomer() (uuid.UUID, error) {
	result := CustomerRPCPayload{}
	if err := c.rpcClient.Call("CustomerServer.CreateCustomer", "", &result); err != nil {
		return uuid.UUID{}, err
	}
	return result.Uuid, nil
}
