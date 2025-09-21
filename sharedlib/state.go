package main

import "github.com/elliottech/lighter-go/client"

var (
	txClient        *client.TxClient
	backupTxClients map[uint8]*client.TxClient
)
