package network_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/bytes"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	betteryapp "bettery/app"
	eventstypes "bettery/x/events/types"
	fundstypes "bettery/x/funds/types"
	guardtypes "bettery/x/guard/types"
)

// findModuleRoot returns the directory containing go.mod, or cwd if not found.
func findModuleRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	cwd, _ := os.Getwd()
	return cwd
}

// queryBalanceViaRPC queries bank balance via ABCI (RPC).
func queryBalanceViaRPC(ctx context.Context, rpcClient interface {
	ABCIQuery(ctx context.Context, path string, data bytes.HexBytes) (*coretypes.ResultABCIQuery, error)
}, address, denom string) (uint64, error) {
	req := &banktypes.QueryBalanceRequest{Address: address, Denom: denom}
	data, err := proto.Marshal(req)
	if err != nil {
		return 0, err
	}
	res, err := rpcClient.ABCIQuery(ctx, "/cosmos.bank.v1beta1.Query/Balance", bytes.HexBytes(data))
	if err != nil {
		return 0, err
	}
	if res.Response.Code != 0 {
		return 0, fmt.Errorf("ABCI query code %d: %s", res.Response.Code, res.Response.Log)
	}
	var qres banktypes.QueryBalanceResponse
	if err := proto.Unmarshal(res.Response.Value, &qres); err != nil {
		return 0, err
	}
	if qres.Balance == nil || qres.Balance.Amount.IsNil() {
		return 0, nil
	}
	return qres.Balance.Amount.Uint64(), nil
}

