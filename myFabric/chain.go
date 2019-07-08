package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocraft/web"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/protos/ledger/queryresult"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)
// Chain 区块链查询展示
type Chain struct {
	*Blockchain
}


// chain ...
func (c *Chain) chain(rw web.ResponseWriter, req *web.Request) {

	fmt.Println("come to chain")
	encoder := json.NewEncoder(rw)

	chain, err := c.BSI.GetBlockchainInfo()
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		logger.Errorf("query chaininfo failed: %s", err)
		return
	}

	rw.WriteHeader(http.StatusOK)
	err = encoder.Encode(Status{Code: http.StatusOK, Message: chain})
	if err != nil{
		logger.Errorf("invoke return failed: %s", err)
	}
}


// block
func (c *Chain) block(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	blockNumber, err := strconv.Atoi(req.PathParams["num"])
	if err != nil {
		// Failure
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	block, err := c.BSI.GetBlock(uint64(blockNumber))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		logger.Errorf("QueryBlock error:%v", err)
		return
	}

	rw.WriteHeader(http.StatusOK)
	err = encoder.Encode(Status{Code: http.StatusOK, Message: block})
	if err != nil{
		logger.Errorf("invoke return failed: %s", err)
	}
}


// tx ...
func (c *Chain) tx(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	txid := req.PathParams["txid"]
	tx, err := c.BSI.GetTransaction(txid)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		logger.Errorf("QueryBlock error:%v", err)
		return
	}

	rw.WriteHeader(http.StatusOK)
	err = encoder.Encode(Status{Code: http.StatusOK, Message: tx})
	if err != nil{
		logger.Errorf("invoke return failed: %s", err)
	}
}

func (c *Chain) txmy(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	txid := req.PathParams["txid"]
	tx, err := c.BSI.GetTransaction(txid)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		logger.Errorf("QueryBlock error:%v", err)
		return
	}

	rw.WriteHeader(http.StatusOK)
	err = encoder.Encode(Status{Code: http.StatusOK, Message: tx})
	if err != nil{
		logger.Errorf("invoke return failed: %s", err)
	}
}


// invoke ...
func (c *Chain) invoke(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		logger.Error("Internal JSON error when reading request body.")
		return
	}

	// Incoming request body may not be empty, client must supply request payload
	if string(reqBody) == "" {
		rw.WriteHeader(http.StatusBadRequest)
		logger.Error("Client must supply a payload for order requests.")
		return
	}
	fmt.Printf("Req body: %s \n", string(reqBody))

	var args struct {
		Args []string `json:"args"`
	}
	err = json.Unmarshal(reqBody, &args)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		logger.Errorf("Error unmarshalling request payload: %s", err)
		return
	}

	txid, err := c.BSI.Invoke(args.Args[0], args.Args[1])

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		logger.Errorf("invoke chaincode failed: %s", err)
		return
	}
	fmt.Printf("response txID: %s", txid)

	rw.WriteHeader(http.StatusOK)
	err = encoder.Encode(Status{Code: http.StatusOK, Message: txid})
	if err != nil{
		logger.Errorf("invoke return failed: %s", err)
	}
}

// query ...
func (c *Chain) query(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		logger.Error("Internal JSON error when reading request body.")
		return
	}

	// Incoming request body may not be empty, client must supply request payload
	if string(reqBody) == "" {
		rw.WriteHeader(http.StatusBadRequest)
		logger.Error("Client must supply a payload for order requests.")
		return
	}
	fmt.Printf("Req body: %s", string(reqBody))

	var args struct {
		Args []string `json:"args"`
	}
	err = json.Unmarshal(reqBody, &args)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		logger.Errorf("Error unmarshalling request payload: %s", err)
		return
	}

	result, err := c.BSI.Find("query",args.Args[0])
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		logger.Errorf("invoke chaincode failed: %s", err)
		return
	}

	rw.WriteHeader(http.StatusOK)
	err = encoder.Encode(Status{Code: http.StatusOK, Message: result})
	if err != nil{
		logger.Errorf("query return failed: %s", err)
	}
}

