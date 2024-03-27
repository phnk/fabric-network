package gc

import (
	"encoding/json"
	"fmt"
	"time"

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
type Job struct {
	Type          string    `json:"Type"`
	Status        string    `json:"Status"`
	JobPay        int       `json:"JobPay"`
	InspectionPay int       `json:"InspectionPay"`
	Deadline      time.Time `json:"Deadline,omitempty"`
	ID            string    `json:"ID"`
	Mower         string    `json:"Mower"`
	Area          string    `json:"Area"`
	Location      string    `json:"Location"`
}

type GeneralContract struct {
	TechnicianID   string   `json:"TechnicianID"`
	MonthlyBalance int      `json:"MonthlyBalance"`
	Jobs           []Job    `json:"Jobs"`
	JobAuthority   []string `json:"JobAuthority"`
}

// InitLedger adds a base set of assets to the ledger
// func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
// 	assets := []Asset{
// 		{ID: "asset1", Color: "blue", Size: 5, Owner: "Tomoko", AppraisedValue: 300},
// 		{ID: "asset2", Color: "red", Size: 5, Owner: "Brad", AppraisedValue: 400},
// 		{ID: "asset3", Color: "green", Size: 10, Owner: "Jin Soo", AppraisedValue: 500},
// 		{ID: "asset4", Color: "yellow", Size: 10, Owner: "Max", AppraisedValue: 600},
// 		{ID: "asset5", Color: "black", Size: 15, Owner: "Adriana", AppraisedValue: 700},
// 		{ID: "asset6", Color: "white", Size: 15, Owner: "Michel", AppraisedValue: 800},
// 	}

// 	atm := time.Now()
// 	for _, asset := range assets {
// 		assetJSON, err := json.Marshal(asset)
// 		if err != nil {
// 			return err
// 		}

// 		err = ctx.GetStub().PutState(asset.ID, assetJSON)
// 		if err != nil {
// 			return fmt.Errorf("failed to put to world state. %v", err)
// 		}
// 	}

// 	return nil
// }

func (s *SmartContract) CreateGeneralContract(ctx contractapi.TransactionContextInterface) error {
	gcID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return err
	}

	fmt.Println("gcID: ", gcID)
	gcExists, err := s.GeneralContractExists(ctx, gcID)

	if err != nil {
		return fmt.Errorf("failed to check if general contract already exists: %v", err)
	}
	if gcExists {
		return fmt.Errorf("general contract already exists")
	}

	gc := GeneralContract{
		TechnicianID:   gcID,
		MonthlyBalance: 0,
		Jobs:           []Job{},
		JobAuthority:   []string{},
	}
	gcJSON, err := json.Marshal(gc)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(gcID, gcJSON)

}

// Remember to remove jobtype when integrated with jespers system
func (s *SmartContract) TakeJob(ctx contractapi.TransactionContextInterface, jobID string, technichianID string, jobType string) error {
	exists, err := s.GeneralContractExists(ctx, technichianID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("General contract for %s does not exist", technichianID)
	}

	jobExistsOnLedger, err := s.JobExistsOnLedger(ctx, jobID, technichianID)

	if err != nil {
		return err
	}

	if jobExistsOnLedger {
		return fmt.Errorf("Job %s already exists on ledger", jobID)
	}

	// Remember to remove jobtype when integrated with jespers system
	_, err = s.JobExistsOffLedger(jobID, technichianID)
	if err != nil {
		return err
	}
	mower, err := getMower(jobID)
	if err != nil {
		return err
	}
	area, err := getArea(jobID)
	if err != nil {
		return err
	}
	location, err := getLocation(jobID)
	if err != nil {
		return err
	}

	if jobType == "" {
		return fmt.Errorf("Job %s does not exist in external system", jobID)
	}

	invokeArgs := [][]byte{[]byte("Create"), []byte(technichianID), []byte(jobID), []byte(mower), []byte(area), []byte(location)}
	response := ctx.GetStub().InvokeChaincode(jobType, invokeArgs, ctx.GetStub().GetChannelID())
	fmt.Println("response status: ", response.Status)
	if response.Status != shim.OK {
		fmt.Println("failed to invoke chaincode. Got error: %s", response.Payload)
		return fmt.Errorf("Failed to invoke chaincode. Got error: %s", response.Payload)
	}
	var createdJob Job
	err = json.Unmarshal(response.Payload, &createdJob)
	if err != nil {
		fmt.Println("Failed to unmarshal, ", err)
		return err
	}
	addJob(ctx, &createdJob)

	return nil
}

