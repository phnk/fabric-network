package customer

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Customer struct {
	ID   string   `json:"ID"`
	SLAs []string `json:"SLAs"`
}

type SLA struct {
	AppraisedValue    int     `json:"AppraisedValue,omitempty"`
	TargetGrassLength float32 `json:"TargetGrassLength"`
	MaxGrassLength    float32 `json:"MaxGrassLength"`
	MinGrassLength    float32 `json:"MinGrassLength"`
	ID                string  `json:"ID"`
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateCustomer(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.CustomerExist(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the customer %s already exists", id)
	}

	newCustomer := Customer{
		ID:   id,
		SLAs: []string{},
	}

	customerJSON, err := json.Marshal(newCustomer)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, customerJSON)
}

func (s *SmartContract) CreateMower(ctx contractapi.TransactionContextInterface, customerID, mowerID string, targetgrasslength float32, maxgrasslength float32, mingrasslength float32) error {
	exists, err := s.CustomerExist(ctx, customerID)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("the customer %s does not exist", customerID)
	}

	customer, readSLAerror := s.ReadCustomer(ctx, customerID)
	if err != nil {
		return readSLAerror
	}

	targetgrasslength_string := fmt.Sprintf("%f", targetgrasslength)
	maxgrasslength_string := fmt.Sprintf("%f", maxgrasslength)
	mingrasslength_string := fmt.Sprintf("%f", mingrasslength)

	invokeArgs := [][]byte{[]byte("CreateSLA"), []byte(mowerID), []byte(targetgrasslength_string), []byte(maxgrasslength_string), []byte(mingrasslength_string)}
	response := ctx.GetStub().InvokeChaincode("mower", invokeArgs, ctx.GetStub().GetChannelID())
	fmt.Println("response status: ", response.Status)
	if response.Status != shim.OK {
		fmt.Println("failed to invoke chaincode. Got error: %s", response.Payload)
		return fmt.Errorf("Failed to invoke chaincode. Got error: %s", response.Payload)
	}
	customer.SLAs = append(customer.SLAs, mowerID)
	customerJSON, err := json.Marshal(customer)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(customerID, customerJSON)

}

