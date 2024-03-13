package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	customer "github.com/nalle631/fabric-network/chaincode/c2b/customer/chaincode"
)

func main() {
	customerChaincode, err := contractapi.NewChaincode(&customer.SmartContract{})
	if err != nil {
		log.Panicf("Error creating customer chaincode: %v", err)
	}

	if err := customerChaincode.Start(); err != nil {
		log.Panicf("Error starting customer chaincode: %v", err)
	}
}
