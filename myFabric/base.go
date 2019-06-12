/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/base64"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	fabAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go/build"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"github.com/op/go-logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
	"github.com/golang/protobuf/proto"
)

// BaseSetupImpl implementation of BaseTestSetup
type BaseSetupImpl struct {
	Identity          msp.Identity
	Targets           []string
	ConfigFile        string
	OrgID             string
	ChannelID         string
	ChannelConfigFile string

	ledgerClient      *ledger.Client
	chClient          *channel.Client
	ChainCodeID		  string
	orgChannelClientContext contextAPI.ChannelProvider
}

// Initial B values for ExampleCC
const (
	ExampleCCInitB    = "200"
	ExampleCCUpgradeB = "400"
	keyExp            = "key-%s-%s"
)

// ExampleCC query and transaction arguments
var defaultQueryArgs = [][]byte{[]byte("query"), []byte("b")}
var defaultTxArgs = [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}

// ExampleCC init and upgrade args
var initArgs = [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte(ExampleCCInitB)}
var upgradeArgs = [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte(ExampleCCUpgradeB)}
var resetArgs = [][]byte{[]byte("a"), []byte("100"), []byte("b"), []byte(ExampleCCInitB)}

var logger  *logging.Logger

//initConfig initializes viper config
func InitConfig() error {
	// viper init

	viper.AddConfigPath("./")
	viper.SetConfigName("core")

	viper.SetEnvPrefix(PROJECT_NAME)
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("fatal error config file: %s ", err)
	}
	return nil
}


// CCDefaultQueryArgs returns  cc query args
func CCDefaultQueryArgs() [][]byte {
	return defaultQueryArgs
}

// CCQueryArgs returns  cc query args
func CCQueryArgs(key string) [][]byte {
	return [][]byte{[]byte("query"), []byte(key)}
}

// CCTxArgs returns  cc query args
func CCTxArgs(from, to, val string) [][]byte {
	return [][]byte{[]byte("move"), []byte(from), []byte(to), []byte(val)}
}

// CCDefaultTxArgs returns  cc move funds args
func CCDefaultTxArgs() [][]byte {
	return defaultTxArgs
}

// CCTxRandomSetArgs returns  cc set args with random key-value pairs
func CCTxRandomSetArgs() [][]byte {
	return [][]byte{[]byte("set"), []byte(GenerateRandomID()), []byte(GenerateRandomID())}
}

//CCTxSetArgs sets the given key value in cc
func CCTxSetArgs(key, value string) [][]byte {
	return [][]byte{[]byte("set"), []byte(key), []byte(value)}
}

//CCInitArgs returns  cc initialization args
func CCInitArgs() [][]byte {
	return initArgs
}

//CCUpgradeArgs returns  cc upgrade args
func CCUpgradeArgs() [][]byte {
	return upgradeArgs
}

// IsJoinedChannel returns true if the given peer has joined the given channel
func IsJoinedChannel(channelID string, resMgmtClient *resmgmt.Client, peer fabAPI.Peer) (bool, error) {
	resp, err := resMgmtClient.QueryChannels(resmgmt.WithTargets(peer))
	if err != nil {
		return false, err
	}
	for _, chInfo := range resp.Channels {
		if chInfo.ChannelId == channelID {
			return true, nil
		}
	}
	return false, nil
}

