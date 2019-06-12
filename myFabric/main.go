package main

import (
	"encoding/json"
	"github.com/spf13/viper"

	//"encoding/json"
	//"errors"
	"fmt"
	"github.com/gocraft/web"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"net/http"

	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
)

const (
	PROJECT_NAME = "myFabric"
)

var(

	mainSDK *fabsdk.FabricSDK
	mainSetup *BaseSetupImpl
	mainChaincodeID string
	org1ChannelClientContext contextApi.ChannelProvider
	chClient *channel.Client
	err error
	defSetup *BaseSetupImpl
)


// Status REST response
type Status struct {
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
}

// Blockchain ...
type Blockchain struct {
	BSI *BaseSetupImpl
}


// setResponseType is a middleware function that sets the appropriate response
// headers. Currently, it is setting the "Content-Type" to "application/json" as
// well as the necessary headers in order to enable CORS for Swagger usage.
func (b *Blockchain) setResponseTypeBSI(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	rw.Header().Set("Content-Type", "application/json")

	// Enable CORS
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "accept, content-type")

	b.BSI = defSetup
	next(rw, req)
}

// NotFound returns a custom landing page when a given hyperledger end point
// had not been defined.
func (b *Blockchain) notFound(rw web.ResponseWriter, req *web.Request){

	rw.WriteHeader(http.StatusNotFound)
	err := json.NewEncoder(rw).Encode(Status{Code: http.StatusNotFound, Message: "Insurance endpoint not found."})

	if err != nil {
		panic(fmt.Errorf("error find when do notFound : %s", err))
	}

}


func buildRouter() *web.Router {

	router := web.New(Blockchain{})
	router.Middleware((*Blockchain).setResponseTypeBSI)
	router.NotFound((*Blockchain).notFound)

	// 表示所有的请求的前缀都有/chain
	chainRouter := router.Subrouter(Chain{}, "/chain")

	//获取具体链的信息http://localhost:5984/chain
	chainRouter.Get("", (*Chain).chain)

	//获取具体某个区块的信息http://localhost:5984/chain/blocks/5
	chainRouter.Get("/blocks/:num", (*Chain).block)

	/*
	chainRouter.Get("/transactions/:txid", (*Chain).tx)
	chainRouter.Post("/invoke", (*Chain).invoke)
	chainRouter.Post("/query", (*Chain).query)
	chainRouter.Get("/queryWithBlock/:key", (*Chain).queryWithBlock)
	*/

	return router
}

func main(){

	fmt.Println("come to main")

	//init core.yml  加载配置文件
	if err = InitConfig(); err != nil {
		// Handle errors reading the config file
		panic(fmt.Errorf("fatal error when initializing config : %s", err))
	}


	/*
	//init
	r := NewWithCC()
	r.Initialize()
	mainSDK = r.GetSDK()
	mainSetup = r.GetSetup()
	mainChaincodeID = r.GetChaincodeID()

	fmt.Println("finish init")


	//set logic key
	aKey := "keyA"
	//bKey := "keyB"

	//prepare context
	org1ChannelClientContext = mainSDK.ChannelContext(mainSetup.ChannelID, fabsdk.WithUser(r.Org1User), fabsdk.WithOrg(r.Org1Name))

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

	*/




	//init
	r := NewWithCC()
	r.Initialize()
	fmt.Printf("finish init %s \n",PROJECT_NAME)

	//set the channel context
	defSetup = &BaseSetupImpl{}

	mainSDK = r.GetSDK()
	mainSetup = r.GetSetup()
	defSetup.ChainCodeID = r.GetChaincodeID()
	defSetup.ChannelID = mainSetup.ChannelID
	defSetup.orgChannelClientContext = mainSDK.ChannelContext(mainSetup.ChannelID, fabsdk.WithUser(r.Org1User),fabsdk.WithOrg(r.Org1Name))
	defSetup.chClient, err = channel.New(defSetup.orgChannelClientContext)
	if err != nil {
		logger.Errorf("channel.New: %s", err)
	}

	defSetup.ledgerClient, err = ledger.New(defSetup.orgChannelClientContext)
	if err != nil {
		logger.Errorf("ledger.New: %s", err)
	}

	//defSetup.ledgerClient = ledgerClient



	// RESTful 接口服务  使用了gocraft/web包实现的，地址是https://github.com/gocraft/web，若本地没有这个包，要先下载。
	// 后期打算用beego来替代这个服务
	//router
	router := buildRouter()
	// 接口服务的地址是写在配置文件中的。
	restAddress := viper.GetString("spi.rest.address")

	fmt.Printf("listen at the address : %s \n",restAddress)
	err = http.ListenAndServe(restAddress, router)
	if err != nil {
		logger.Errorf("ListenAndServe: %s", err)
	}


}

