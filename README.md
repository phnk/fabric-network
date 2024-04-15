# Code structure
There are three main parts of the implementation. The test-network, the chaincode and the application. Each of these found in their respective folder on github.
## test-network
The test-network part of the implementation is to ease the testing and use of the Hyperledger Fabric Network. This is done by the network.sh script which in turn uses the scripts located in the scripts folder. The network.sh script is used for testing and developing a hyperledger fabric network. It can be used for many purposes however the most important ones for this thesis is:
* Up: Creates a hyperledger fabric network with two organisations and their certificates.
* Down: Brings down the network and all organisations peer, meaning that all channels deployed chaincode is removed.
* createChannel: Creates a channel on the Fabric Network and joins the two organisations created at initialization of the project to the channel.
* DeployCC: Deploys a chaincode to the specified channel

## Chaincode
A chaincode is self executing code that exists on the ledger. Chaincode is Hyperledger Fabrics take on smart contracts (briefly mentioned in 3.2.2) and it is the chaincode that is responsible for the functionality of the contracts. Functionality meaning the modification and calculations of the contracts stored on the Fabric blockchain. For this thesis chaincodes are devided into two categories. Business-to-Business chaincode and Customer-to-Business chaincode.
### Business-to-Business
Business-to-Business chaincode are made for the interaction between service-providers and the service-owner. There are two different levels to them. The first one is the job-contract chaincode which creates a General Contract. The General Contract handles everything related to the service-provider, for example the monthly payout, the services that the service-provider have, how a service-provider takes on a service and how they can confirm that a service is completed. There can only be one General Contract for each service-provider organisation and the id for the contract is automaticly set to the organisations MSP (membership service provider) id. The second level is service chaincode.
A service chaincode represent a service that are available for a service-provider. The service chaincode is responisble for creating and managing a service contract. These are created from within a general contract when a service provider get assigned to a service. The structure of the service contracts can be seen in the figure below.
<p align="center">
  <img src="img/service_structure.png" />
</p>
If a new type of service is availabe a corresponding chaincode is created for that service, thus new services can be added on demand in the Fabric network. A sequence diagram of how a General Contract is created and how a job is taken can be seen in the image below:
<p align="center">
  <img src="img/TakeSequence.png" />
</p>

### Customer-to-Business
Customer-to-Business chaincode are created for the interaction between service-buyers and the service-owner. Simillar to the Business-to-Business chaincode, there exists two levels of chaincode. The customer chaincode and the SLA chaincode. The customer chaincode are responsible for managing the customer contract. A customer contract contains the customer id and all their active SLA:s. When a customer buys a service the customer contract creates a new SLA for the service and adds it to the contract. The process of registering a customer and buying a service can be seen in the sequence diagram below.
<p align="center">
  <img src="img/BuySequence.png" />
</p>

More information about the Customer-to-Business chaincodes can be found on the projects github in the chaincode folder. There, all the functionalities of the chaincodes can be studied.

## Application
Applications are used outside of the Fabric network with the main functionality of interacting with the chaincode. Each organization partisipating in the Fabric network are required to implement their own application. This means that each service-provider owns their own application wich uses their own crypographic identification and certificates. In this thesis two applications has been created, one for the customer organisation and one for a service provider organisation. These can be referenced to while creating new applications for new organisations, however they should only be used for testing since they use simple cryptographic identification and certificates.

### B2B-Application
The B2B-app is a REST API that are used by a service-provider to interact with their General Contract. The B2B-app in this thesis is only created for one service-provider meaning that if a service-provider wants to join the Fabric Network, they have to create their own application using the organisations cryptographic credentials and certificates. The endpoints that the service-provider can be seen in the image below.
<p align="center">
  <img src="img/b2bEndpoints.png" />
</p>

For example if a service-provider wants to take on a job/service they use the /job/take endpoint which will tell the General Contract to create a new service should the service not already be taken by another service-provider. The identification for each service-provider is their MSPID which corresponds to their organisations MSP and is handled within the chaincode.



### C2B-Application
The C2B-App is a REST API that handles the customers interactions with the fabric network. The endpoints that the customers can be used to interact with the Fabric Network can be seen in the image below.
<p align="center">
  <img src="img/c2bEndpoints.png" />
</p>
For example when a customer wants to buy a service it should send their request to the :customer_id/sla endpoint which in turn will invoke the customer contract chaincode mentioned in the chaincode section. Since there are only one customer organisation there is only one application required for all customers. This means however that the identification of a customer is done with a customers id contrary to the identification of service-providers mentioned above.





