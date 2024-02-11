package rpcRepo

import (
	RPCClient "github.com/aerosystems/auth-service/pkg/rpc_client"
)

type MailRepo struct {
	rpcClient *RPCClient.ReconnectRPCClient
}

func NewMailRepo(rpcClient *RPCClient.ReconnectRPCClient) *MailRepo {
	return &MailRepo{
		rpcClient: rpcClient,
	}
}

type MailRPCPayload struct {
	To      string
	Subject string
	Body    string
}

func (ms *MailRepo) SendEmail(to, subject, body string) error {
	var result string
	if err := ms.rpcClient.Call(
		"MailServer.SendEmail",
		MailRPCPayload{
			To:      to,
			Subject: subject,
			Body:    body,
		},
		&result); err != nil {
		return err
	}

	return nil
}
