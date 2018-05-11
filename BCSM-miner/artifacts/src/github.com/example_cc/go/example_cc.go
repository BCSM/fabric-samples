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

package main


import (
	"net"
	"fmt"
	"strings"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("example_cc0")

type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("########### BCSM-miner Init ###########")

	var counter_text string
	var counter int
	var err error

	counter = 0
	counter_text = strconv.Itoa(counter)

	err = stub.PutState("counter", []byte(counter_text))
	if err == nil {
		return shim.Success(nil)
	}else{
		return shim.Error(err.Error())
	}
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### BCSM-miner Invoke ###########")

	function, args := stub.GetFunctionAndParameters()

	if function == "query" {
		return t.query(stub, args)
	}

	if function == "upload" {
		return t.upload(stub, args)
	}

	logger.Errorf("Unknown action, check the first argument, must be one of 'query' or 'upload'. But got: %v", args[0])
	return shim.Error(fmt.Sprintf("Unknown action, check the first argument, must be one of 'query' or 'upload'. But got: %v", args[0]))
}

func (t *SimpleChaincode) upload(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var Tx string

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1, the Tx")
	}

	Tx = args[0]

	logger.Infof("start to make the network connection to the outside")

	// temporary fix with fixed IP
	conn, _ := net.Dial("tcp", "172.17.0.1:6602")
	conn.Write([]byte(Tx))

	logger.Infof("waiting for response")
	buff := make([]byte, 1024)
	n, _ := conn.Read(buff)

	response_str := string(buff[:n])

	if strings.Compare(response_str, "1") == 0 {
		var counter int
		var counter_text string

		counterbytes, err := stub.GetState("counter")
		if err != nil {
			return shim.Error("Failed to get counter state")
		}

		if counterbytes == nil {
			return shim.Error("Counter is not valid")
		}

		counter_text = string(counterbytes)
		counter, _ = strconv.Atoi(counter_text)

		counter = counter + 1
		counter_text = strconv.Itoa(counter)

		err = stub.PutState("counter", []byte(counter_text))
		if err != nil {
			return shim.Error(err.Error())
		}

		err = stub.PutState(counter_text, []byte(Tx))
		if err != nil {
			return shim.Error(err.Error())
		}

		return shim.Success(nil)
	}else{
		return shim.Error("fail the Rx verification")
	}
}

func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var A string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

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
	logger.Infof("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}
