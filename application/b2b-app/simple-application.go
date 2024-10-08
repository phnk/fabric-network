/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/joho/godotenv"
	"github.com/nalle631/arrowheadfunctions"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

const (
	mspID               = "Org1MSP"
	cryptoPath          = "../../test-network/organizations/peerOrganizations/org1.example.com"
	certPath            = cryptoPath + "/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem"
	keyPath             = cryptoPath + "/users/User1@org1.example.com/msp/keystore/"
	tlsCertPath         = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	tlsKeyPath          = cryptoPath + "/peers/peer0.org1.example.com/tls/server.key"
	peerEndpoint        = "localhost:7051"
	gatewayPeer         = "peer0.org1.example.com"
	arrowheadcertsPath  = "./certs"
	arrowheadKey        = arrowheadcertsPath + "/technician-cert.pem"
	arrowheadCert       = arrowheadcertsPath + "/technician-cert.pem"
	arrowheadTruststore = arrowheadcertsPath + "/truststore.pem"
)

type Contract struct {
	Contract *client.Contract
}
type Job struct {
	Type          string    `json:"Type"`
	Status        string    `json:"Status"`
	JobPay        int       `json:"JobPay"`
	InspectionPay int       `json:"InspectionPay"`
	Deadline      time.Time `json:"Deadline,omitempty"`
	ID            string    `json:"ID"`
	Mower         string    `json:"Mower"`
	Address       string    `json:"Adress"`
}

type GeneralContract struct {
	TechnicianID   string   `json:"TechnicianID"`
	MonthlyBalance int      `json:"MonthlyBalance"`
	Jobs           []Job    `json:"Jobs"`
	JobAuthority   []string `json:"JobAuthority"`
}

type TakeJobParams struct {
	JobID string `json:"workId"`
}

type JobDoneParams struct {
	JobID string `json:"JobID"`
}

var technichianID = "Org1MSP"

//var jobID = "9"

func main() {
	godotenv.Load()
	//serviceRegistryIP := os.Getenv("SERVICEREGISTRYADDRESS")
	serviceRegistryIP := "127.0.0.1"
	
	//serviceRegistryPort, err := strconv.Atoi(os.Getenv("SERVICEREGISTRYPORT"))
	serviceRegistryPort, err := strconv.Atoi("8553")
	if err != nil {
		panic(err)
	}
	var rsrDTO arrowheadfunctions.System
	rsrDTO.Address = os.Getenv("SYSTEMADDRESS")
	rsrDTO.AuthenticationInfo = ""
	rsrDTO.Port, err = strconv.Atoi(os.Getenv("SYSTEMPORT"))
	if err != nil {
		panic(err)
	}
	rsrDTO.SystemName = os.Getenv("SYSTEMNAME")
	arrowheadfunctions.RegisterSystem(rsrDTO, serviceRegistryIP, serviceRegistryPort, arrowheadCert, arrowheadKey, arrowheadTruststore)
	var service arrowheadfunctions.Service
	service.Interfaces = []string{os.Getenv("SERVICEINTERFACE")}
	service.Metadata.Method = os.Getenv("SERVICEMETHOD")
	service.ProviderSystem = rsrDTO
	service.Secure = os.Getenv("SERVICESECURE")
	service.ServiceDefinition = os.Getenv("SERVICEDEFINITION")
	service.ServiceUri = os.Getenv("SERVUCEURI")

	arrowheadfunctions.PublishService(service, serviceRegistryIP, serviceRegistryPort, arrowheadCert, arrowheadKey, arrowheadTruststore)
	router := CreateRouter()
	StartRouter(router)

}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	files, err := os.ReadDir(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key directory: %w", err))
	}
	privateKeyPEM, err := os.ReadFile(path.Join(keyPath, files[0].Name()))

	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}

func StartRouter(r *gin.Engine) {
	r.Run(":5000")
	// srv := &http.Server{
	// 	Addr:    ":8080", // Set port number
	// 	Handler: r,
	// }

	// err := srv.ListenAndServeTLS(certPath, keyPath)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// r.RunTLS(":8080", tlsCertPath, tlsKeyPath)
}

func CreateRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/gc", ReadGCHandler)
	r.GET("/gc/jobs", GetAllJobsHandler)
	r.POST("/gc/create", CreateHandler)
	r.POST("/job/create", CreateJobHandler)
	r.POST("/job/take", TakeJobHandler)
	r.POST("/job/done_correct", FinishJobCorrectErrorHandler)
	r.POST("/job/done_wrong", FinishJobWrongErrorHandler)
	return r
}