// Initialize reads configuration from file and sets up client and channel
func (setup *BaseSetupImpl) Initialize(sdk *fabsdk.FabricSDK) error {

	mspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(setup.OrgID))
	adminIdentity, err := mspClient.GetSigningIdentity(GetSDKAdmins()[0])
	if err != nil {
		return errors.WithMessage(err, "failed to get client context")
	}
	setup.Identity = adminIdentity

	var cfgBackends []core.ConfigBackend
	configBackend, err := sdk.Config()
	if err != nil {
		//For some tests SDK may not have backend set, try with config file if backend is missing
		cfgBackends = append(cfgBackends, configBackend)
		if err != nil {
			return errors.Wrapf(err, "failed to get config backend from config: %s", err)
		}
	} else {
		cfgBackends = append(cfgBackends, configBackend)
	}

	targets, err := OrgTargetPeers([]string{setup.OrgID}, cfgBackends...)
	if err != nil {
		return errors.Wrapf(err, "loading target peers from config failed")
	}
	setup.Targets = targets

	r, err := os.Open(setup.ChannelConfigFile)
	if err != nil {
		return errors.Wrapf(err, "opening channel config file failed")
	}
	defer func() {
		if err = r.Close(); err != nil {
			test.Logf("close error %v", err)
		}

	}()

	// Create channel for tests
	req := resmgmt.SaveChannelRequest{ChannelID: setup.ChannelID, ChannelConfig: r, SigningIdentities: []msp.SigningIdentity{adminIdentity}}
	if err = InitializeChannel(sdk, setup.OrgID, req, targets); err != nil {
		return errors.WithMessage(err, "failed to initialize channel")
	}

	return nil
}

// GetDeployPath returns the path to the chaincode fixtures
func GetDeployPath() string {
	//const ccPath = "chaincode"
	//return path.Join(goPath(), "src", Project)
	return  goPath()
}

// GetChannelConfigPath returns the path to the named channel config file
func GetChannelConfigPath(filename string) string {
	return path.Join(goPath(), "src", PROJECT_NAME, GetChannelConfig(), filename)
}

// GetConfigPath returns the path to the named config fixture file
func GetConfigPath(filename string) string {
	const configPath = "fixtures/config"
	return path.Join(goPath(), "src", PROJECT_NAME, configPath, filename)
}

// GetConfigOverridesPath returns the path to the named config override fixture file
func GetConfigOverridesPath(filename string) string {
	const configPath = "fixtures/config"
	return path.Join(goPath(), "src", PROJECT_NAME, configPath, "overrides", filename)
}

// GetCryptoConfigPath returns the path to the named crypto-config override fixture file
func GetCryptoConfigPath(filename string) string {
	const configPath = "fixtures/fabric/v1/crypto-config"
	return path.Join(goPath(), "src", PROJECT_NAME, configPath, filename)
}

// goPath returns the current GOPATH. If the system
// has multiple GOPATHs then the first is used.
func goPath() string {
	gpDefault := build.Default.GOPATH
	gps := filepath.SplitList(gpDefault)

	return gps[0]
}

// OrgContext provides SDK client context for a given org
type OrgContext struct {
	OrgID                string
	CtxProvider          contextAPI.ClientProvider
	SigningIdentity      msp.SigningIdentity
	ResMgmt              *resmgmt.Client
	Peers                []fabAPI.Peer
	AnchorPeerConfigFile string
}

// CreateChannelAndUpdateAnchorPeers creates the channel and updates all of the anchor peers for all orgs
func CreateChannelAndUpdateAnchorPeers(t *testing.T, sdk *fabsdk.FabricSDK, channelID string, channelConfigFile string, orgsContext []*OrgContext) error {
	ordererCtx := sdk.Context(fabsdk.WithUser(GetSDKAdmins()[0]), fabsdk.WithOrg(GetSDKOrders()[0]))

	// Channel management client is responsible for managing channels (create/update channel)
	chMgmtClient, err := resmgmt.New(ordererCtx)
	if err != nil {
		return errors.New("failed to get a new resmgmt client for orderer")
	}

	var lastConfigBlock uint64
	var signingIdentities []msp.SigningIdentity
	for _, orgCtx := range orgsContext {
		signingIdentities = append(signingIdentities, orgCtx.SigningIdentity)
	}

	req := resmgmt.SaveChannelRequest{
		ChannelID:         channelID,
		ChannelConfigPath: GetChannelConfigPath(channelConfigFile),
		SigningIdentities: signingIdentities,
	}
	_, err = chMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	if err != nil {
		return err
	}

	lastConfigBlock = WaitForOrdererConfigUpdate(t, orgsContext[0].ResMgmt, channelID, true, lastConfigBlock)

	for _, orgCtx := range orgsContext {
		req := resmgmt.SaveChannelRequest{
			ChannelID:         channelID,
			ChannelConfigPath: GetChannelConfigPath(orgCtx.AnchorPeerConfigFile),
			SigningIdentities: []msp.SigningIdentity{orgCtx.SigningIdentity},
		}
		if _, err := orgCtx.ResMgmt.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com")); err != nil {
			return err
		}

		lastConfigBlock = WaitForOrdererConfigUpdate(t, orgCtx.ResMgmt, channelID, false, lastConfigBlock)
	}

	return nil
}

