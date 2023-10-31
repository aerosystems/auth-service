package RPCServices

import (
	"github.com/google/uuid"
	"net/rpc"
)

type CustomerService interface {
	CreateCustomer() (uuid.UUID, error)
}

type CustomerRPC struct {
	rpcClient *rpc.Client
}

func NewCustomerRPC(rpcClient *rpc.Client) *CustomerRPC {
	return &CustomerRPC{
		rpcClient: rpcClient,
	}
}

type CustomerRPCPayload struct {
	Uuid uuid.UUID
}

func (c *CustomerRPC) CreateCustomer() (uuid.UUID, error) {
	result := CustomerRPCPayload{}
	if err := c.rpcClient.Call("CustomerServer.CreateCustomer", nil, &result); err != nil {
		return uuid.UUID{}, err
	}
	return result.Uuid, nil
}
