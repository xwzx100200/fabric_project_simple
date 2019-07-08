/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric/protos/ledger/queryresult"
)

var logger = shim.NewLogger("cc")

// AgriChaincode example simple Chaincode implementation
type AgriChaincode struct {
}

// Init ...
func (t *AgriChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	txID := stub.GetTxID()
	fmt.Println("fmt.Println_prj2_Init start")
	logger.Debugf("[txID %s] ########### prj2_Init ###########\n", txID)
	funcs, args := stub.GetFunctionAndParameters()

	logger.Debugf("************start Init funcs = [%s] args =[%#v]# ************\n",funcs,args)

	/*err := t.reset(stub, txID, args)
	if err != nil {
		return shim.Error(err.Error())
	}*/

	if transientMap, err := stub.GetTransient(); err == nil {
		if transientData, ok := transientMap["result"]; ok {
			logger.Debugf("[txID %s] Transient data in 'init' : %s\n", txID, transientData)
			return shim.Success(transientData)
		}
	}
	return shim.Success(nil)

}

func (t *AgriChaincode) resetCC(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Deletes an entity from its state
	if err := t.reset(stub, stub.GetTxID(), args); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *AgriChaincode) reset(stub shim.ChaincodeStubInterface, txID string, args []string) error {
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error


	logger.Debugf("************ prj2_reset txID = [%s] args =[%#v] ************\n",txID,args)

	if len(args) != 4 {
		return errors.New("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return errors.New("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return errors.New("Expecting integer value for asset holding")
	}
	logger.Debugf("[txID %s] Aval = %d, Bval = %d\n", txID, Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return err
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return err
	}

	return nil
}

// Query ...
func (t *AgriChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Unknown supported call")
}

//set sets given key-value in state
func (t *AgriChaincode) set(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	logger.Debugf("************ prj2_set  args =[%#v] ************\n",args)
	var err error

	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting a key and a value")
	}

	// Initialize the chaincode
	key := args[1]
	value := args[2]
	eventID := "testEvent"
	if len(args) >= 4 {
		eventID = args[3]
	}

	logger.Debugf("Setting value for key[%s]", key)

	// Write the state to the ledger
	err = stub.PutState(key, []byte(value))
	if err != nil {
		logger.Errorf("Failed to set value for key[%s] : ", key, err)
		return shim.Error(err.Error())
	}

	err = stub.SetEvent(eventID, []byte("Test Payload"))
	if err != nil {
		logger.Errorf("Failed to set event for key[%s] : ", key, err)
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Invoke ...
// Transaction makes payment of X units from A to B
func (t *AgriChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	function, args := stub.GetFunctionAndParameters()

	logger.Debugf("************ prj2_Invoke fun = [%s] args =[%#v] ************\n",function,args)


	if function == "invokecc" {
		return t.invokeCC(stub, args)
	}

	if function == "reset" {
		return t.resetCC(stub, args)
	}

	if function != "invoke" {
		return shim.Error("Unknown function call")
	}

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting at least 2")
	}

	if args[0] == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	}

	if args[0] == "query" {
		// queries an entity state
		return t.query(stub, args)
	}

	if args[0] == "set" {
		// setting an entity state
		return t.set(stub, args)
	}

	if args[0] == "move" {
		eventID := "testEvent"
		if len(args) >= 5 {
			eventID = args[4]
		}
		if err := stub.SetEvent(eventID, []byte("Test Payload")); err != nil {
			return shim.Error("Unable to set CC event: testEvent. Aborting transaction ...")
		}
		return t.move(stub, args)
	}

	if args[0] == "history" {
		//case "history":
		return t.history(stub, args)
	}

	if args[0] == "rangeQuery" {
		fmt.Println("###############rangeQuery start##################")
		return t.rangeQuery(stub, args)
	}

	if args[0] == "richQueryColor"{
		fmt.Println("###############richQueryColor start##################")
		return t.richQueryColor(stub, args)
	}

	if args[0] == "addIndexes" {
		fmt.Println("###############addIndexes start##################")
		return t.addIndexes(stub, args)
	}

	if args[0] == "querySql" {
		fmt.Println("###############querySql##################")
		return t.getQueryResultForQueryStrings(stub, args)
	}


	return shim.Error("Unknown action, check the first argument, must be one of 'delete', 'query', or 'move'")
}