// JoinPeersToChannel joins all peers in all of the given orgs to the given channel
func JoinPeersToChannel(channelID string, orgsContext []*OrgContext) error {
	for _, orgCtx := range orgsContext {
		err := orgCtx.ResMgmt.JoinChannel(
			channelID,
			resmgmt.WithRetry(retry.DefaultResMgmtOpts),
			resmgmt.WithOrdererEndpoint("orderer.example.com"),
			resmgmt.WithTargets(orgCtx.Peers...),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to join peers in org [%s] to channel [%s]", orgCtx.OrgID, channelID)
		}
	}
	return nil
}

// InstallChaincodeWithOrgContexts installs the given chaincode to orgs
func InstallChaincodeWithOrgContexts(orgs []*OrgContext, ccPkg *resource.CCPackage, ccPath, ccID, ccVersion string) error {
	for _, orgCtx := range orgs {
		if err := InstallChaincode(orgCtx.ResMgmt, ccPkg, ccPath, ccID, ccVersion, orgCtx.Peers); err != nil {
			return errors.Wrapf(err, "failed to install chaincode to peers in org [%s]", orgCtx.OrgID)
		}
	}

	return nil
}

// InstallChaincode installs the given chaincode to the given peers
func InstallChaincode(resMgmt *resmgmt.Client, ccPkg *resource.CCPackage, ccPath, ccName, ccVersion string, localPeers []fabAPI.Peer) error {
	installCCReq := resmgmt.InstallCCRequest{Name: ccName, Path: ccPath, Version: ccVersion, Package: ccPkg}
	_, err := resMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return err
	}

	installed, err := queryInstalledCC(resMgmt, ccName, ccVersion, localPeers)

	if err != nil {
		return err
	}

	if !installed {
		return errors.New("chaincode was not installed on all peers")
	}

	return nil
}

// InstantiateChaincode instantiates the given chaincode to the given channel
func InstantiateChaincode(resMgmt *resmgmt.Client, channelID, ccName, ccPath, ccVersion string, ccPolicyStr string, args [][]byte, collConfigs ...*cb.CollectionConfig) (resmgmt.InstantiateCCResponse, error) {
	ccPolicy, err := cauthdsl.FromString(ccPolicyStr)
	if err != nil {
		return resmgmt.InstantiateCCResponse{}, errors.Wrapf(err, "error creating CC policy [%s]", ccPolicyStr)
	}

	return resMgmt.InstantiateCC(
		channelID,
		resmgmt.InstantiateCCRequest{
			Name:       ccName,
			Path:       ccPath,
			Version:    ccVersion,
			Args:       args,
			Policy:     ccPolicy,
			CollConfig: collConfigs,
		},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	)
}

// UpgradeChaincode upgrades the given chaincode on the given channel
func UpgradeChaincode(resMgmt *resmgmt.Client, channelID, ccName, ccPath, ccVersion string, ccPolicyStr string, args [][]byte, collConfigs ...*cb.CollectionConfig) (resmgmt.UpgradeCCResponse, error) {
	ccPolicy, err := cauthdsl.FromString(ccPolicyStr)
	if err != nil {
		return resmgmt.UpgradeCCResponse{}, errors.Wrapf(err, "error creating CC policy [%s]", ccPolicyStr)
	}

	return resMgmt.UpgradeCC(
		channelID,
		resmgmt.UpgradeCCRequest{
			Name:       ccName,
			Path:       ccPath,
			Version:    ccVersion,
			Args:       args,
			Policy:     ccPolicy,
			CollConfig: collConfigs,
		},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	)
}

