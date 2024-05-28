package mower

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type SLA struct {
	AppraisedValue    int     `json:"AppraisedValue,omitempty"`
	ServiceLevel      string  `json:"ServiceLevel"`
	TargetGrassLength float32 `json:"TargetGrassLength"`
	MaxGrassLength    float32 `json:"MaxGrassLength"`
	MinGrassLength    float32 `json:"MinGrassLength"`
	ID                string  `json:"ID"`
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateSLA(ctx contractapi.TransactionContextInterface, id string, serviceLevel string, targetgrasslength float32, maxgrasslength float32, mingrasslength float32) (*SLA, error) {
	fmt.Println("In CreateSLA in mower contract")

	exists, err := s.SLAExists(ctx, id)
	if err != nil {
		fmt.Println("sla aready exists")
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("the SLA %s already exists", id)
	}

	newSLA := SLA{
		AppraisedValue:    0,
		ID:                id,
		ServiceLevel:      serviceLevel,
		TargetGrassLength: targetgrasslength,
		MaxGrassLength:    maxgrasslength,
		MinGrassLength:    mingrasslength,
	}

	fmt.Println("SLA before evaluation: ", newSLA)

	slaValue, err := s.EvaluateSLA(ctx, newSLA.ServiceLevel, newSLA.TargetGrassLength, newSLA.MaxGrassLength, newSLA.MinGrassLength)
	fmt.Println("slaValue: ", slaValue)
	if err != nil {
		fmt.Println("error evaluating SLA: ", err)
		return nil, err
	}

	newSLA.AppraisedValue = slaValue
	fmt.Println("SLA after evaluation: ", newSLA)
	slaJSON, err := json.Marshal(newSLA)
	if err != nil {
		fmt.Println("Error marshalling SLA: ")
		return nil, err
	}
	ctx.GetStub().PutState(id, slaJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to put to world state. %v", err)
	}

	return &newSLA, nil
}

func (s *SmartContract) ChangeServiceLevel(ctx contractapi.TransactionContextInterface, slaID string, newServiceLevel string) (*SLA, error) {
	fmt.Println("In ChangeServiceLevel")
	exists, err := s.SLAExists(ctx, slaID)
	if err != nil {
		fmt.Println("error checking if sla exists")
		return nil, err
	}
	if !exists {
		fmt.Println("SLA not found")
		return nil, err
	}

	sla, err := s.ReadSLA(ctx, slaID)
	if err != nil {
		fmt.Println("Error reading sla")
		return nil, err
	}

	switch newServiceLevel {
	case "standard":
		sla.ServiceLevel = "standard"
	case "gold":
		fmt.Println("In gold mode")
		sla.ServiceLevel = "gold"
	case "platinum":
		sla.ServiceLevel = "platinum"

	default:
		fmt.Println("In default mode")
		return nil, fmt.Errorf("invalid service level")
	}

	newAppraisedValue, err := s.EvaluateSLA(ctx, sla.ServiceLevel, sla.TargetGrassLength, sla.MaxGrassLength, sla.MinGrassLength)
	if err != nil {
		fmt.Println("error evaluating SLA")
		return nil, err
	}

	sla.AppraisedValue = newAppraisedValue

	slaJSON, err := json.Marshal(sla)
	if err != nil {
		return nil, err
	}
	err = ctx.GetStub().PutState(slaID, slaJSON)
	if err != nil {
		return nil, err
	}
	return sla, nil
}

func (s *SmartContract) EvaluateSLA(ctx contractapi.TransactionContextInterface, serviceLevel string, targetGrassLength float32, maxGrassLength float32, minGrassLength float32) (int, error) {

	spread := maxGrassLength - minGrassLength

	fmt.Println("spread: ", spread)

	// Invert the spread for cost calculation (larger spread, lower cost)
	inverseSpread := 1.0 / spread

	// Cost factor based on target length (shorter target, higher cost)
	targetFactor := 1.0 / targetGrassLength

	fmt.Println("Target factor: ", targetFactor)

	// Combine factors with a weighting factor (adjust weight as needed)
	costFactor := (inverseSpread * 0.7) + (targetFactor * 0.3)

	fmt.Println("Cost factor: ", costFactor)

	// Base cost (adjust as needed)
	var baseCost float32
	switch serviceLevel {
	case "standard":
		baseCost = 50
	case "gold":
		baseCost = 100
	case "platinum":
		baseCost = 200

	default:
		return 0, fmt.Errorf("invalid service level: %s", serviceLevel)
	}

	// Calculate final cost
	monthlyCost := baseCost * (costFactor + 1)

	fmt.Println("Monthly cost: ", monthlyCost)

	fmt.Println("Monthly cost integer: ", int(monthlyCost))

	return int(monthlyCost), nil

}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadSLA(ctx contractapi.TransactionContextInterface, id string) (*SLA, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		fmt.Println("Error getting world state")
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		println("assetJSON is nil")
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset SLA
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		fmt.Println("error unmarshalling SLA")
		return nil, err
	}

	fmt.Println("SLA: ", asset)
	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateTargetGrassLength(ctx contractapi.TransactionContextInterface, id string, targetgrasslength float32) (*SLA, error) {
	exists, err := s.SLAExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
	sla, readSLAerror := s.ReadSLA(ctx, id)
	if err != nil {
		return nil, readSLAerror
	}

	sla.TargetGrassLength = targetgrasslength

	newValue, err := s.EvaluateSLA(ctx, sla.ServiceLevel, sla.TargetGrassLength, sla.MaxGrassLength, sla.MinGrassLength)
	if err != nil {
		return nil, err
	}

	sla.AppraisedValue = newValue
	assetJSON, err := json.Marshal(sla)
	if err != nil {
		return nil, err
	}
	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to put to world state. %v", err)
	}
	return sla, nil
}

func (s *SmartContract) UpdateGrassLengthInterval(ctx contractapi.TransactionContextInterface, id string, maxgrasslength float32, mingrasslength float32) (*SLA, error) {
	exists, err := s.SLAExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
	sla, readSLAerror := s.ReadSLA(ctx, id)
	if err != nil {
		return nil, readSLAerror
	}

	sla.MaxGrassLength = maxgrasslength
	sla.MinGrassLength = mingrasslength

	newValue, err := s.EvaluateSLA(ctx, sla.ServiceLevel, sla.TargetGrassLength, sla.MaxGrassLength, sla.MinGrassLength)
	if err != nil {
		return nil, err
	}

	sla.AppraisedValue = newValue
	assetJSON, err := json.Marshal(sla)
	if err != nil {
		return nil, err
	}
	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to put to world state. %v", err)
	}

	return sla, nil
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteSLA(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.SLAExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) SLAExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	slaJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return slaJSON != nil, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllSLA(ctx contractapi.TransactionContextInterface) ([]*SLA, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var slas []*SLA
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset SLA
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		slas = append(slas, &asset)
	}

	return slas, nil
}
