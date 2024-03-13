package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	mower "github.com/nalle631/fabric-network/chaincode/c2b/mower/chaincode"
)

func main() {
	mowerChaincode, err := contractapi.NewChaincode(&mower.SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-private-data chaincode: %v", err)
	}

	if err := mowerChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-private-data chaincode: %v", err)
	}
}