func (t *AgriChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {


	logger.Debugf("************ prj2_move stub =[%#v] args =[%#v] ************\n",stub,args)

	txID := stub.GetTxID()
	// must be an invoke
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error
	if len(args) < 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4, function followed by 2 names and 1 value")
	}

	A = args[1]
	B = args[2]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	logger.Debugf("[txID %s] Aval = %d, Bval = %d\n", txID, Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	if transientMap, err := stub.GetTransient(); err == nil {
		if transientData, ok := transientMap["result"]; ok {
			logger.Debugf("[txID %s] Transient data in 'move' : %s\n", txID, transientData)
			return shim.Success(transientData)
		}
	}
	return shim.Success(nil)
}

// Deletes an entity from state
func (t *AgriChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[1]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// Query callback representing the query of a chaincode
func (t *AgriChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	logger.Debugf("************ prj2_query stub =[%#v] args =[%#v] ************\n",stub,args)

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[1]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	logger.Debugf("[txID %s] Query Response:%s\n", stub.GetTxID(), jsonResp)
	return shim.Success(Avalbytes)
}

type argStruct struct {
	Args []string `json:"Args"`
}

func asBytes(args []string) [][]byte {
	bytes := make([][]byte, len(args))
	for i, arg := range args {
		bytes[i] = []byte(arg)
	}
	return bytes
}

// invokeCC invokes another chaincode
// arg0: ID of chaincode to invoke
// arg1: Chaincode arguments in the form: {"Args": ["arg0", "arg1",...]}
func (t *AgriChaincode) invokeCC(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting ID of chaincode to invoke and args")
	}

	ccID := args[0]
	invokeArgsJSON := args[1]

	argStruct := argStruct{}
	if err := json.Unmarshal([]byte(invokeArgsJSON), &argStruct); err != nil {
		return shim.Error(fmt.Sprintf("Invalid invoke args: %s", err))
	}

	if err := stub.PutState(stub.GetTxID()+"_invokedcc", []byte(ccID)); err != nil {
		return shim.Error(fmt.Sprintf("Error putting state: %s", err))
	}

	return stub.InvokeChaincode(ccID, asBytes(argStruct.Args), "")
}


// history callback representing the query of a chaincode
func (t *AgriChaincode) history(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var key string
	var err error

	logger.Debugf("************ prj2_history args =[%#v] ************\n",args)

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting =1")
	}

	key = args[1]

	type History struct {
		*queryresult.KeyModification `json:"history"`
	}
	result := struct {
		Current struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"current"`
		Historys []History `json:"historys"`
	}{}

	// Get the state from the ledger
	value, err := stub.GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	if value == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + key + "\"}"
		return shim.Error(jsonResp)
	}


	result.Current.Key = key
	result.Current.Value = string(value)

	historyIterator, err := stub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer historyIterator.Close()

	i := 0
	for historyIterator.HasNext() {
		history, err := historyIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		i++
		result.Historys = append(result.Historys, History{history})
	}

	jsonResp, err := json.Marshal(result)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(jsonResp)
}

func (t *AgriChaincode) rangeQuery(stub shim.ChaincodeStubInterface, args []string) pb.Response {


	logger.Debugf("[txID %s] ########### prj2_cc rangeQuery ###########\n", stub.GetTxID())

	startKey := args[1]
	endKey := args[2]
	logger.Debugf("startKey:%s,endKey:%s",startKey,endKey)

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)

	if err != nil {
		return shim.Error("Query by Range failed")
	}
	defer resultsIterator.Close() //释放迭代器

	var buffer bytes.Buffer
	bArrayMemberAlreadyWritten := false
	buffer.WriteString(`{"result":[`)

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next() //获取迭代器中的每一个值
		if err != nil {
			return shim.Error("Fail")
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString(string(queryResponse.Value)) //将查询结果放入Buffer中
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString(`]}`)
	logger.Debugf("Query result: %s", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (t *AgriChaincode) addIndexes(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	indexName := args[1]
	attribute1 := args[2]
	attribute2 := args[3]

	logger.Debugf("stub.CreateCompositeKey result : %s,%s",attribute1,attribute2)

	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{attribute1, attribute2}) //创建Color与ID的组合键

	logger.Debugf("stub.CreateCompositeKey result : %s",colorNameIndexKey) //color~idred123456

	if err != nil {
		return shim.Error("Fail to create Composite key")
	}

	value := []byte{0x00}
	err = stub.PutState(colorNameIndexKey, value)  // 将索引信息保保存在Key中
	if err != nil {
		return shim.Error("PutState failed")
	}
	return shim.Success(nil)
}


func (t *AgriChaincode) richQueryColor(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	color := args[1]
	indexName := args[2]

	colorIdResultsIterator, err := stub.GetStateByPartialCompositeKey (indexName, []string{color}) //返回包含给出颜色的组合键的迭代器

	if err != nil {
		return shim.Error(err.Error())
	}
	defer colorIdResultsIterator.Close()

	var dataArr []string = []string{}

	logger.Debugf("colorIdResultsIterator.HasNext() result : %v",colorIdResultsIterator.HasNext())

	if !(colorIdResultsIterator.HasNext()) {
		return shim.Success([]byte("没有查到数据"))
	}

	for colorIdResultsIterator.HasNext() {
		colorIdKey, err := colorIdResultsIterator.Next()

		if err != nil {
			return shim.Error(err.Error())
		}
		objectType, compisiteKeys, err := stub.SplitCompositeKey(string(colorIdKey.Key)) //通过SplitCompositeKey 解析出Car的主键 ID

		returnColor := compisiteKeys[0]
		returnId := compisiteKeys[1]

		logger.Debugf("indexName:%s",objectType)
		logger.Debugf("returnColor:%s",returnColor)
		logger.Debugf("returnId:%s",returnId)

		carBytes, err := stub.GetState(returnId)  // 根据解析出的ID获取数据
		dataArr = append(dataArr,string(carBytes))
	}

	joinStr :=strings.Join(dataArr,"_")


	logger.Debugf("搜索结果的数组dataArr的长度：%v",len(dataArr))

	return shim.Success([]byte(joinStr))
}


func (t *AgriChaincode) getQueryResultForQueryStrings(stub shim.ChaincodeStubInterface, args []string)pb.Response {

	// 统一的方法用于couchDB的查询，返回查询到的数据，queryString为查询语句
	queryString := args[1]
	logger.Debugf("- getQueryResultForQueryStrings queryString:\n%s\n", queryString)
	resultsIterator, err := stub.GetQueryResult(queryString)
	logger.Debugf("------执行了查询语句了")
	defer resultsIterator.Close()
	if err != nil {
		logger.Debugf("------执行查询语句出错：",err.Error())
		shim.Error(err.Error())
	}
	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString(`{"result":[`)
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")
		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString(`]}`)
	fmt.Printf("- getQueryResultForQueryStrings queryResult:\n%s\n", buffer.String())
	return shim.Success(buffer.Bytes())
}


func main() {
	fmt.Println("prj2_main start")
	logger.Info("prj2_main start")
	err := shim.Start(new(AgriChaincode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}
