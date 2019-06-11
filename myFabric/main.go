package main

import (

	//"encoding/json"
	//"errors"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
)

var(

	mainSDK *fabsdk.FabricSDK
	mainTestSetup *BaseSetupImpl
	mainChaincodeID string
	org1ChannelClientContext contextApi.ChannelProvider
	chClient *channel.Client
	err error
)


// Status REST response
type Status struct {
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
}

// Blockchain ...
type Blockchain struct {

}

func main(){

	fmt.Println("come to main")


	//init
	r := NewWithExampleCC()
	r.Initialize()
	mainSDK = r.SDK()
	mainTestSetup = r.TestSetup()
	mainChaincodeID = r.ExampleChaincodeID()

	fmt.Println("finish init")


	//set logic key
	aKey := "keyA"
	//bKey := "keyB"

	//prepare context
	org1ChannelClientContext = mainSDK.ChannelContext(mainTestSetup.ChannelID, fabsdk.WithUser(r.Org1User), fabsdk.WithOrg(r.Org1Name))

	//get channel client
	chClient, err = channel.New(org1ChannelClientContext)
	if err != nil {panic(err)}

	//ledgerClient, err := ledger.New(org1ChannelClientContext)
	if err != nil {panic(err)}

	//1.set single
	txid := SetKeyData(org1ChannelClientContext, mainChaincodeID, "250",aKey)
	fmt.Printf("key is %s,txid is %s \n",aKey,txid)

	//2.get data from key
	val := GetValueFromKey(chClient,mainChaincodeID,aKey)
	fmt.Printf("val is %s \n",val)




}

