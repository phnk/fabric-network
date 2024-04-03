package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	gc "github.com/nalle631/fabric-network/chaincode/b2b/job-contract/chaincode"
)

func main() {
	gcChaincode, err := contractapi.NewChaincode(&gc.SmartContract{})

	if err != nil {
		log.Panicf("Error creating asset-transfer-private-data chaincode: %v", err)
	}

	if err := gcChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-private-data chaincode: %v", err)
	}
}
