package gc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/nalle631/arrowheadfunctions"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	arrowheadTruststoreString = "-----BEGIN CERTIFICATE-----MIIDMDCCAhigAwIBAgIEZV4AEDANBgkqhkiG9w0BAQsFADAXMRUwEwYDVQQDDAxhcnJvd2hlYWQuZXUwHhcNMjMxMTIyMTMyMDE2WhcNMzMxMTIyMTMyMDE2WjAlMSMwIQYDVQQDDBptYWluY2xvdWQubHR1LmFycm93aGVhZC5ldTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAJ+Ng4zU1o1Mtw28v3GeQEm0QSs8QvBu3GUTIB+cYc9rCgTj0UCCIyati7pM6cCVvBD7qGqGeppi7VNye/HXDhDUff6jFVOlTvkGGtF/lBgU3nOsiL2CU6F6vTeUP/kNPC1NegVGqqOezHaaWpyKlsPyN4iye/G5cF2Cewi8WYFAyhfOkNzdScqu4JhJvfp7EKlxqT/oh3dq6MpkazPjiM63lMf68BbyL/PIFlIvRTxLi7FADYIX25G5dQj13s+gtpCyjQKj2MML/kA2MiSRQzxz7a+/NdiZrRcLfvmY0Em+Ok81cnyXSntRA7sZEixmOAhceq3tp/WRUCnvilcmW20CAwEAAaN2MHQwDwYDVR0TBAgwBgEB/wIBAjBCBgNVHSMEOzA5gBSaipkeMUwesuYttprtxx4xyt07/qEbpBkwFzEVMBMGA1UEAwwMYXJyb3doZWFkLmV1ggRc1T6tMB0GA1UdDgQWBBRcoKfiP4nJcWlBQ5Dm9oBGRTqy3zANBgkqhkiG9w0BAQsFAAOCAQEAjXIlg7WaK3/+6l6J6VRIlFHCApNs55DndBvFcI1EqLrrFYSFJ2a950YKFOWafAd4t12M5e+9G/4NrdjEB4DCZhI4OGvwW4znQE7Dt3g6TPNOrHIFm5snepBZXr4B6aw+/gQTUJfR1HrxX5RLySmDJdKLFHDh3FipKMUQhVEfuLu/e0NSV1rWW3t1+ksI/JOkuD8nQKaNLDBB16p3o3MhT3J0U2jMPHlzeaIJ1tfxraexrP6qkFiX0oFebsLEL7XL7ZAGXHW+0R2W2F55bAwjIbRatvbTOEQ1me8Brt1aER4AbV3RkF+ae7jwi0UN0v7ReOxxmcJKI2OehlohvF0DQA==-----END CERTIFICATE-----"
	arrowheadCertString       = "-----BEGIN CERTIFICATE-----MIIDeDCCAmCgAwIBAgIUZWAD8bKll1BC6MAYDcgBjpkA6aMwDQYJKoZIhvcNAQELBQAwJTEjMCEGA1UEAwwabWFpbmNsb3VkLmx0dS5hcnJvd2hlYWQuZXUwHhcNMjQwNDAxMDkzMjM3WhcNMzQwNDAxMDkzMjM3WjAwMS4wLAYDVQQDDCV0ZWNobmljaWFuLm1haW5jbG91ZC5sdHUuYXJyb3doZWFkLmV1MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAp8sdeceqGkSFPN/kN0dcQ8Xwf+OrDUO1cIkndhWS+wW02YhC1/s4OHQaj9jGvgy869t2DzGvCbF0eFfJfiEyFw+rks6MAFdSN9WW8imHI2461PtX/1/U1JgVWgbKEEZPxG+KrXKjif0tYcFt453HL8oPwyl4KCsrJdttM2QjJX9Tamxh6xFsiJhJpImGH2R7fH2vqyRriULwX1Zd3UxwPfjZfIR2PhcgOIgyzQ5cqu3qNpNlDuEDWChG/fJxvY/CXFJsIgv1NXacDHr+SNzKjRTRd2B6ErshWi4Q5FSBYP3glMwTMX+rERnCB1SHwrPXoobvT6luZ/sVkiieZk2cTwIDAQABo4GUMIGRMCwGA1UdEQQlMCOCCWxvY2FsaG9zdIIKdGVjaG5pY2lhbocEfwAAAYcEI+ShuDBCBgNVHSMEOzA5gBRcoKfiP4nJcWlBQ5Dm9oBGRTqy36EbpBkwFzEVMBMGA1UEAwwMYXJyb3doZWFkLmV1ggRlXgAQMB0GA1UdDgQWBBQrcvp7JGRkRdRA4Tnb7oYUKeSF4TANBgkqhkiG9w0BAQsFAAOCAQEAkT64CMhB73lOfjMOQzSlNrLxASsRVPvnmLsUWi2iJskMEE5mvRYelLAWBDQa3Z+uejdNuQMRVIaJlhb8ZoY1bEYn4SHQgs0KXKv3hnSP134dFvYGKQIC0gL/UesLmlrVJ1y3kZvNIJzvGXTpqIIIBjLdl1lN1kqkKWdIChIztuqGnZfHIKmq+bq2SkzpqvTICGncOJqy4ZPQu6qvK4C+LvcnLa5CRfRDxq8uCs69BPxiIAIcyMJhD91XgkJ/Kxh6EBiCfefK2zZQN383ydF76GTE4Osr45Dg/h7oi1OfFXcwDp7E/bE9jhySXp1ZAnkfgHy0NDw8NU/WFIPFbOvBFg==-----END CERTIFICATE-----"
	arrowheadKeyString        = "-----BEGIN PRIVATE KEY-----MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCnyx15x6oaRIU83+Q3R1xDxfB/46sNQ7VwiSd2FZL7BbTZiELX+zg4dBqP2Ma+DLzr23YPMa8JsXR4V8l+ITIXD6uSzowAV1I31ZbyKYcjbjrU+1f/X9TUmBVaBsoQRk/Eb4qtcqOJ/S1hwW3jnccvyg/DKXgoKysl220zZCMlf1NqbGHrEWyImEmkiYYfZHt8fa+rJGuJQvBfVl3dTHA9+Nl8hHY+FyA4iDLNDlyq7eo2k2UO4QNYKEb98nG9j8JcUmwiC/U1dpwMev5I3MqNFNF3YHoSuyFaLhDkVIFg/eCUzBMxf6sRGcIHVIfCs9eihu9PqW5n+xWSKJ5mTZxPAgMBAAECggEALXHvu8n+IjsoswIYt3gWXyy/JIQvEdqiy6X6EBtrwZ0cDEbBg+nAolmf0BHwUgz1JhQ8d4UHWWK8ntN3+TdYb7KIz6wtcvIzjHfG+DOTLF9wg7rHbJ0x8Zp3PfjUxW+lrxhewPdpn7f4kJ9o+dsD1ceuWTdkGc0HVKuHegHHGyJew3d0JZjfz91PU+5ncVsiQRenK8E1AfiboVY5CGRKewuZpuP3NOjdHKwLPlhuI5uUijjbau8Nsre7Jns+PSib+Lmm5dzsXqXRz981ZTm5MZgktp6hM9IdWztCOAxoZdliiMWwRcEHvLRXxs7gQsFxsjfD34SxZXcSTydJG5TEYQKBgQDXxAqWvAIfH1BnhMPy0V+QN/nkgj9YuKsDvGVGWrOKDaC9d09tKek8Ct50XB1m8IjVdkndec2QP/ZNUogTGcSr/SrgimFsF6mIBLuuC8Mh6L2QhPbcEdBe86Ye7HMvVce7SIrmkPFpil3f0YVYvZ1MC4T+HBicfvuLamVjs8QhJwKBgQDHFQdP8UmqG9DLLKP63K252k8A2VWsdJywhi9zFZn7kpbdw/51jXFjcxOT45rodXWMqbXRt5jDK/Zauu2w3w7FPXgHSX+jvyD7X0cC3rP5TXgTMAj9IMz5auC1mwTShBq1YHFjYhZatj2XzVoWC5x+yZwPYjqWvyMjQv+KYkNUmQKBgQCreLX9ks6Q0z7/9cgPygsLPyEOU+Bp7uCh6HAH2H9EoI305MOvepZLzEt759TKATCNjCMpquoN0Hc6ffN8Uoc1M86GghLoZ2momjJZICnkYeV8296fiFyziSik/L6RiLdhhEY29EuW14rBG+7AVniSfbkkhUmd3WClLFoFQVEGEwKBgH6FPLpvi/NR7iXRXv1lFftRZHgTp0EMczA0dx9akRuyk7KohqTKmU8sqTqJob8uNuCIUobPeYRAAjazKbAIcmijog5vhXDZXOqkKIsIYbSEqVT4aB4GpH22kMyZkjz/u8GdYzJX+gD4ZLh+x0vPYUuqcUXNlZKMMpaMU77sqAGxAoGBAI/yAGLH/ZwQLGcY77lHWK0QFlwOH8VwZQxez9nsJ+yMiYL6L7u4HH1aOjgEIupwI8B7FHRsbyNBwowVz3w1GCpgcLHfo6upG+89pmF5mmIAd02MAAX04E0qIKytaa13wCZlIY8WKrDurEValgumqEhIqOwiewNvV86q3J8J4REV-----END PRIVATE KEY-----"
	arrowheadKey              = "technician-key.pem"
	arrowheadCert             = "technician-cert.pem"
	arrowheadTruststore       = "truststore.pem"
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