func (s *SmartContract) ReadCustomer(ctx contractapi.TransactionContextInterface, id string) (*Customer, error) {
	customerJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if customerJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var customer Customer
	err = json.Unmarshal(customerJSON, &customer)
	if err != nil {
		return nil, err
	}

	return &customer, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateTargetGrassLength(ctx contractapi.TransactionContextInterface, customerID string, mowerID string, targetgrasslength float32) error {
	exists, err := s.CustomerExist(ctx, customerID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the customer %s does not exist", customerID)
	}

	// overwriting original asset with new asset
	customer, readSLAerror := s.ReadCustomer(ctx, customerID)
	if err != nil {
		return readSLAerror
	}

	targetgrasslength_string := fmt.Sprintf("%f", targetgrasslength)

	invokeArgs := [][]byte{[]byte("UpdateTargetGrassLength"), []byte(mowerID), []byte(targetgrasslength_string)}
	for _, id := range customer.SLAs {
		if id == mowerID {
			response := ctx.GetStub().InvokeChaincode("mower", invokeArgs, ctx.GetStub().GetChannelID())
			fmt.Println("response status: ", response.Status)
			if response.Status != shim.OK {
				fmt.Println("failed to invoke chaincode. Got error: %s", response.Payload)
				return fmt.Errorf("Failed to invoke chaincode. Got error: %s", response.Payload)
			}
			return nil
		}
	}
	return fmt.Errorf("could not update target grasslength")
}

func (s *SmartContract) UpdateGrassLengthInterval(ctx contractapi.TransactionContextInterface, customerID string, mowerID string, maxgrasslength float32, mingrasslength float32) error {
	exists, err := s.CustomerExist(ctx, customerID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the customer %s does not exist", customerID)
	}

	customer, readSLAerror := s.ReadCustomer(ctx, customerID)
	if err != nil {
		return readSLAerror
	}

	maxgrasslength_string := fmt.Sprintf("%f", maxgrasslength)
	mingrasslength_string := fmt.Sprintf("%f", mingrasslength)

	invokeArgs := [][]byte{[]byte("UpdateGrassLengthInterval"), []byte(mowerID), []byte(maxgrasslength_string), []byte(mingrasslength_string)}
	for _, id := range customer.SLAs {
		if id == mowerID {
			response := ctx.GetStub().InvokeChaincode("mower", invokeArgs, ctx.GetStub().GetChannelID())
			fmt.Println("response status: ", response.Status)
			if response.Status != shim.OK {
				fmt.Println("failed to invoke chaincode. Got error: %s", response.Payload)
				return fmt.Errorf("Failed to invoke chaincode. Got error: %s", response.Payload)
			}
			return nil
		}
	}
	return fmt.Errorf("could not update grasslength interval")
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) RemoveMowerSLA(ctx contractapi.TransactionContextInterface, customerID string, mowerID string) error {
	exists, err := s.CustomerExist(ctx, customerID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", customerID)
	}

	customer, readSLAerror := s.ReadCustomer(ctx, customerID)
	if err != nil {
		return readSLAerror
	}

	invokeArgs := [][]byte{[]byte("DeleteSLA"), []byte(mowerID)}
	for i, id := range customer.SLAs {
		if id == mowerID {
			response := ctx.GetStub().InvokeChaincode("mower", invokeArgs, ctx.GetStub().GetChannelID())
			fmt.Println("response status: ", response.Status)
			if response.Status != shim.OK {
				fmt.Println("failed to invoke chaincode. Got error: %s", response.Payload)
				return fmt.Errorf("Failed to invoke chaincode. Got error: %s", response.Payload)
			}
			fmt.Println("CustomerSLAs before remove: ", customer.SLAs)
			newSLAs := remove(customer.SLAs, i)
			customer.SLAs = newSLAs
			customerJSON, err := json.Marshal(customer)
			if err != nil {
				return err
			}
			fmt.Println("CustomerSLAs after remove: ", customer.SLAs)
			return ctx.GetStub().PutState(customerID, customerJSON)
		}
	}
	return fmt.Errorf("could not find mower with ID %s", mowerID)
}

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) CustomerExist(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	customerJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return customerJSON != nil, nil
}

func (s *SmartContract) ReadMowerSLA(ctx contractapi.TransactionContextInterface, customerID string, mowerID string) (*SLA, error) {
	exists, err := s.CustomerExist(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("the asset %s does not exist", customerID)
	}

	customer, readSLAerror := s.ReadCustomer(ctx, customerID)
	if err != nil {
		return nil, readSLAerror
	}

	invokeArgs := [][]byte{[]byte("ReadSLA"), []byte(mowerID)}
	for _, id := range customer.SLAs {
		if id == mowerID {
			response := ctx.GetStub().InvokeChaincode("mower", invokeArgs, ctx.GetStub().GetChannelID())

			fmt.Println("response status: ", response.Status)
			fmt.Println("response.payload: ", response.Payload)
			if response.Status != shim.OK {
				fmt.Println("failed to invoke chaincode. Got error: %s", response.Payload)
				return nil, fmt.Errorf("Failed to invoke chaincode. Got error: %s", response.Payload)
			}

			var slaResponse SLA
			err = json.Unmarshal(response.Payload, &slaResponse)
			if err != nil {
				return nil, err
			}

			return &slaResponse, nil
		}
	}
	return nil, fmt.Errorf("could not find mower with ID %s", mowerID)
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllSLA(ctx contractapi.TransactionContextInterface, customerID string) ([]*SLA, error) {
	exists, err := s.CustomerExist(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("the asset %s does not exist", customerID)
	}

	customer, readSLAerror := s.ReadCustomer(ctx, customerID)
	if err != nil {
		return nil, readSLAerror
	}
	var allSLA []*SLA
	for _, slaID := range customer.SLAs {
		sla, err := s.ReadMowerSLA(ctx, customerID, slaID)
		if err != nil {
			return nil, err
		}
		allSLA = append(allSLA, sla)
	}

	fmt.Println("allSLAs: ", allSLA)
	return allSLA, nil
}