// DiscoverLocalPeers queries the local peers for the given MSP context and returns all of the peers. If
// the number of peers does not match the expected number then an error is returned.
func DiscoverLocalPeers(ctxProvider contextAPI.ClientProvider, expectedPeers int) ([]fabAPI.Peer, error) {
	ctx, err := contextImpl.NewLocal(ctxProvider)
	if err != nil {
		return nil, errors.Wrap(err, "error creating local context")
	}

	discoveredPeers, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			peers, serviceErr := ctx.LocalDiscoveryService().GetPeers()
			if serviceErr != nil {
				return nil, errors.Wrapf(serviceErr, "error getting peers for MSP [%s]", ctx.Identifier().MSPID)
			}
			if len(peers) < expectedPeers {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Expecting %d peers but got %d", expectedPeers, len(peers)), nil)
			}
			return peers, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return discoveredPeers.([]fabAPI.Peer), nil
}

// EnsureChannelCreatedAndPeersJoined creates a channel, joins all peers in the given orgs to the channel and updates the anchor peers of each org.
func EnsureChannelCreatedAndPeersJoined(t *testing.T, sdk *fabsdk.FabricSDK, channelID string, channelTxFile string, orgsContext []*OrgContext) error {
	joined, err := IsJoinedChannel(channelID, orgsContext[0].ResMgmt, orgsContext[0].Peers[0])
	if err != nil {
		return err
	}

	if joined {
		return nil
	}

	// Create the channel and update anchor peers for all orgs
	if err := CreateChannelAndUpdateAnchorPeers(t, sdk, channelID, channelTxFile, orgsContext); err != nil {
		return err
	}

	return JoinPeersToChannel(channelID, orgsContext)
}

// WaitForOrdererConfigUpdate waits until the config block update has been committed.
// In Fabric 1.0 there is a bug that panics the orderer if more than one config update is added to the same block.
// This function may be invoked after each config update as a workaround.
func WaitForOrdererConfigUpdate(t *testing.T, client *resmgmt.Client, channelID string, genesis bool, lastConfigBlock uint64) uint64 {

	blockNum, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			chConfig, err := client.QueryConfigFromOrderer(channelID, resmgmt.WithOrdererEndpoint("orderer.example.com"))
			if err != nil {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), err.Error(), nil)
			}

			currentBlock := chConfig.BlockNumber()
			if currentBlock <= lastConfigBlock && !genesis {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Block number was not incremented [%d, %d]", currentBlock, lastConfigBlock), nil)
			}
			return &currentBlock, nil
		},
	)

	require.NoError(t, err)
	return *blockNum.(*uint64)
}

func queryInstalledCC(resMgmt *resmgmt.Client, ccName, ccVersion string, peers []fabAPI.Peer) (bool, error) {
	installed, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			ok, err := isCCInstalled(resMgmt, ccName, ccVersion, peers)
			if err != nil {
				return &ok, err
			}
			if !ok {
				return &ok, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Chaincode [%s:%s] is not installed on all peers in Org1", ccName, ccVersion), nil)
			}
			return &ok, nil
		},
	)

	if err != nil {
		s, ok := status.FromError(err)
		if ok && s.Code == status.GenericTransient.ToInt32() {
			return false, nil
		}
		return false, errors.WithMessage(err, "isCCInstalled invocation failed")
	}

	return *(installed).(*bool), nil
}