// waitForBlockTimeAfter blocks until the chain's latest block time is at or after the given Unix timestamp.
func waitForBlockTimeAfter(ctx context.Context, rpcClient interface {
	Status(ctx context.Context) (*coretypes.ResultStatus, error)
}, endTimeUnix uint64, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	end := time.Unix(int64(endTimeUnix), 0)
	for time.Now().Before(deadline) {
		status, err := rpcClient.Status(ctx)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		if !status.SyncInfo.LatestBlockTime.Before(end) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return fmt.Errorf("timed out waiting for block time >= %v", end)
}

// TestNetworkCreateEvent runs against an already-running chain (e.g. started with
// `ignite chain serve -v`). It sends CreateEvent, SetOwner, SetCompanyPercent,
// participant mints and CreatePartEvent, then ValidateEvent, and asserts events
// and balance changes. Use this to test the blockchain and TypeScript indexer together.
//
// Prerequisites:
//  1. Start the chain: ignite chain serve -v  (or betteryd start)
//  2. Optionally start the indexer (e.g. in Docker with RPC_URL=http://host.docker.internal:26657)
//  3. Run the test: go test -count=1 ./testutil/network -run TestNetworkCreateEvent -v
//
// Environment:
//   - RPC_URL: RPC endpoint (default http://localhost:26657)
//   - CHAIN_ID: chain ID (default: read from node status)
//   - BETTERYD_HOME: keyring home, must contain the validator key (default ~/.betteryd)
//   - KEY_NAME: key name to use as creator/owner (default "alice", common with ignite)
//   - TEST_NETWORK_INDEXER_WAIT: optional duration to wait after test for indexer (e.g. 60s)
func TestNetworkCreateEvent(t *testing.T) {
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		rpcURL = "http://localhost:26657"
	}
	keyName := os.Getenv("KEY_NAME")
	if keyName == "" {
		keyName = "alice"
	}

	// Connect to the live node.
	wsURL := strings.Replace(rpcURL, "http://", "ws://", 1)
	if !strings.Contains(wsURL, "://") {
		wsURL = "ws://" + wsURL
	}
	if !strings.HasSuffix(wsURL, "/websocket") {
		wsURL = strings.TrimSuffix(wsURL, "/") + "/websocket"
	}
	rpcClient, err := rpchttp.New(rpcURL, wsURL)
	require.NoError(t, err, "connect to RPC: ensure the chain is running (e.g. ignite chain serve -v)")

	ctx := context.Background()
	status, err := rpcClient.Status(ctx)
	require.NoError(t, err, "RPC status: is the chain running?")
	chainID := os.Getenv("CHAIN_ID")
	if chainID == "" {
		chainID = status.NodeInfo.Network
	}
	require.NotEmpty(t, chainID, "chain ID required (from node or CHAIN_ID env)")

	// gRPC connection for account/auth queries (AccountRetriever uses it).
	grpcAddr := os.Getenv("GRPC_URL")
	if grpcAddr == "" {
		u, _ := url.Parse(rpcURL)
		host := "localhost"
		if u != nil && u.Host != "" {
			host = u.Hostname()
		}
		grpcAddr = host + ":9090"
	}
	grpcConn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "gRPC dial %s (set GRPC_URL if your node uses a different port)", grpcAddr)
	t.Cleanup(func() { _ = grpcConn.Close() })

	// App codec and TxConfig (same as the running chain).
	db := dbm.NewMemDB()
	appOpts := viper.New()
	app := betteryapp.New(log.NewNopLogger(), db, nil, false, appOpts)
	txConfig := app.TxConfig()
	codec := app.AppCodec()

	// Keyring: try BETTERYD_HOME, then common locations. Paths relative to module root (where go.mod is).
	moduleRoot := findModuleRoot()
	resolve := func(p string) string {
		if p == "" {
			return ""
		}
		if filepath.IsAbs(p) {
			return p
		}
		return filepath.Join(moduleRoot, p)
	}
	homeEnv := os.Getenv("BETTERYD_HOME")
	var candidateHomes []string
	if homeEnv != "" {
		candidateHomes = []string{resolve(homeEnv)}
	} else {
		userHome, _ := os.UserHomeDir()
		candidateHomes = []string{
			userHome + "/.betteryd",
			filepath.Join(moduleRoot, "data"),
			filepath.Join(moduleRoot, ".bettery"),
			filepath.Join(userHome, ".ignite", "chains", "bettery"),
		}
	}
	var kb keyring.Keyring
	var keyringHome string
	for _, home := range candidateHomes {
		if home == "" {
			continue
		}
		k, err := keyring.New("bettery", "test", home, nil, codec)
		if err != nil {
			continue
		}
		_, err = k.Key(keyName)
		if err == nil {
			kb = k
			keyringHome = home
			break
		}
	}
	if kb == nil {
		t.Fatalf("key %q not found in any of: %v (set BETTERYD_HOME to the chain home, e.g. ./data when using ignite chain serve)", keyName, candidateHomes)
	}
	t.Logf("using key %q from %s", keyName, keyringHome)
	record, err := kb.Key(keyName)
	require.NoError(t, err)
	creatorAddr, err := record.GetAddress()
	require.NoError(t, err)

	clientCtx := client.Context{}.
		WithCodec(codec).
		WithInterfaceRegistry(app.InterfaceRegistry()).
		WithTxConfig(txConfig).
		WithKeyring(kb).
		WithClient(rpcClient).
		WithGRPCClient(grpcConn).
		WithNodeURI(rpcURL).
		WithChainID(chainID).
		WithFromName(keyName).
		WithFromAddress(creatorAddr).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithSkipConfirmation(true).
		WithCmdContext(ctx)

	// Gas price accepted by the app (match chain config).
	minGasPrices := "0.000006ubet"

	// --- CreateEvent ---
	// endTime: participants can join until block time < endTime; validation allowed only after block time >= endTime.
	// Use enough lead time for all participant txs (20 × 2 txs = 40 txs, ~2s/block) plus wait before ValidateEvent.
	endTime := uint64(time.Now().Unix() + 120)
	msg := &eventstypes.MsgCreateEvent{
		Creator:  creatorAddr.String(),
		Question: "Who wins the match?",
		Answers:  []string{"Team A", "Team B"},
		EndTime:  endTime,
		Category: "sports",
		RoomId:   "room-1",
	}
	txf := tx.Factory{}.
		WithTxConfig(clientCtx.TxConfig).
		WithChainID(chainID).
		WithKeybase(clientCtx.Keyring).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithGas(200000).
		WithGasPrices(minGasPrices)
	txf, err = txf.Prepare(clientCtx)
	require.NoError(t, err)
	txBuilder, err := txf.BuildUnsignedTx(msg)
	require.NoError(t, err)
	err = tx.Sign(clientCtx.CmdContext, txf, clientCtx.FromName, txBuilder, true)
	require.NoError(t, err)
	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	require.NoError(t, err)
	res, err := rpcClient.BroadcastTxCommit(ctx, txBytes)
	require.NoError(t, err)
	require.Equal(t, abci.CodeTypeOK, res.CheckTx.Code, "CheckTx failed: %v", res.CheckTx.Log)
	require.Equal(t, abci.CodeTypeOK, res.TxResult.Code, "TxResult failed: %v", res.TxResult.Log)

	var attrs map[string]string
	for _, ev := range res.TxResult.Events {
		if ev.Type != "CREATE_EVENT" {
			continue
		}
		attrs = make(map[string]string, len(ev.Attributes))
		for _, a := range ev.Attributes {
			attrs[string(a.Key)] = string(a.Value)
		}
		break
	}
	require.NotEmpty(t, attrs, "CREATE_EVENT not found in DeliverTx events")
	require.Equal(t, msg.Creator, attrs["creator"])
	require.Equal(t, msg.Question, attrs["question"])
	require.Equal(t, fmt.Sprintf("%d", msg.EndTime), attrs["endTime"])

	eventIDStr := attrs["id"]
	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
	require.NoError(t, err)

	// --- SetOwner (guard) ---
	setOwnerMsg := &guardtypes.MsgSetOwner{Creator: creatorAddr.String()}
	ownerTxf := tx.Factory{}.
		WithTxConfig(clientCtx.TxConfig).
		WithChainID(chainID).
		WithKeybase(clientCtx.Keyring).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithGas(200000).
		WithGasPrices(minGasPrices)
	ownerTxf, err = ownerTxf.Prepare(clientCtx)
	require.NoError(t, err)
	ownerBuilder, err := ownerTxf.BuildUnsignedTx(setOwnerMsg)
	require.NoError(t, err)
	err = tx.Sign(clientCtx.CmdContext, ownerTxf, clientCtx.FromName, ownerBuilder, true)
	require.NoError(t, err)
	ownerBytes, _ := clientCtx.TxConfig.TxEncoder()(ownerBuilder.GetTx())
	ownerRes, err := rpcClient.BroadcastTxCommit(ctx, ownerBytes)
	require.NoError(t, err)
	require.Equal(t, abci.CodeTypeOK, ownerRes.CheckTx.Code)
	require.Equal(t, abci.CodeTypeOK, ownerRes.TxResult.Code)

	// --- SetCompanyPercent ---
	setOwnerPercent := &fundstypes.MsgSetCompanyPercent{Creator: creatorAddr.String(), Percent: 1}
	percentTxf := tx.Factory{}.
		WithTxConfig(clientCtx.TxConfig).
		WithChainID(chainID).
		WithKeybase(clientCtx.Keyring).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithGas(200000).
		WithGasPrices(minGasPrices)
	percentTxf, err = percentTxf.Prepare(clientCtx)
	require.NoError(t, err)
	percentBuilder, err := percentTxf.BuildUnsignedTx(setOwnerPercent)
	require.NoError(t, err)
	err = tx.Sign(clientCtx.CmdContext, percentTxf, clientCtx.FromName, percentBuilder, true)
	require.NoError(t, err)
	percentBytes, _ := clientCtx.TxConfig.TxEncoder()(percentBuilder.GetTx())
	percentRes, err := rpcClient.BroadcastTxCommit(ctx, percentBytes)
	require.NoError(t, err)
	require.Equal(t, abci.CodeTypeOK, percentRes.CheckTx.Code)
	require.Equal(t, abci.CodeTypeOK, percentRes.TxResult.Code)

	// --- Participants: mint + CreatePartEvent ---
	type participantInfo struct {
		Address string
		Amount  uint64
		Answer  string
	}
	var participants []participantInfo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var lastPartRes *coretypes.ResultBroadcastTxCommit

	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("participant-%d", i)
		record, _, err := clientCtx.Keyring.NewMnemonic(name, keyring.English, "", "", hd.Secp256k1)
		require.NoError(t, err)
		partAddr, err := record.GetAddress()
		require.NoError(t, err)

		mintMsg := &fundstypes.MsgMintToken{Creator: creatorAddr.String(), Receiver: partAddr.String()}
		mintTxf := tx.Factory{}.
			WithTxConfig(clientCtx.TxConfig).
			WithChainID(chainID).
			WithKeybase(clientCtx.Keyring).
			WithAccountRetriever(clientCtx.AccountRetriever).
			WithGas(200000).
			WithGasPrices(minGasPrices)
		mintTxf, err = mintTxf.Prepare(clientCtx)
		require.NoError(t, err)
		mintBuilder, err := mintTxf.BuildUnsignedTx(mintMsg)
		require.NoError(t, err)
		err = tx.Sign(clientCtx.CmdContext, mintTxf, clientCtx.FromName, mintBuilder, true)
		require.NoError(t, err)
		mintBytes, _ := clientCtx.TxConfig.TxEncoder()(mintBuilder.GetTx())
		mintRes, err := rpcClient.BroadcastTxCommit(ctx, mintBytes)
		require.NoError(t, err)
		require.Equal(t, abci.CodeTypeOK, mintRes.CheckTx.Code, "participant %d mint CheckTx: %s", i, mintRes.CheckTx.Log)
		require.Equal(t, abci.CodeTypeOK, mintRes.TxResult.Code, "participant %d mint DeliverTx: %s", i, mintRes.TxResult.Log)

		amount := uint64(r.Int63n(1000000000) + 1)
		answer := msg.Answers[r.Intn(len(msg.Answers))]
		participants = append(participants, participantInfo{Address: partAddr.String(), Amount: amount, Answer: answer})

		partMsg := &eventstypes.MsgCreatePartEvent{
			Creator: partAddr.String(),
			EventId: eventID,
			Answers: answer,
			Amount:  fmt.Sprintf("%d", amount),
		}
		partCtx := clientCtx.WithFromAddress(partAddr).WithFromName(name)
		partTxf := tx.Factory{}.
			WithTxConfig(partCtx.TxConfig).
			WithChainID(chainID).
			WithKeybase(partCtx.Keyring).
			WithAccountRetriever(partCtx.AccountRetriever).
			WithGas(200000).
			WithGasPrices(minGasPrices)
		partTxf, err = partTxf.Prepare(partCtx)
		require.NoError(t, err)
		partBuilder, err := partTxf.BuildUnsignedTx(partMsg)
		require.NoError(t, err)
		err = tx.Sign(partCtx.CmdContext, partTxf, partCtx.FromName, partBuilder, true)
		require.NoError(t, err)
		partBytes, _ := partCtx.TxConfig.TxEncoder()(partBuilder.GetTx())
		lastPartRes, err = rpcClient.BroadcastTxCommit(ctx, partBytes)
		require.NoError(t, err)
		require.Equal(t, abci.CodeTypeOK, lastPartRes.CheckTx.Code, "participant %d CheckTx: %s", i, lastPartRes.CheckTx.Log)
		require.Equal(t, abci.CodeTypeOK, lastPartRes.TxResult.Code, "participant %d DeliverTx: %s", i, lastPartRes.TxResult.Log)
	}

	var partFound bool
	for _, ev := range lastPartRes.TxResult.Events {
		if ev.Type == "PARTICIPATE_EVENT" {
			partFound = true
			break
		}
	}
	require.True(t, partFound, "PARTICIPATE_EVENT not found in participant tx events")

	// Balances before ValidateEvent
	balancesBefore := make([]uint64, len(participants))
	for i, p := range participants {
		bal, err := queryBalanceViaRPC(ctx, rpcClient, p.Address, "ubet")
		require.NoError(t, err)
		balancesBefore[i] = bal
	}

	// Wait until chain block time is past the event's endTime (validation not allowed before that).
	err = waitForBlockTimeAfter(ctx, rpcClient, endTime, 2*time.Minute)
	require.NoError(t, err)

	// --- ValidateEvent ---
	validateMsg := &eventstypes.MsgValidateEvent{
		Creator: creatorAddr.String(),
		EventId: eventID,
		Answers: msg.Answers[0],
		Source:  "test",
	}
	validateTxf := tx.Factory{}.
		WithTxConfig(clientCtx.TxConfig).
		WithChainID(chainID).
		WithKeybase(clientCtx.Keyring).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithGas(3000000).
		WithGasPrices(minGasPrices)
	validateTxf, err = validateTxf.Prepare(clientCtx)
	require.NoError(t, err)
	validateBuilder, err := validateTxf.BuildUnsignedTx(validateMsg)
	require.NoError(t, err)
	err = tx.Sign(clientCtx.CmdContext, validateTxf, clientCtx.FromName, validateBuilder, true)
	require.NoError(t, err)
	validateBytes, _ := clientCtx.TxConfig.TxEncoder()(validateBuilder.GetTx())
	validateRes, err := rpcClient.BroadcastTxCommit(ctx, validateBytes)
	require.NoError(t, err)
	require.Equal(t, abci.CodeTypeOK, validateRes.CheckTx.Code, "ValidateEvent CheckTx: %s", validateRes.CheckTx.Log)
	require.Equal(t, abci.CodeTypeOK, validateRes.TxResult.Code, "ValidateEvent DeliverTx: %s", validateRes.TxResult.Log)

	var validateEventFound bool
	for _, ev := range validateRes.TxResult.Events {
		if ev.Type != "VALIDATE_EVENT" {
			continue
		}
		validateEventFound = true
		validateAttrs := make(map[string]string, len(ev.Attributes))
		for _, a := range ev.Attributes {
			validateAttrs[string(a.Key)] = string(a.Value)
		}
		require.Equal(t, creatorAddr.String(), validateAttrs["creator"])
		require.Equal(t, fmt.Sprintf("%d", eventID), validateAttrs["eventId"])
		require.Equal(t, msg.Answers[0], validateAttrs["answer"])
		require.Equal(t, "test", validateAttrs["source"])
		break
	}
	require.True(t, validateEventFound, "VALIDATE_EVENT not found in ValidateEvent tx events")

	// --- Balance assertions (letsPayWinners formula) ---
	winningAnswer := msg.Answers[0]
	var totalPool, winnerPool uint64
	for _, p := range participants {
		totalPool += p.Amount
		if p.Answer == winningAnswer {
			winnerPool += p.Amount
		}
	}
	companyPercent := uint64(1)
	companyFee := totalPool * companyPercent / 100
	rewardPool := totalPool - companyFee

	balancesAfter := make([]uint64, len(participants))
	for i, p := range participants {
		bal, err := queryBalanceViaRPC(ctx, rpcClient, p.Address, "ubet")
		require.NoError(t, err)
		balancesAfter[i] = bal
	}
	for i, p := range participants {
		before, after := balancesBefore[i], balancesAfter[i]
		if p.Answer == winningAnswer {
			expectedReward := uint64(0)
			if winnerPool > 0 {
				expectedReward = rewardPool * p.Amount / winnerPool
			}
			require.Equal(t, before+expectedReward, after,
				"participant %d (winner) before=%d after=%d expectedReward=%d", i, before, after, expectedReward)
		} else {
			require.Equal(t, before, after, "participant %d (loser) balance unchanged", i)
		}
	}
}