func (s *SmartContract) JobDoneCorrectError(ctx contractapi.TransactionContextInterface, jobID string) error {
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	exists, err := s.GeneralContractExists(ctx, mspID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("General contract for %s does not exist", mspID)
	}

	isJobDone, err := checkIfDone(jobID)
	if err != nil {
		fmt.Println("Error checking if job is done")
		return err
	}
	if !isJobDone {
		return fmt.Errorf("Job %s is not done", jobID)
	}

	job, err := s.ReadJob(ctx, jobID, mspID)
	if err != nil {
		fmt.Println("Error reading job, ", err)
		return err
	}

	gc, err := s.ReadGeneralContract(ctx, mspID)

	if err != nil {
		fmt.Println("Error reading general contract, ", err)
		return err
	}

	gc.MonthlyBalance = gc.MonthlyBalance + job.JobPay + job.InspectionPay
	job.Status = "Done"
	err = updateJobStatus(job, gc, "Done")

	if err != nil {
		fmt.Println("Error updating job status, ", err)
		return err
	}

	jobJSON, err := json.Marshal(job)

	if err != nil {
		fmt.Println("Error marshalling, ", err)
		return err
	}

	err = ctx.GetStub().PutState(jobID, jobJSON)

	if err != nil {
		fmt.Println("Error putting job to world state: ", err)
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	gcJSON, err := json.Marshal(gc)
	if err != nil {
		fmt.Println("Error marshalling, ", err)
		return err
	}

	err = ctx.GetStub().PutState(mspID, gcJSON)
	if err != nil {
		fmt.Println("Error putting job to world state: ", err)
		return fmt.Errorf("failed to put to world state. %v", err)
	}
	return nil
}

func (s *SmartContract) JobDoneWrongError(ctx contractapi.TransactionContextInterface, jobID string) error {
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	exists, err := s.GeneralContractExists(ctx, mspID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("General contract for %s does not exist", mspID)
	}

	isJobDone, err := checkIfDone(jobID)
	if err != nil {
		fmt.Println("Error checking if job is done")
		return err
	}
	if !isJobDone {
		return fmt.Errorf("Job %s is not done", jobID)
	}

	job, err := s.ReadJob(ctx, jobID, mspID)
	if err != nil {
		fmt.Println("Error reading job, ", err)
		return err
	}

	gc, err := s.ReadGeneralContract(ctx, mspID)

	if err != nil {
		fmt.Println("Error reading general contract, ", err)
		return err
	}

	gc.MonthlyBalance = gc.MonthlyBalance + job.InspectionPay
	job.Status = "Done"
	err = updateJobStatus(job, gc, "Done")

	if err != nil {
		fmt.Println("Error updating job status, ", err)
		return err
	}

	jobJSON, err := json.Marshal(job)

	if err != nil {
		fmt.Println("Error marshalling, ", err)
		return err
	}

	err = ctx.GetStub().PutState(jobID, jobJSON)

	if err != nil {
		fmt.Println("Error putting job to world state: ", err)
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	gcJSON, err := json.Marshal(gc)
	if err != nil {
		fmt.Println("Error marshalling, ", err)
		return err
	}

	err = ctx.GetStub().PutState(mspID, gcJSON)
	if err != nil {
		fmt.Println("Error putting job to world state: ", err)
		return fmt.Errorf("failed to put to world state. %v", err)
	}
	return nil
}

func updateJobStatus(job *Job, gc *GeneralContract, status string) error {
	for i, jobInGc := range gc.Jobs {
		if job.ID == jobInGc.ID {
			gc.Jobs[i] = *job
			break
		}
	}
	return nil
}

func checkIfDone(jobID string) (bool, error) {
	//Check jespers system if the job is done
	return true, nil
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadJob(ctx contractapi.TransactionContextInterface, jobID string, technicianID string) (*Job, error) {
	generalContractJSON, err := ctx.GetStub().GetState(technicianID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if generalContractJSON == nil {
		return nil, fmt.Errorf("the is no general contract for %s", technicianID)
	}

	var gc GeneralContract
	err = json.Unmarshal(generalContractJSON, &gc)
	if err != nil {
		return nil, err
	}

	for _, job := range gc.Jobs {
		if job.ID == jobID {
			return &job, nil
		}
	}

	return nil, fmt.Errorf("the job %s does not exist", jobID)
}

func (s *SmartContract) ReadGeneralContract(ctx contractapi.TransactionContextInterface, technicianID string) (*GeneralContract, error) {
	generalContractJSON, err := ctx.GetStub().GetState(technicianID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if generalContractJSON == nil {
		return nil, fmt.Errorf("the is no general contract for %s", technicianID)
	}

	var gc GeneralContract
	err = json.Unmarshal(generalContractJSON, &gc)
	if err != nil {
		return nil, err
	}

	return &gc, nil
}

// DeleteAsset deletes an given asset from the world state.
// func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
// 	exists, err := s.AssetExists(ctx, id)
// 	if err != nil {
// 		return err
// 	}
// 	if !exists {
// 		return fmt.Errorf("the asset %s does not exist", id)
// 	}

// 	return ctx.GetStub().DelState(id)
// }

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) GeneralContractExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	generalContractJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return generalContractJSON != nil, nil
}

func (s *SmartContract) JobExistsOnLedger(ctx contractapi.TransactionContextInterface, jobID string, gcID string) (bool, error) {
	generalContractJSON, err := ctx.GetStub().GetState(gcID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	var gc GeneralContract
	err = json.Unmarshal(generalContractJSON, &gc)
	if err != nil {
		return false, err
	}

	for _, job := range gc.Jobs {
		if job.ID == jobID {
			return true, nil
		}
	}

	return false, nil
}

func (s *SmartContract) JobExistsOffLedger(jobID string, technicianID string) (string, error) {
	// TODO: check jespers system if the job exists or not and what type of job it is
	// result := rand.Intn(2)
	// if result == 0 {
	// 	return "razor", nil
	// }
	return "bumpy", nil
}

func getMower(jobID string) (string, error) {
	return "1", nil
}

func getArea(jobID string) (string, error) {
	return "100", nil
}

func getLocation(jobID string) (string, error) {
	return "Porsön", nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllJobs(ctx contractapi.TransactionContextInterface) ([]*Job, error) {

	var gc GeneralContract
	gcID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return nil, err
	}
	fmt.Println("GCID: ", gcID)
	gcJSON, err := ctx.GetStub().GetState(gcID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	json.Unmarshal(gcJSON, &gc)
	var gcJobs []*Job
	for _, job := range gc.Jobs {
		gcJobs = append(gcJobs, &job)
	}
	return gcJobs, nil
}

func addJob(ctx contractapi.TransactionContextInterface, job *Job) error {
	var gc GeneralContract
	gcID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return err
	}
	fmt.Println("GCID: ", gcID)
	gcJSON, err := ctx.GetStub().GetState(gcID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	json.Unmarshal(gcJSON, &gc)

	gc.Jobs = append(gc.Jobs, *job)

	gcBytes, err := json.Marshal(gc)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(gcID, gcBytes)
}
