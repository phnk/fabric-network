package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	battery "github.com/nalle631/fabric-network/chaincode/b2b/battery-contract/chaincode"
)

func main() {
	razorChaincode, err := contractapi.NewChaincode(&battery.SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-private-data chaincode: %v", err)
	}

	if err := razorChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-private-data chaincode: %v", err)
	}
}
