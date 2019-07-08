package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/spf13/viper"
	//"encoding/json"
	//"errors"
	"fmt"
	"github.com/gocraft/web"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"net/http"

	//"io/ioutil"
	//"net/http"
	//
	//"github.com/gocraft/web"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
)


const (
	PROJECT_NAME = "prj2"
)


var(

	mainSDK *fabsdk.FabricSDK
	mainSetup *BaseSetupImpl
	mainChaincodeID string
	orgChannelClientContext contextApi.ChannelProvider
	chClient *channel.Client
	err error

	defSetup *BaseSetupImpl

)

// Status REST response
type Status struct {
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
}

type Car struct {
	Color      string `json:"Color"`
	ID         string `json:"ID"`
	Price      string `json:"Price"`
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

func (b *Blockchain)addTestData(rw web.ResponseWriter, req *web.Request)  {
	//初始化一些测试数据
	encoder := json.NewEncoder(rw)
	var resultArr []map[string]interface{} = []map[string]interface{}{}
	testArr := createData()
	for i:=0;i<len(testArr) ;i++  {
		mycar := testArr[i]
		carKey := mycar.ID
		b,_  := json.Marshal(mycar)
		txid := SetKeyData(defSetup.orgChannelClientContext, defSetup.ChainCodeID, carKey,string(b),)
		//加入color~id的索引
		defSetup.addIndexes("color~id",mycar.Color,mycar.ID)
		fmt.Printf("key is %s,txid is %s \n",carKey,txid)
		if txid == ""{
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		var txDic map[string]interface{} = map[string]interface{}{carKey:testArr[i],"txId":string(txid)}
		resultArr = append(resultArr,txDic)
	}



	rw.WriteHeader(http.StatusOK)
	err = encoder.Encode(Status{Code: http.StatusOK, Message: resultArr})
	if err != nil{
		logger.Errorf("invoke return failed: %s", err)
	}
}


func buildRouter() *web.Router {

	router := web.New(Blockchain{})
	router.Middleware((*Blockchain).setResponseTypeBSI)
	router.NotFound((*Blockchain).notFound)

	//加载test数据
	router.Get("/addTestData",(*Blockchain).addTestData)


	chainRouter := router.Subrouter(Chain{}, "/chain")

	chainRouter.Get("", (*Chain).chain)

	chainRouter.Get("/blocks/:num", (*Chain).block)

	chainRouter.Get("/transactions/:txid", (*Chain).tx)

	//另一种方式获取到交易信息,和上一条一样的功能
	chainRouter.Get("/transactionsmy/:txid",(*Chain).txmy)

	//获取从startKey到endKey中间的数据，[startKey,endKey)  包含开始键、不包括结束键的半闭半开区间
	chainRouter.Get("/rangeQuery/:startKey/:endKey",(*Chain).rangeQuery)

	//查询某种颜色的汽车
	chainRouter.Get("/queryWithColor/:color",(*Chain).queryWithColor)

	//couchDB查询语句查询某种颜色的汽车
	chainRouter.Get("/querySqlWithColor/:color",(*Chain).querySqlWithColor)

	chainRouter.Post("/invoke", (*Chain).invoke)

	chainRouter.Post("/query", (*Chain).query)

	chainRouter.Get("/queryWithBlock/:key", (*Chain).queryWithBlock)

	return router
}

func createData() []Car {
	// 创建初始化模拟数据

	var car1 Car = Car{Color:"red",ID:"123456",Price:"100.00"}
	var car2 Car = Car{Color:"blue",ID:"234567",Price:"200.00"}
	var car3 Car = Car{Color:"green",ID:"345678",Price:"300.00"}
	var car4 Car = Car{Color:"red",ID:"456789",Price:"400.00"}

	var test []Car = []Car{car1,car2,car3,car4}
	return test

}

func main(){

	fmt.Println("come to main")


	//init core.yml
	if err = InitConfig(); err != nil {
		// Handle errors reading the config file
		panic(fmt.Errorf("fatal error when initializing config : %s", err))
	}

	/*
	r := NewWithCC()
	r.Initialize()
	mainSDK = r.GetSDK()
	mainSetup = r.GetSetup()
	mainChaincodeID = r.GetChaincodeID()

	fmt.Println("finish init")


	//set logic key
	aKey := "AkeyA"
	//bKey := "keyB"

	//prepare context
	orgChannelClientContext = mainSDK.ChannelContext(mainSetup.ChannelID, fabsdk.WithUser(r.Org1User), fabsdk.WithOrg(r.Org1Name))

	//get channel client
	chClient, err = channel.New(orgChannelClientContext)
	if err != nil {panic(err)}

	//ledgerClient, err := ledger.New(org1ChannelClientContext)
	if err != nil {panic(err)}

	//1.set single
	txid := SetKeyData(orgChannelClientContext, mainChaincodeID, aKey,"250",)
	fmt.Printf("key is %s,txid is %s \n",aKey,txid)

	//2.get data from key
	val := GetValueFromKey(chClient,mainChaincodeID,"query",aKey)
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


	//router
	router := buildRouter()
	restAddress := viper.GetString("spi.rest.address")

	fmt.Printf("listen at the address : %s \n",restAddress)
	err = http.ListenAndServe(restAddress, router)
	if err != nil {
		logger.Errorf("ListenAndServe: %s", err)
	}
}


/**

	1.完成多key的联合查询
       GetStateByRange(startKey, endKey)  开始位置，结束位置，获取出中间部分的数据

	2.支持模糊查询功能
		 a、CreateCompositeKey 给定一组属性，将这些属性组合起来构造一个复合键
         b、SplitCompositeKey 给定一个复合键，将其拆分为复合键所用的属性
		 c、GetStateByPartialCompositeKey 根据局部的复合键返回所有的匹配的键值

	3.支持group查询功能

	4.增加缓存功能，提高查询的效率

*/