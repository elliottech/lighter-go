package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elliottech/lighter-go/client"
	lighterHTTP "github.com/elliottech/lighter-go/client/http"
	"github.com/elliottech/lighter-go/types"
	"github.com/elliottech/lighter-go/types/txtypes"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var chainId uint32

func main() {
	host := flag.String("host", getEnv("LIGHTER_HOST", "0.0.0.0"), "host to listen on")
	port := flag.String("port", getEnv("LIGHTER_PORT", "8080"), "port to listen on")
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/generate-api-key", withRecover(handleGenerateAPIKey))
	mux.HandleFunc("/create-client", withRecover(handleCreateClient))
	mux.HandleFunc("/check-client", withRecover(handleCheckClient))
	mux.HandleFunc("/sign-change-pub-key", withRecover(handleSignChangePubKey))
	mux.HandleFunc("/sign-create-order", withRecover(handleSignCreateOrder))
	mux.HandleFunc("/sign-create-grouped-orders", withRecover(handleSignCreateGroupedOrders))
	mux.HandleFunc("/sign-cancel-order", withRecover(handleSignCancelOrder))
	mux.HandleFunc("/sign-withdraw", withRecover(handleSignWithdraw))
	mux.HandleFunc("/sign-create-sub-account", withRecover(handleSignCreateSubAccount))
	mux.HandleFunc("/sign-cancel-all-orders", withRecover(handleSignCancelAllOrders))
	mux.HandleFunc("/sign-modify-order", withRecover(handleSignModifyOrder))
	mux.HandleFunc("/sign-transfer", withRecover(handleSignTransfer))
	mux.HandleFunc("/sign-create-public-pool", withRecover(handleSignCreatePublicPool))
	mux.HandleFunc("/sign-update-public-pool", withRecover(handleSignUpdatePublicPool))
	mux.HandleFunc("/sign-mint-shares", withRecover(handleSignMintShares))
	mux.HandleFunc("/sign-burn-shares", withRecover(handleSignBurnShares))
	mux.HandleFunc("/sign-update-leverage", withRecover(handleSignUpdateLeverage))
	mux.HandleFunc("/create-auth-token", withRecover(handleCreateAuthToken))
	mux.HandleFunc("/sign-update-margin", withRecover(handleSignUpdateMargin))
	mux.HandleFunc("/sign-stake-assets", withRecover(handleSignStakeAssets))
	mux.HandleFunc("/sign-unstake-assets", withRecover(handleSignUnstakeAssets))
	mux.HandleFunc("/sign-approve-integrator", withRecover(handleSignApproveIntegrator))
	mux.HandleFunc("/sign-update-account-config", withRecover(handleSignUpdateAccountConfig))
	mux.HandleFunc("/sign-update-account-asset-config", withRecover(handleSignUpdateAccountAssetConfig))

	addr := *host + ":" + *port
	server := &http.Server{Addr: addr, Handler: mux}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		server.Close()
	}()

	log.Printf("listening on %s\n", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v\n", err)
	}
}

