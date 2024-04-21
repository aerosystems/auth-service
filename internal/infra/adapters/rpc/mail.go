package RpcRepo

import (
	RpcClient "github.com/aerosystems/auth-service/pkg/rpc_client"
)

type MailAdapter struct {
	rpcClient *RpcClient.ReconnectRpcClient
}

func NewMailAdapter(rpcClient *RpcClient.ReconnectRpcClient) *MailAdapter {
	return &MailAdapter{
		rpcClient: rpcClient,
	}
}

type MailRPCPayload struct {
	To      string
	Subject string
	Body    string
}

func (ma *MailAdapter) SendEmail(to, subject, body string) error {
	var result string
	if err := ma.rpcClient.Call(
		"Server.SendEmail",
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
