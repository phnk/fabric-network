package bumpy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/nalle631/arrowheadfunctions"
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

func createFile(dirName, filename string, content string) (string, error) {
	filename = path.Join(dirName, filename)
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file")
		return "", err
	}
	l, err := f.WriteString(content)
	if err != nil {
		fmt.Println("Error writing to file")
		return "", err
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()

	if err != nil {
		return "", err
	}
	return filename, nil

}

func createDirectory(dirName string) {
	os.Mkdir(dirName, 0777)
}

func (s *SmartContract) JobExistsOffLedger(jobID string, technicianID string) (bool, error) {
	dirName := "tmp"
	createDirectory(dirName)
	certPath, err := createFile(dirName, arrowheadCert, arrowheadCertString)
	if err != nil {
		return false, err
	}

	keyPath, err := createFile(dirName, arrowheadKey, arrowheadKeyString)
	if err != nil {
		return false, err
	}

	trustsorePath, err := createFile(dirName, arrowheadTruststore, arrowheadTruststoreString)
	if err != nil {
		return false, err
	}
	orchIP := "arrowhead-orchestrator"
	orchPort := 8441
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
	jsonOrchResponse := arrowheadfunctions.Orchestration(requestBody, orchIP, orchPort, certPath, keyPath, trustsorePath)
	json.Unmarshal(jsonOrchResponse, &orchResponse)
	chosenResponse := orchResponse.Response[0]
	fmt.Println("response from neginfo: ", chosenResponse)

	req, err := http.NewRequest("POST", "https://"+chosenResponse.Provider.Address+":"+strconv.Itoa(chosenResponse.Provider.Port)+chosenResponse.ServiceUri, nil)
	if err != nil {
		return false, err
	}

	client := arrowheadfunctions.GetClient(certPath, keyPath, trustsorePath)
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
