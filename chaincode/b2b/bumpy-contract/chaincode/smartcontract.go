package bumpy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/nalle631/arrowheadfunctions"
)

const (
	arrowheadcertsPath  = "../certs"
	arrowheadKey        = arrowheadcertsPath + "/technician-key.pem"
	arrowheadCert       = arrowheadcertsPath + "technician-cert.pem"
	arrowheadTruststore = arrowheadcertsPath + "/truststore.pem"
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
	Address       string    `json:"Address"`
}

func (s *SmartContract) Create(ctx contractapi.TransactionContextInterface, technichianID string, jobID string, mower string, address string, deadline string) (*Job, error) {
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

	timeDeadline, err := time.Parse("2006-01-02T15:04:05Z07:00", deadline)
	if err != nil {
		fmt.Println("Error parsing deadline: ", err)
		return nil, err
	}
	job := Job{
		Type:          "bumpy",
		Status:        "Ongoing",
		Deadline:      timeDeadline,
		JobPay:        50,
		InspectionPay: 50,
		ID:            jobID,
		Mower:         mower,
		Address:       address,
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
	orchIP := "35.228.60.153"
	orchPort := 8443
	var requestBody arrowheadfunctions.Orchestrate
	requestBody.OrchestrationFlags.EnableInterCloud = false
	requestBody.OrchestrationFlags.OverrideStore = false
	requestBody.RequestedService.InterfaceRequirements = []string{"HTTP-SECURE-JSON"}
	requestBody.RequestedService.ServiceDefinitionRequirement = "assign-worker"
	requestBody.RequesterSystem.SystemName = "technician"
	requestBody.RequesterSystem.AuthenticationInfo = ""
	requestBody.RequesterSystem.Port = 5000
	requestBody.RequesterSystem.Address = "35.228.161.184"

	// TODO: check jespers system if the job exists or not and what type of job it is
	var orchResponse arrowheadfunctions.OrchResponse
	jsonOrchResponse := arrowheadfunctions.Orchestration(requestBody, orchIP, orchPort, arrowheadCert, arrowheadKey, arrowheadTruststore)
	json.Unmarshal(jsonOrchResponse, &orchResponse)
	chosenResponse := orchResponse.Response[0]
	fmt.Println("response from neginfo: ", chosenResponse)

	req, err := http.NewRequest("POST", "https://"+chosenResponse.Provider.Address+":"+strconv.Itoa(chosenResponse.Provider.Port)+chosenResponse.ServiceUri, nil)
	if err != nil {
		return false, err
	}

	client := arrowheadfunctions.GetClient(arrowheadCert, arrowheadKey, arrowheadTruststore)
	serviceResp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making HTTP request using client. ", err)
		return false, err
	}
	if serviceResp.StatusCode == 404 {
		return false, nil
	}

	return true, nil

}
