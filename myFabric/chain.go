package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocraft/web"
	"net/http"
	"strconv"
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


