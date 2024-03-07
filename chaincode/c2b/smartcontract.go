package chaincode

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
	AppraisedValue int      `json:"AppraisedValue"`
	GrassLength    float32  `json:"GrassLength"`
	ID             string   `json:"ID"`
	Partisipants   []string `json:"Partisipants"`
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateSLA(ctx contractapi.TransactionContextInterface, id string, grasslength float32, partisipants []string, appraisedValue int) error {
	exists, err := s.SLAExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := SLA{
		ID:             id,
		GrassLength:    grasslength,
		Partisipants:   partisipants,
		AppraisedValue: appraisedValue,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadSLA(ctx contractapi.TransactionContextInterface, id string) (*SLA, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset SLA
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateGrassLenth(ctx contractapi.TransactionContextInterface, id string, grasslength float32) error {
	exists, err := s.SLAExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
	sla, readSLAerror := s.ReadSLA(ctx, id)
	if err != nil {
		return readSLAerror
	}

	var newAppraisedValue int
	if grasslength < sla.GrassLength {

		floatvalue := float32(sla.AppraisedValue)
		newAppraisedValue = int(floatvalue * 1.1)
	} else if grasslength > sla.GrassLength {

		floatvalue := float32(sla.AppraisedValue)
		newAppraisedValue = int(floatvalue * 0.9)
	} else {
		newAppraisedValue = sla.AppraisedValue
	}
	asset := SLA{
		GrassLength:    grasslength,
		AppraisedValue: newAppraisedValue,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
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
