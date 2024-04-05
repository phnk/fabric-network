# Code structure
There are three main parts of the implementation. The test-network, the chaincode and the application. Each of these found in their respective folder.
## test-network
The test-network part of the implementation is to ease the testing and use of the Hyperledger Fabric Network. This is done by the network.sh script which in turn uses the scripts located in the scripts folder. The network.sh script is used for testing and developing a hyperledger fabric network. It can be used for many purposes however the most important ones for this thesis is:
* Up: Creates a hyperledger fabric network with two organisations and their certificates.
* Down: Brings down the network and all organisations peer, meaning that all channels deployed chaincode is removed.
* createChannel: Creates a channel on the Fabric Network and joins the two organisations created at initialization of the project to the channel.
* DeployCC: Deploys a chaincode to the specified channel

## Chaincode
Chaincodes are devided into two categories. Business-to-Business chaincode and Customer-to-Business chaincode.
### Business-to-Business
Business-to-Business chaincode are made for the interaction between service-providers and the service-owner. There are two different levels to them. The first one is the job-contract chaincode which creates a General Contract. The General Contract handles everything related to the service-provider, for example the monthly payout, the services that the service-provider have, how a service-provider takes on a service and how they can confirm that a service is completed. There can only be one General Contract for each service-provider organisation and the id for the contract is automaticly set to the organisations MSP (membership service provider) id. The second level is service chaincode.
A service chaincode represent a service that are available for a service-provider. The service chaincode is responisble for creating and managing a service contract. These are created from within a general contract when a service provider get assigned to a service. The structure of the service contracts can be seen in the figure below.
<p align="center">
  <img src="img/service_structure" />
</p>