// History worldState history
type History struct {
	*queryresult.KeyModification `json:"history"`
	Value                        string `json:"value"`
	BlockNum                     uint64 `json:"blockNum"`
}


// QueryResult chaincode query result
type QueryResult struct {
	Current struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"current"`
	Historys []History `json:"historys"`
}

// queryWithBlock ...
func (c *Chain) queryWithBlock(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	key := req.PathParams["key"]
	result, err := c.BSI.Find("history",key)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		logger.Errorf("query chaincode failed: %s", err)
		return
	}

	query := QueryResult{}

	err = json.Unmarshal([]byte(result), &query)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		logger.Errorf("query chaincode failed: %s", err)
		return
	}

	for k, v := range query.Historys {
		query.Historys[k].Value = string(v.GetValue())

		// 在Channel接口中添加QueryBlockByTxID方法并实现
		block, err := c.BSI.ledgerClient.QueryBlockByTxID(fab.TransactionID(v.GetTxId()))
		if err != nil {
			logger.Errorf("QueryBlockByTxID error:%v", err)
			continue
		}

		query.Historys[k].BlockNum = block.Header.Number
	}

	rw.WriteHeader(http.StatusOK)
	err = encoder.Encode(Status{Code: http.StatusOK, Message: query})
	if err != nil{
		logger.Errorf("query return failed: %s", err)
	}
}

func (c *Chain) rangeQuery(rw web.ResponseWriter, req *web.Request) {

	startKey := req.PathParams["startKey"]
	endKey := req.PathParams["endKey"]

	result := c.BSI.RangeQuery(startKey,endKey)

	fmt.Println(startKey,endKey,result)

	encoder := json.NewEncoder(rw)

	rw.WriteHeader(http.StatusOK)
	err = encoder.Encode(Status{Code: http.StatusOK, Message: result})
	if err != nil{
		logger.Errorf("invoke return failed: %s", err)
	}
}

func (c *Chain)queryWithColor(rw web.ResponseWriter, req *web.Request)   {

	color := req.PathParams["color"]
	result := c.BSI.queryWithColor(color,"color~id")

	var carResult []Car = []Car{}
 	resultArr := strings.Split(result,"_")

	for i :=0 ; i<len(resultArr); i++ {
		mycar := Car{}
		err := json.Unmarshal([]byte(resultArr[i]),&mycar)

		if err != nil {
			logger.Errorf("invoke return failed: %s", err)
			panic("Unmarshal fail")
		}
		carResult = append(carResult,mycar)
	}

	encoder := json.NewEncoder(rw)

	rw.WriteHeader(http.StatusOK)
	err = encoder.Encode(Status{Code: http.StatusOK, Message: carResult})
	if err != nil{
		logger.Errorf("invoke return failed: %s", err)
	}
}

func (c *Chain)querySqlWithColor(rw web.ResponseWriter, req *web.Request)   {
	color := req.PathParams["color"]
	colorSql := fmt.Sprintf(`{"selector":{"Color":"%s"}}`,color)
	fmt.Println("------------sql:",colorSql)

	result := c.BSI.querySqlWithColor(colorSql)

	fmt.Println("------------结果:",result)


	type ResultModel struct {
		Key    string
		Record Car
	}

	type sqlQueryResult struct {
		Results []ResultModel `json:"result"`
	}

	var sqlQueryResultModel = sqlQueryResult{}
	err := json.Unmarshal([]byte(result),&sqlQueryResultModel)
	if err != nil {
		logger.Errorf("Unmarshal fail: %s", err)
		panic("Unmarshal fail")
	}

	var carResult []Car = []Car{}
	for i :=0 ; i<len(sqlQueryResultModel.Results); i++ {
		resultModel := sqlQueryResultModel.Results[i];

		if err != nil {
			logger.Errorf("invoke return failed: %s", err)
			panic("Unmarshal fail")
		}
		carResult = append(carResult,resultModel.Record)
	}

	encoder := json.NewEncoder(rw)

	rw.WriteHeader(http.StatusOK)
	err = encoder.Encode(Status{Code: http.StatusOK, Message: carResult})
	if err != nil{
		logger.Errorf("invoke return failed: %s", err)
	}
}

