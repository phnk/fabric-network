package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	bumpy "github.com/nalle631/fabric-network/chaincode/b2b/bumpy-contract/chaincode"
)

func main() {
	bumpyChaincode, err := contractapi.NewChaincode(&bumpy.SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-private-data chaincode: %v", err)
	}

	if err := bumpyChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-private-data chaincode: %v", err)
	}
}
