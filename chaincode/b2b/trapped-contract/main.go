package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	trapped "github.com/nalle631/fabric-network/chaincode/b2b/trapped-contract/chaincode"
)

func main() {
	razorChaincode, err := contractapi.NewChaincode(&trapped.SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-private-data chaincode: %v", err)
	}

	if err := razorChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-private-data chaincode: %v", err)
	}
}
