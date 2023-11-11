package RPCServices

import (
	"net/rpc"
)

type CustomerService interface {
	CreateCustomer() (int, error)
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
	UserId int
}

func (c *CustomerRPC) CreateCustomer() (int, error) {
	result := CustomerRPCPayload{}
	if err := c.rpcClient.Call("CustomerServer.CreateCustomer", nil, &result); err != nil {
		return 0, err
	}
	return result.UserId, nil
}