type GeneralContract struct {
	TechnicianID   string   `json:"TechnicianID"`
	MonthlyBalance int      `json:"MonthlyBalance"`
	Jobs           []Job    `json:"Jobs"`
	JobAuthority   []string `json:"JobAuthority"`
}

type OffLedgerResponse struct {
	WorkID    string    `json:"workId"`
	ProductID string    `json:"productId"`
	EventType string    `json:"eventType"`
	Address   string    `json:"address"`
	StartTime time.Time `json:"startTime"`
}

type ServiceLevelResponse struct {
	ServiceLevel string `json:"ServiceLevel"`
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
func (s *SmartContract) TakeJob(ctx contractapi.TransactionContextInterface, jobID string, technichianID string) error {
	fmt.Println("In TakeJob")
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
	jobInfo, err := s.JobExistsOffLedger(jobID, technichianID)
	if err != nil {
		return err
	}

	if jobInfo.WorkID == "" {
		return fmt.Errorf("Job %s does not exist in external system", jobID)
	}

	serviceLevelResponse, err := http.Get("http://35.228.161.184:5001/sla/" + jobInfo.ProductID + "/servicelevel")
	if err != nil {
		return err
	}

	serviceLevelJson, err := io.ReadAll(serviceLevelResponse.Body)
	if err != nil {
		fmt.Println("Eror reading servicelevel")
		return err
	}

	var servicelevel ServiceLevelResponse
	err = json.Unmarshal(serviceLevelJson, &servicelevel)
	fmt.Println("serviceLevel: ", servicelevel)
	if err != nil {
		return err
	}
	switch servicelevel.ServiceLevel {
	case "standard":
		jobInfo.StartTime = jobInfo.StartTime.AddDate(0, 0, 7)

	case "gold":
		jobInfo.StartTime = jobInfo.StartTime.AddDate(0, 0, 5)

	case "platinum":
		jobInfo.StartTime = jobInfo.StartTime.AddDate(0, 0, 3)
	}
	deadline := jobInfo.StartTime.String()
	invokeArgs := [][]byte{[]byte("Create"), []byte(technichianID), []byte(jobID), []byte(jobInfo.ProductID), []byte(jobInfo.Address), []byte(deadline)}
	response := ctx.GetStub().InvokeChaincode(jobInfo.EventType, invokeArgs, ctx.GetStub().GetChannelID())
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

func createDirectory(dirName string) {
	os.Mkdir(dirName, 0777)
}

func createFile(dirName, filename string, content string) error {
	filename = path.Join(dirName, filename)
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file")
		return err
	}
	l, err := f.WriteString(content)
	if err != nil {
		fmt.Println("Error writing to file")
		return err
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()

	if err != nil {
		return err
	}
	return nil

}

func (s *SmartContract) JobExistsOffLedger(jobID string, technicianID string) (*OffLedgerResponse, error) {
	dirName := "tmp"
	createDirectory(dirName)
	err := createFile(dirName, arrowheadCert, arrowheadCertString)
	if err != nil {
		return nil, err
	}

	err = createFile(dirName, arrowheadKey, arrowheadKeyString)
	if err != nil {
		return nil, err
	}

	err = createFile(dirName, arrowheadTruststore, arrowheadTruststoreString)
	if err != nil {
		return nil, err
	}

	serviceRegistryIP := "arrowhead-orchestrator"
	fmt.Println("In JobExistsOffLedger")
	serviceRegistryPort := 8441
	// TODO: check jespers system if the job exists or not and what type of job it is
	var orchBody arrowheadfunctions.Orchestrate
	var orchSystem arrowheadfunctions.System
	orchSystem.Address = "35.228.161.184"
	orchSystem.AuthenticationInfo = ""
	orchSystem.Port = 5000
	orchSystem.SystemName = "technician"
	orchBody.OrchestrationFlags.EnableInterCloud = false
	orchBody.OrchestrationFlags.OverrideStore = false
	orchBody.RequestedService.InterfaceRequirements = []string{"HTTP-SECURE-JSON"}
	orchBody.RequestedService.ServiceDefinitionRequirement = "assign-worker"
	orchBody.RequesterSystem = orchSystem
	orchResponseJSON := arrowheadfunctions.Orchestration(orchBody, serviceRegistryIP, serviceRegistryPort, arrowheadCert, arrowheadKey, arrowheadTruststore)
	var orchResponse arrowheadfunctions.OrchResponse
	json.Unmarshal(orchResponseJSON, &orchResponse)
	fmt.Println("orchResponse: ", orchResponse)
	choosenSystem := orchResponse.Response[0]
	fmt.Println("Choosen system: ", choosenSystem)
	fmt.Println("response from neginfo: ", choosenSystem)

	req, err := http.NewRequest("POST", "https://"+choosenSystem.Provider.Address+":"+strconv.Itoa(choosenSystem.Provider.Port)+choosenSystem.ServiceUri, nil)
	if err != nil {
		fmt.Println("fatal error when creating request")
		log.Fatal(err)
	}

	client := arrowheadfunctions.GetClient("certificates/usercert.pem", "certificates/userkey.pem", "certificates/truststore.pem")
	serviceResp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making HTTP request using client. ", err)
	}
	body, err := io.ReadAll(serviceResp.Body)
	if err != nil {
		fmt.Println("Error reading response body. ", err)
	}
	var assignWorkResponse OffLedgerResponse
	err = json.Unmarshal(body, &assignWorkResponse)
	if err != nil {
		return nil, err
	}

	return &assignWorkResponse, nil
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
