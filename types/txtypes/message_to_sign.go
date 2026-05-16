package txtypes

func MessageToSign(txInfo TxInfo, chainId uint32) string {
	switch typed := txInfo.(type) {
	case *L2ChangePubKeyTxInfo:
		return typed.GetL1SignatureBody()
	case *L2TransferTxInfo:
		return typed.GetL1SignatureBody(chainId)
	case *L2ApproveIntegratorTxInfo:
		return typed.GetL1SignatureBody(chainId)
	default:
		return ""
	}
}