// --- Helpers ---

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func withRecover(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				writeJSON(w, http.StatusInternalServerError, ErrorResp{Error: fmt.Sprintf("panic: %v", rec)})
			}
		}()
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorResp{Error: "method not allowed"})
			return
		}
		h(w, r)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func decodeBody(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func getClient(apiKeyIndex int, accountIndex int64) (*client.TxClient, error) {
	return client.GetClient(uint8(apiKeyIndex), accountIndex)
}

func createTxAttributesFromSkipNonce(skipNonce uint8) *types.L2TxAttributes {
	attr := types.L2TxAttributes{}
	if skipNonce == 1 {
		attr.SkipNonce = &skipNonce
	}
	return &attr
}

func createIntegratorTxAttributes(integratorAccountIndex int64, integratorTakerFee uint32, integratorMakerFee uint32, skipNonce uint8) *types.L2TxAttributes {
	attr := types.L2TxAttributes{}
	attr.IntegratorAccountIndex = &integratorAccountIndex
	attr.IntegratorTakerFee = &integratorTakerFee
	attr.IntegratorMakerFee = &integratorMakerFee
	if skipNonce == 1 {
		attr.SkipNonce = &skipNonce
	}
	return &attr
}

func getTransactOpts(skipNonce uint8, nonce int64) *types.TransactOpts {
	txAttributes := createTxAttributesFromSkipNonce(skipNonce)
	return &types.TransactOpts{
		Nonce:        &nonce,
		TxAttributes: txAttributes,
	}
}

func getIntegratorTransactOpts(integratorAccountIndex int64, integratorTakerFee uint32, integratorMakerFee uint32, skipNonce uint8, nonce int64) *types.TransactOpts {
	txAttributes := createIntegratorTxAttributes(integratorAccountIndex, integratorTakerFee, integratorMakerFee, skipNonce)
	return &types.TransactOpts{
		Nonce:        &nonce,
		TxAttributes: txAttributes,
	}
}

func messageToSign(txInfo txtypes.TxInfo) string {
	switch typed := txInfo.(type) {
	case *txtypes.L2ChangePubKeyTxInfo:
		return typed.GetL1SignatureBody()
	case *txtypes.L2TransferTxInfo:
		return typed.GetL1SignatureBody(chainId)
	case *txtypes.L2ApproveIntegratorTxInfo:
		return typed.GetL1SignatureBody(chainId)
	default:
		return ""
	}
}

func convertTxInfoToResponse(txInfo txtypes.TxInfo, err error) SignedTxResp {
	if err != nil {
		return SignedTxResp{Error: err.Error()}
	}
	if txInfo == nil {
		return SignedTxResp{Error: "nil transaction info"}
	}

	txInfoStr, err := txInfo.GetTxInfo()
	if err != nil {
		return SignedTxResp{Error: err.Error()}
	}

	resp := SignedTxResp{
		TxType: txInfo.GetTxType(),
		TxInfo: txInfoStr,
		TxHash: txInfo.GetTxHash(),
	}

	if msg := messageToSign(txInfo); msg != "" {
		resp.MessageToSign = msg
	}

	return resp
}

// --- Handlers ---

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleGenerateAPIKey(w http.ResponseWriter, _ *http.Request) {
	privateKeyStr, publicKeyStr, err := client.GenerateAPIKey()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ApiKeyResp{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, ApiKeyResp{PrivateKey: privateKeyStr, PublicKey: publicKeyStr})
}

func handleCreateClient(w http.ResponseWriter, r *http.Request) {
	var req CreateClientReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResp{Error: err.Error()})
		return
	}

	chainId = uint32(req.ChainID)
	httpClient := lighterHTTP.NewClient(req.URL)

	_, err := client.CreateClient(httpClient, req.PrivateKey, chainId, uint8(req.ApiKeyIndex), req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResp{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, ErrorResp{})
}

func handleCheckClient(w http.ResponseWriter, r *http.Request) {
	var req CheckClientReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResp{Error: err.Error()})
		return
	}

	if err := c.Check(); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResp{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, ErrorResp{})
}