func isCCInstalled(resMgmt *resmgmt.Client, ccName, ccVersion string, peers []fabAPI.Peer) (bool, error) {
	installedOnAllPeers := true
	for _, peer := range peers {
		resp, err := resMgmt.QueryInstalledChaincodes(resmgmt.WithTargets(peer))
		if err != nil {
			return false, errors.WithMessage(err, "querying for installed chaincodes failed")
		}

		found := false
		for _, ccInfo := range resp.Chaincodes {
			if ccInfo.Name == ccName && ccInfo.Version == ccVersion {
				found = true
				break
			}
		}
		if !found {
			installedOnAllPeers = false
		}
	}
	return installedOnAllPeers, nil
}

//GetKeyName creates random key name based on test name
func GetKeyName(t *testing.T) string {
	return fmt.Sprintf(keyExp, t.Name(), GenerateRandomID())
}

//ResetKeys resets given set of keys in example cc to given value
func ResetKeys(t *testing.T, ctx contextAPI.ChannelProvider, chaincodeID, value string, keys ...string) {
	chClient, err := channel.New(ctx)
	require.NoError(t, err, "Failed to create new channel client for resetting keys")
	for _, key := range keys {
		// Synchronous transaction
		_, e := chClient.Execute(
			channel.Request{
				ChaincodeID: chaincodeID,
				Fcn:         "invoke",
				Args:        CCTxSetArgs(key, value),
			},
			channel.WithRetry(retry.DefaultChannelOpts))
		require.NoError(t, e, "Failed to reset keys")
	}
}





//###################################################################################################################


//ResetKeys resets given set of keys in  cc to given value
func SetKeyData(ctx contextAPI.ChannelProvider, chaincodeID, value string, key string) fabAPI.TransactionID{
	chClient, err := channel.New(ctx)
	if err != nil {print(err)}

	// Synchronous transaction
	respone, e := chClient.Execute(
		channel.Request{
			ChaincodeID: chaincodeID,
			Fcn:         "invoke",
			Args:        CCTxSetArgs(key, value),
		},
		channel.WithRetry(retry.DefaultChannelOpts))

	if e != nil {print(e)}

	return respone.TransactionID
}


func GetValueFromKey(chClient *channel.Client,ccID, key string) string{

	const (
		maxRetries = 10
		retrySleep = 500 * time.Millisecond
	)

	for r := 0; r < maxRetries; r++ {
		response, err := chClient.Query(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: CCQueryArgs(key)},
			channel.WithRetry(retry.DefaultChannelOpts))
		if err == nil {
			actual := string(response.Payload)
			if actual != "" {
				return actual
			}
		}

		time.Sleep(retrySleep)
	}

	return ""
}

type BlockchainInfo struct {
	Height            uint64 `json:"height"`
	CurrentBlockHash  string `json:"currentBlockHash"`
	PreviousBlockHash string `json:"previousBlockHash"`
}

// GetBlockchainInfo ...
func (setup *BaseSetupImpl) GetBlockchainInfo() (BlockchainInfo, error) {
	var blockchainInfo BlockchainInfo
	b, err := setup.ledgerClient.QueryInfo()
	if err != nil {
		return blockchainInfo, err
	}
	blockchainInfo.Height = b.BCI.Height
	blockchainInfo.CurrentBlockHash = base64.StdEncoding.EncodeToString(b.BCI.CurrentBlockHash)
	blockchainInfo.PreviousBlockHash = base64.StdEncoding.EncodeToString(b.BCI.PreviousBlockHash)

	return blockchainInfo, nil
}

type Block struct {
	Version           int           `protobuf:"bytes,1,opt,name=version" json:"version"`
	Timestamp         string        `protobuf:"bytes,2,opt,name=timestamp" json:"timestamp"`
	Transactions      []Transaction `protobuf:"bytes,3,opt,name=transactions" json:"transactions"`
	StateHash         string        `protobuf:"bytes,4,opt,name=stateHash" json:"stateHash"`
	PreviousBlockHash string        `protobuf:"bytes,5,opt,name=previousBlockHash" json:"previousBlockHash"`
	// NonHashData       LocalLedgerCommitTimestamp `protobuf:"bytes,6,opt,name=nonHashData" json:"nonHashData,omitempty"`
}