func Create(contract *client.Contract) {
	fmt.Printf("\n--> Submit Transaction: create, function creates a key value pair on the ledger \n")

	_, err := contract.SubmitTransaction("CreateGeneralContract")
	fmt.Println("err in create", err)
	if err != nil {
		switch err := err.(type) {
		case *client.EndorseError:
			fmt.Printf("Endorse error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
		case *client.SubmitError:
			fmt.Printf("Submit error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
		case *client.CommitStatusError:
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("Timeout waiting for transaction %s commit status: %s", err.TransactionID, err)
			} else {
				fmt.Printf("Error obtaining commit status for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
			}
		case *client.CommitError:
			fmt.Printf("Transaction %s failed to commit with status %d: %s\n", err.TransactionID, int32(err.Code), err)
		default:
			panic(fmt.Errorf("unexpected error type %T: %w", err, err))
		}

		// Any error that originates from a peer or orderer node external to the gateway will have its details
		// embedded within the gRPC status error. The following code shows how to extract that.
		statusErr := status.Convert(err)

		details := statusErr.Details()
		if len(details) > 0 {
			fmt.Println("Error Details:")

			for _, detail := range details {
				switch detail := detail.(type) {
				case *gateway.ErrorDetail:
					fmt.Printf("- address: %s, mspId: %s, message: %s\n", detail.Address, detail.MspId, detail.Message)
				}
			}
		}
	}
	fmt.Printf("*** Transaction committed successfully\n")
}

func CreateHandler(c *gin.Context) {
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	defer gw.Close()

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "gc"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	// chaincodeName2 := "bumpy"

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)

	contract := network.GetContract(chaincodeName)
	Create(contract)
	c.IndentedJSON(http.StatusOK, gin.H{"message": "General contract created"})
}

func createJob(contract *client.Contract, jobID string) {
	fmt.Println("\n--> Submit Transaction: UpdateAsset asset70, asset70 does not exist and should return an error")

	_, err := contract.SubmitTransaction("Create", technichianID, jobID, "5", "Tomoko", "300")

	switch err := err.(type) {
	case *client.EndorseError:
		fmt.Printf("Endorse error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
	case *client.SubmitError:
		fmt.Printf("Submit error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
	case *client.CommitStatusError:
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("Timeout waiting for transaction %s commit status: %s", err.TransactionID, err)
		} else {
			fmt.Printf("Error obtaining commit status for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
		}
	case *client.CommitError:
		fmt.Printf("Transaction %s failed to commit with status %d: %s\n", err.TransactionID, int32(err.Code), err)
	default:
		panic(fmt.Errorf("unexpected error type %T: %w", err, err))
	}

	// Any error that originates from a peer or orderer node external to the gateway will have its details
	// embedded within the gRPC status error. The following code shows how to extract that.
	statusErr := status.Convert(err)

	details := statusErr.Details()
	if len(details) > 0 {
		fmt.Println("Error Details:")

		for _, detail := range details {
			switch detail := detail.(type) {
			case *gateway.ErrorDetail:
				fmt.Printf("- address: %s, mspId: %s, message: %s\n", detail.Address, detail.MspId, detail.Message)
			}
		}
	}
}

func CreateJobHandler(c *gin.Context) {
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	id1 := id.Credentials()
	fmt.Println("id1: ", string(id1[:]))
	fmt.Println("mspID: ", id.MspID())
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	defer gw.Close()

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "gc"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	// chaincodeName2 := "bumpy"

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)

	contract := network.GetContract(chaincodeName)
	createJob(contract, c.Param("jobID"))
	c.IndentedJSON(http.StatusOK, gin.H{"message": "job created"})
}

// Submit a transaction to query ledger state.
func takeJob(contract *client.Contract, jobID string) {
	fmt.Println("\n--> Submit Transaction: TakeJob, function updates a key value pair on the ledger \n")

	fmt.Println("jobID: ", jobID)

	//Remember to remove jobtype when integrated with jespers system
	submitResult, err := contract.SubmitTransaction("TakeJob", jobID, technichianID)
	fmt.Println("err: ", err, status.Code(err))
	if err != nil {
		switch err := err.(type) {
		case *client.EndorseError:
			fmt.Printf("Endorse error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
		case *client.SubmitError:
			fmt.Printf("Submit error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
		case *client.CommitStatusError:
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("Timeout waiting for transaction %s commit status: %s", err.TransactionID, err)
			} else {
				fmt.Printf("Error obtaining commit status for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
			}
		case *client.CommitError:
			fmt.Printf("Transaction %s failed to commit with status %d: %s\n", err.TransactionID, int32(err.Code), err)
		default:
			panic(fmt.Errorf("unexpected error type %T: %w", err, err))
		}

		// Any error that originates from a peer or orderer node external to the gateway will have its details
		// embedded within the gRPC status error. The following code shows how to extract that.
		statusErr := status.Convert(err)

		details := statusErr.Details()
		if len(details) > 0 {
			fmt.Println("Error Details:")

			for _, detail := range details {
				switch detail := detail.(type) {
				case *gateway.ErrorDetail:
					fmt.Printf("- address: %s, mspId: %s, message: %s\n", detail.Address, detail.MspId, detail.Message)
				}
			}
		}
	}

	fmt.Println("Result:", submitResult)
}

func TakeJobHandler(c *gin.Context) {
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	id1 := id.Credentials()
	fmt.Println("id1: ", string(id1[:]))
	fmt.Println("mspID: ", id.MspID())
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	defer gw.Close()

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "gc"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	// chaincodeName2 := "bumpy"

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)

	contract := network.GetContract(chaincodeName)

	var params TakeJobParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	takeJob(contract, params.JobID)
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Job added to your general contract."})
}

func finishJobCorrectError(contract *client.Contract, jobID string) {
	fmt.Println("\n--> Submit Transaction: Finish job correct error, function updates a key value pair on the ledger \n")

	submitResult, err := contract.SubmitTransaction("JobDoneCorrectError", jobID)
	if err != nil {
		switch err := err.(type) {
		case *client.EndorseError:
			fmt.Printf("Endorse error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
		case *client.SubmitError:
			fmt.Printf("Submit error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
		case *client.CommitStatusError:
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("Timeout waiting for transaction %s commit status: %s", err.TransactionID, err)
			} else {
				fmt.Printf("Error obtaining commit status for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
			}
		case *client.CommitError:
			fmt.Printf("Transaction %s failed to commit with status %d: %s\n", err.TransactionID, int32(err.Code), err)
		default:
			panic(fmt.Errorf("unexpected error type %T: %w", err, err))
		}

		// Any error that originates from a peer or orderer node external to the gateway will have its details
		// embedded within the gRPC status error. The following code shows how to extract that.
		statusErr := status.Convert(err)

		details := statusErr.Details()
		if len(details) > 0 {
			fmt.Println("Error Details:")

			for _, detail := range details {
				switch detail := detail.(type) {
				case *gateway.ErrorDetail:
					fmt.Printf("- address: %s, mspId: %s, message: %s\n", detail.Address, detail.MspId, detail.Message)
				}
			}
		}
	}

	fmt.Println("Result:", submitResult)
}

func FinishJobCorrectErrorHandler(c *gin.Context) {
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	id1 := id.Credentials()
	fmt.Println("id1: ", string(id1[:]))
	fmt.Println("mspID: ", id.MspID())
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	defer gw.Close()

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "gc"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	// chaincodeName2 := "bumpy"

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)

	contract := network.GetContract(chaincodeName)

	var params JobDoneParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	finishJobCorrectError(contract, params.JobID)
	c.IndentedJSON(http.StatusOK, gin.H{"message": "finished job with correct error"})
}

func finishJobWrongError(contract *client.Contract, jobID string) {
	fmt.Println("\n--> Submit Transaction: FinishJob wrong error, function updates a key value pair on the ledger \n")

	submitResult, err := contract.SubmitTransaction("JobDoneWrongError", jobID)
	if err != nil {
		switch err := err.(type) {
		case *client.EndorseError:
			fmt.Printf("Endorse error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
		case *client.SubmitError:
			fmt.Printf("Submit error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
		case *client.CommitStatusError:
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("Timeout waiting for transaction %s commit status: %s", err.TransactionID, err)
			} else {
				fmt.Printf("Error obtaining commit status for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
			}
		case *client.CommitError:
			fmt.Printf("Transaction %s failed to commit with status %d: %s\n", err.TransactionID, int32(err.Code), err)
		default:
			panic(fmt.Errorf("unexpected error type %T: %w", err, err))
		}

		// Any error that originates from a peer or orderer node external to the gateway will have its details
		// embedded within the gRPC status error. The following code shows how to extract that.
		statusErr := status.Convert(err)

		details := statusErr.Details()
		if len(details) > 0 {
			fmt.Println("Error Details:")

			for _, detail := range details {
				switch detail := detail.(type) {
				case *gateway.ErrorDetail:
					fmt.Printf("- address: %s, mspId: %s, message: %s\n", detail.Address, detail.MspId, detail.Message)
				}
			}
		}
	}

	fmt.Println("Result:", submitResult)
}

func FinishJobWrongErrorHandler(c *gin.Context) {
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	id1 := id.Credentials()
	fmt.Println("id1: ", string(id1[:]))
	fmt.Println("mspID: ", id.MspID())
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	defer gw.Close()

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "gc"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	// chaincodeName2 := "bumpy"

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)

	contract := network.GetContract(chaincodeName)
	var params JobDoneParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	finishJobWrongError(contract, params.JobID)
	c.IndentedJSON(http.StatusOK, gin.H{"message": "finished job with wrong error"})
}

// Evaluate a transaction by key to query ledger state.
func ReadGC(contract *client.Contract) *GeneralContract {
	fmt.Printf("\n--> Evaluate Transaction: Read, function returns key value pair\n")

	evaluateResult, err := contract.EvaluateTransaction("ReadGeneralContract", technichianID)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	var gc GeneralContract
	err = json.Unmarshal(evaluateResult, &gc)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal result: %w", err))
	}

	fmt.Println("Result: ", gc)

	return &gc
}

func ReadGCHandler(c *gin.Context) {
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	id1 := id.Credentials()
	fmt.Println("id1: ", string(id1[:]))
	fmt.Println("mspID: ", id.MspID())
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	defer gw.Close()

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "gc"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	// chaincodeName2 := "bumpy"

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)

	contract := network.GetContract(chaincodeName)
	readResult := ReadGC(contract)
	c.IndentedJSON(http.StatusOK, readResult)
}

func readJob(contract *client.Contract, jobID string) {
	fmt.Printf("\n--> Evaluate Transaction: Read, function returns key value pair\n")

	evaluateResult, err := contract.EvaluateTransaction("ReadJob", jobID)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	fmt.Println("Result: ", string(evaluateResult[:]))
}

func getAllJobs(contract *client.Contract) ([]byte, error) {
	fmt.Printf("\n--> Evaluate Transaction: Read, function returns key value pair\n")

	evaluateResult, err := contract.EvaluateTransaction("GetAllJobs")
	if err != nil {
		return nil, err
	}

	fmt.Println("Result: ", string(evaluateResult[:]))

	return evaluateResult, nil
}

func GetAllJobsHandler(c *gin.Context) {
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	id1 := id.Credentials()
	fmt.Println("id1: ", string(id1[:]))
	fmt.Println("mspID: ", id.MspID())
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	defer gw.Close()

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "gc"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	// chaincodeName2 := "bumpy"

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)

	contract := network.GetContract(chaincodeName)
	result, err := getAllJobs(contract)
	if err != nil {
		c.IndentedJSON(400, "Couln't get all jobs")
	}
	c.IndentedJSON(http.StatusOK, result)
}

// Submit transaction, passing in the wrong number of arguments ,expected to throw an error containing details of any error responses from the smart contract.
func exampleErrorHandling(contract *client.Contract) {
	fmt.Println("\n--> Submit Transaction: UpdateAsset asset70, asset70 does not exist and should return an error")

	_, err := contract.SubmitTransaction("UpdateAsset", "asset70", "blue", "5", "Tomoko", "300")
	if err == nil {
		panic("******** FAILED to return an error")
	}

	fmt.Println("*** Successfully caught the error:")

	switch err := err.(type) {
	case *client.EndorseError:
		fmt.Printf("Endorse error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
	case *client.SubmitError:
		fmt.Printf("Submit error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
	case *client.CommitStatusError:
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("Timeout waiting for transaction %s commit status: %s", err.TransactionID, err)
		} else {
			fmt.Printf("Error obtaining commit status for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
		}
	case *client.CommitError:
		fmt.Printf("Transaction %s failed to commit with status %d: %s\n", err.TransactionID, int32(err.Code), err)
	default:
		panic(fmt.Errorf("unexpected error type %T: %w", err, err))
	}

	// Any error that originates from a peer or orderer node external to the gateway will have its details
	// embedded within the gRPC status error. The following code shows how to extract that.
	statusErr := status.Convert(err)

	details := statusErr.Details()
	if len(details) > 0 {
		fmt.Println("Error Details:")

		for _, detail := range details {
			switch detail := detail.(type) {
			case *gateway.ErrorDetail:
				fmt.Printf("- address: %s, mspId: %s, message: %s\n", detail.Address, detail.MspId, detail.Message)
			}
		}
	}
}

// Format JSON data
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}
