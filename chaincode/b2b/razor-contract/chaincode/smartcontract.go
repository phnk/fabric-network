package razor

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Job struct {
	Type     string    `json:"Type"`
	Status   string    `json:"Status"`
	Pay      int       `json:"Pay"`
	Deadline time.Time `json:"Deadline,omitempty"`
	ID       string    `json:"ID"`
	Mower    string    `json:"Mower"`
	Area     string    `json:"Area"`
	Location string    `json:"Location"`
}

func (s *SmartContract) Create(ctx contractapi.TransactionContextInterface, technichianID string, jobID string, mower string, area string, location string) (*Job, error) {
	jobExistsOnLedger, err := s.JobExistsOnLedger(ctx, jobID)
	fmt.Println("Mower: ", mower)

	if err != nil {
		fmt.Println("Error when checking if job exists on ledger: ", err)
		return nil, err
	}

	if jobExistsOnLedger {
		fmt.Println("Job already exists on ledger")
		return nil, fmt.Errorf("Job %s already exists on ledger", jobID)
	}

	existsOffLedger, err := s.JobExistsOffLedger(jobID, technichianID)
	if err != nil {
		fmt.Println("Error checking if job exists off ledger, ", err)
		return nil, err
	}

	if !existsOffLedger {
		fmt.Println("Job does not exist")
		return nil, fmt.Errorf("Job %s does not exist off ledger", jobID)
	}
	job := Job{
		Type:     "razor",
		Status:   "Ongoing",
		Pay:      150,
		ID:       jobID,
		Mower:    mower,
		Area:     area,
		Location: location,
	}
	jobJSON, err := json.Marshal(job)
	if err != nil {
		fmt.Println("Error marshalling job: ", err)
		return nil, err
	}
	err = ctx.GetStub().PutState(jobID, jobJSON)
	if err != nil {
		fmt.Println("Error putting job to world state: ", err)
		return nil, fmt.Errorf("failed to put to world state. %v", err)
	}

	return &job, nil
}

func (s *SmartContract) JobExistsOnLedger(ctx contractapi.TransactionContextInterface, jobID string) (bool, error) {
	jobJSON, err := ctx.GetStub().GetState(jobID)
	fmt.Println("jobJSON: ", string(jobJSON))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	if jobJSON == nil {
		return false, nil
	}

	var job Job
	err = json.Unmarshal(jobJSON, &job)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *SmartContract) JobExistsOffLedger(jobID string, technicianID string) (bool, error) {
	// TODO: check jespers system if the job exists or not and what type of job it is
	return true, nil
}