type Transaction struct {
	Type        int32                     `protobuf:"bytes,1,opt,name=type" json:"type"`
	ChaincodeID string                    `protobuf:"bytes,2,opt,name=chaincodeID" json:"chaincodeID"`
	Payload     string                    `protobuf:"bytes,3,opt,name=payload" json:"payload"`
	UUID        string                    `protobuf:"bytes,4,opt,name=uuid" json:"uuid"`
	Timestamp   google_protobuf.Timestamp `protobuf:"bytes,5,opt,name=timestamp" json:"timestamp"`
	Cert        string                    `protobuf:"bytes,6,opt,name=cert" json:"cert"`
	Signature   string                    `protobuf:"bytes,7,opt,name=signature" json:"signature"`
}

// GetBlock ...
func (setup *BaseSetupImpl) GetBlock(num uint64) (Block, error) {
	var block Block
	block.Transactions = make([]Transaction, 0)

	b, err := setup.ledgerClient.QueryBlock(num)
	if err != nil {
		return block, nil
	}

	block.StateHash = base64.StdEncoding.EncodeToString(b.Header.DataHash)
	block.PreviousBlockHash = base64.StdEncoding.EncodeToString(b.Header.PreviousHash)

	for i := 0; i < len(b.Data.Data); i++ {
		// parse block
		envelope, err := utils.ExtractEnvelope(b, i)

		if err != nil {
			return block, err
		}

		payload, err := utils.ExtractPayload(envelope)
		if err != nil {
			return block, err
		}

		channelHeader, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
		if err != nil {
			return block, err
		}

		block.Version = int(cb.HeaderType(channelHeader.Type))
		switch cb.HeaderType(channelHeader.Type) {
		case cb.HeaderType_MESSAGE:
			break
		case cb.HeaderType_CONFIG:
			configEnvelope := &cb.ConfigEnvelope{}
			if err := proto.Unmarshal(payload.Data, configEnvelope); err != nil {
				return block, err
			}
			break
		case cb.HeaderType_CONFIG_UPDATE:
			configUpdateEnvelope := &cb.ConfigUpdateEnvelope{}
			if err := proto.Unmarshal(payload.Data, configUpdateEnvelope); err != nil {
				return block, err
			}
			break
		case cb.HeaderType_ENDORSER_TRANSACTION:
			tx, err := utils.GetTransaction(payload.Data)
			if err != nil {
				return block, err
			}

			channelHeader := &cb.ChannelHeader{}
			if err := proto.Unmarshal(payload.Header.ChannelHeader, channelHeader); err != nil {
				return block, err
			}

			signatureHeader := &cb.SignatureHeader{}
			if err := proto.Unmarshal(payload.Header.SignatureHeader, signatureHeader); err != nil {
				return block, err
			}

			for _, action := range tx.Actions {
				var transaction Transaction
				transaction.ChaincodeID = setup.ChainCodeID
				transaction.Payload = base64.StdEncoding.EncodeToString(action.Payload)

				transaction.Type = channelHeader.Type
				transaction.UUID = channelHeader.TxId
				transaction.Cert = base64.StdEncoding.EncodeToString(signatureHeader.Creator)
				transaction.Signature = base64.StdEncoding.EncodeToString(envelope.Signature)
				if channelHeader != nil && channelHeader.Timestamp != nil {
					transaction.Timestamp.Seconds = channelHeader.Timestamp.Seconds
					transaction.Timestamp.Nanos = channelHeader.Timestamp.Nanos
				}

				block.Transactions = append(block.Transactions, transaction)
			}
			break
		case cb.HeaderType_ORDERER_TRANSACTION:
			break
		case cb.HeaderType_DELIVER_SEEK_INFO:
			break
		default:
			return block, fmt.Errorf("Unknown message")
		}
	}
	return block, nil
}