func handleSignChangePubKey(w http.ResponseWriter, r *http.Request) {
	var req ChangePubKeyReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	pubKeyBytes, err := hexutil.Decode(req.PubKey)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}
	if len(pubKeyBytes) != 40 {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: fmt.Sprintf("invalid pub key length. expected 40 but got %v", len(pubKeyBytes))})
		return
	}
	var pubKey [40]byte
	copy(pubKey[:], pubKeyBytes)

	tx := &types.ChangePubKeyReq{PubKey: pubKey}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetChangePubKeyTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	orderExpiry := req.OrderExpiry
	if orderExpiry == -1 {
		orderExpiry = time.Now().Add(time.Hour * 24 * 28).UnixMilli()
	}

	tx := &types.CreateOrderTxReq{
		MarketIndex:      req.MarketIndex,
		ClientOrderIndex: req.ClientOrderIndex,
		BaseAmount:       req.BaseAmount,
		Price:            req.Price,
		IsAsk:            req.IsAsk,
		Type:             req.OrderType,
		TimeInForce:      req.TimeInForce,
		ReduceOnly:       req.ReduceOnly,
		TriggerPrice:     req.TriggerPrice,
		OrderExpiry:      orderExpiry,
	}
	ops := getIntegratorTransactOpts(req.IntegratorAccountIndex, req.IntegratorTakerFee, req.IntegratorMakerFee, req.SkipNonce, req.Nonce)

	txInfo, err := c.GetCreateOrderTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignCreateGroupedOrders(w http.ResponseWriter, r *http.Request) {
	var req CreateGroupedOrdersReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	orders := make([]*types.CreateOrderTxReq, len(req.Orders))
	for i, o := range req.Orders {
		orderExpiry := o.OrderExpiry
		if orderExpiry == -1 {
			orderExpiry = time.Now().Add(time.Hour * 24 * 28).UnixMilli()
		}
		orders[i] = &types.CreateOrderTxReq{
			MarketIndex:      o.MarketIndex,
			ClientOrderIndex: o.ClientOrderIndex,
			BaseAmount:       o.BaseAmount,
			Price:            o.Price,
			IsAsk:            o.IsAsk,
			Type:             o.Type,
			TimeInForce:      o.TimeInForce,
			ReduceOnly:       o.ReduceOnly,
			TriggerPrice:     o.TriggerPrice,
			OrderExpiry:      orderExpiry,
		}
	}

	tx := &types.CreateGroupedOrdersTxReq{
		GroupingType: req.GroupingType,
		Orders:       orders,
	}
	ops := getIntegratorTransactOpts(req.IntegratorAccountIndex, req.IntegratorTakerFee, req.IntegratorMakerFee, req.SkipNonce, req.Nonce)

	txInfo, err := c.GetCreateGroupedOrdersTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignCancelOrder(w http.ResponseWriter, r *http.Request) {
	var req CancelOrderReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.CancelOrderTxReq{
		MarketIndex: req.MarketIndex,
		Index:       req.OrderIndex,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetCancelOrderTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignWithdraw(w http.ResponseWriter, r *http.Request) {
	var req WithdrawReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.WithdrawTxReq{
		AssetIndex: req.AssetIndex,
		RouteType:  req.RouteType,
		Amount:     req.Amount,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetWithdrawTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignCreateSubAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateSubAccountReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	ops := getTransactOpts(req.SkipNonce, req.Nonce)
	txInfo, err := c.GetCreateSubAccountTransaction(ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignCancelAllOrders(w http.ResponseWriter, r *http.Request) {
	var req CancelAllOrdersReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.CancelAllOrdersTxReq{
		TimeInForce: req.TimeInForce,
		Time:        req.Time,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetCancelAllOrdersTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignModifyOrder(w http.ResponseWriter, r *http.Request) {
	var req ModifyOrderReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.ModifyOrderTxReq{
		MarketIndex:  req.MarketIndex,
		Index:        req.Index,
		BaseAmount:   req.BaseAmount,
		Price:        req.Price,
		TriggerPrice: req.TriggerPrice,
	}
	ops := getIntegratorTransactOpts(req.IntegratorAccountIndex, req.IntegratorTakerFee, req.IntegratorMakerFee, req.SkipNonce, req.Nonce)

	txInfo, err := c.GetModifyOrderTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignTransfer(w http.ResponseWriter, r *http.Request) {
	var req TransferReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	var memo [32]byte
	memoStr := req.Memo
	if len(memoStr) == 66 {
		if memoStr[0:2] == "0x" {
			memoStr = memoStr[2:66]
		} else {
			writeJSON(w, http.StatusOK, SignedTxResp{Error: fmt.Sprintf("memo expected to be 32 bytes or 64 hex encoded or 66 if 0x hex encoded -- long but received %v", len(memoStr))})
			return
		}
	}

	if len(memoStr) == 64 {
		b, err := hex.DecodeString(memoStr)
		if err != nil {
			writeJSON(w, http.StatusOK, SignedTxResp{Error: fmt.Sprintf("failed to decode hex string. err: %v", err)})
			return
		}
		for i := 0; i < 32; i++ {
			memo[i] = b[i]
		}
	} else if len(memoStr) == 32 {
		for i := 0; i < 32; i++ {
			memo[i] = byte(memoStr[i])
		}
	} else {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: fmt.Sprintf("memo expected to be 32 bytes or 64 hex encoded or 66 if 0x hex encoded -- long but received %v", len(memoStr))})
		return
	}

	tx := &types.TransferTxReq{
		ToAccountIndex: req.ToAccountIndex,
		AssetIndex:     req.AssetIndex,
		FromRouteType:  req.FromRouteType,
		ToRouteType:    req.ToRouteType,
		Amount:         req.Amount,
		USDCFee:        req.USDCFee,
		Memo:           memo,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetTransferTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignCreatePublicPool(w http.ResponseWriter, r *http.Request) {
	var req CreatePublicPoolReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.CreatePublicPoolTxReq{
		OperatorFee:          req.OperatorFee,
		InitialTotalShares:   req.InitialTotalShares,
		MinOperatorShareRate: req.MinOperatorShareRate,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetCreatePublicPoolTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignUpdatePublicPool(w http.ResponseWriter, r *http.Request) {
	var req UpdatePublicPoolReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.UpdatePublicPoolTxReq{
		PublicPoolIndex:      req.PublicPoolIndex,
		Status:               req.Status,
		OperatorFee:          req.OperatorFee,
		MinOperatorShareRate: req.MinOperatorShareRate,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetUpdatePublicPoolTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignMintShares(w http.ResponseWriter, r *http.Request) {
	var req MintSharesReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.MintSharesTxReq{
		PublicPoolIndex: req.PublicPoolIndex,
		ShareAmount:     req.ShareAmount,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetMintSharesTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignBurnShares(w http.ResponseWriter, r *http.Request) {
	var req BurnSharesReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.BurnSharesTxReq{
		PublicPoolIndex: req.PublicPoolIndex,
		ShareAmount:     req.ShareAmount,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetBurnSharesTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignUpdateLeverage(w http.ResponseWriter, r *http.Request) {
	var req UpdateLeverageReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.UpdateLeverageTxReq{
		MarketIndex:           req.MarketIndex,
		InitialMarginFraction: req.InitialMarginFraction,
		MarginMode:            req.MarginMode,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetUpdateLeverageTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleCreateAuthToken(w http.ResponseWriter, r *http.Request) {
	var req CreateAuthTokenReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, StrOrErrResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, StrOrErrResp{Error: err.Error()})
		return
	}

	deadline := req.Deadline
	if deadline == 0 {
		deadline = time.Now().Add(time.Hour * 7).Unix()
	}

	authToken, err := c.GetAuthToken(time.Unix(deadline, 0))
	if err != nil {
		writeJSON(w, http.StatusOK, StrOrErrResp{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, StrOrErrResp{Result: authToken})
}

func handleSignUpdateMargin(w http.ResponseWriter, r *http.Request) {
	var req UpdateMarginReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.UpdateMarginTxReq{
		MarketIndex: req.MarketIndex,
		USDCAmount:  req.USDCAmount,
		Direction:   req.Direction,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetUpdateMarginTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignStakeAssets(w http.ResponseWriter, r *http.Request) {
	var req StakeAssetsReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.StakeAssetsTxReq{
		StakingPoolIndex: req.StakingPoolIndex,
		ShareAmount:      req.ShareAmount,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetStakeAssetsTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignUnstakeAssets(w http.ResponseWriter, r *http.Request) {
	var req UnstakeAssetsReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.UnstakeAssetsTxReq{
		StakingPoolIndex: req.StakingPoolIndex,
		ShareAmount:      req.ShareAmount,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetUnstakeAssetsTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignApproveIntegrator(w http.ResponseWriter, r *http.Request) {
	var req ApproveIntegratorReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.ApproveIntegratorTxReq{
		IntegratorAccountIndex: req.IntegratorIndex,
		MaxPerpsTakerFee:       req.MaxPerpsTakerFee,
		MaxPerpsMakerFee:       req.MaxPerpsMakerFee,
		MaxSpotTakerFee:        req.MaxSpotTakerFee,
		MaxSpotMakerFee:        req.MaxSpotMakerFee,
		ApprovalExpiry:         req.ApprovalExpiry,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetApproveIntegratorTx(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignUpdateAccountConfig(w http.ResponseWriter, r *http.Request) {
	var req UpdateAccountConfigReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.UpdateAccountConfigTxReq{
		AccountTradingMode: req.AccountTradingMode,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetUpdateAccountConfigTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}

func handleSignUpdateAccountAssetConfig(w http.ResponseWriter, r *http.Request) {
	var req UpdateAccountAssetConfigReq
	if err := decodeBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, SignedTxResp{Error: err.Error()})
		return
	}

	c, err := getClient(req.ApiKeyIndex, req.AccountIndex)
	if err != nil {
		writeJSON(w, http.StatusOK, SignedTxResp{Error: err.Error()})
		return
	}

	tx := &types.UpdateAccountAssetConfigTxReq{
		AssetIndex:      req.AssetIndex,
		AssetMarginMode: req.AssetMarginMode,
	}
	ops := getTransactOpts(req.SkipNonce, req.Nonce)

	txInfo, err := c.GetUpdateAccountAssetConfigTransaction(tx, ops)
	writeJSON(w, http.StatusOK, convertTxInfoToResponse(txInfo, err))
}
