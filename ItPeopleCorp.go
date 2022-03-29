package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	//"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	//"github.com/hyperledger/fabric-chaincode-go/shim"
	//"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid" //oracle
	"github.com/hyperledger/fabric/core/chaincode/shim"    //oracle
	"github.com/hyperledger/fabric/protos/peer"            //oracle
)

var CAR_ASSET = "CAR"

//ROLES used
var MANUFACTURER_ROLE = "MANUFACTURER"
var DEALER_ROLE = "DEALER"
var CUSTOMER_ROLE = "CUSTOMER"

//STATUS used
var CREATED = "CREATED"
var READY_FOR_SALE = "READY_FOR_SALE"
var SOLD = "SOLD"

//MSPIDs used
var MANUFACTURER_MSPID = "mfg-instance"
var DEALER_MSPID = "dealer-instance"

//  Chaincode implementation
type ItPeopleCorp struct {
}

//  CarAsset
type CarAsset struct {
	CarId        string `json:"carId"`
	ChasisNumber string `json:"chasisNumber"`
	Manufacturer string `json:"manufacturer"`
	Dealer       string `json:"dealer"`
	Customer     string `json:"customer"`
	Address      string `json:"address"`
	Status       string `json:"status"`
	DateTime     string `json:"dateTime"`
	ObjectType   string `json:"docType"` //docType is used to distinguish the various types of objects in state database
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(ItPeopleCorp))
	if err != nil {
		fmt.Printf("Error starting Xebest Trace chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *ItPeopleCorp) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *ItPeopleCorp) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "createCar" { // when manufactuer creates car, this function will be called
		return t.createCar(stub, args)
	} else if function == "deliverToDealer" { // this function will be called once manufacturer deliver car to dealer
		return t.deliverToDealer(stub, args)
	} else if function == "saleToCustomer" { // once the car is sold to customer, this function will be called
		return t.saleToCustomer(stub, args)
	} else if function == "queryRecord" { // query a record
		return t.queryRecord(stub, args)
	} else if function == "del" { // query a record
		return t.del(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ===========================================================
// del
// ===========================================================
func (t *ItPeopleCorp) del(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	id := args[0]
	err := stub.DelState(id)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// ===============================================
// queryByCert_id - read a certificate from chaincode state
// ===============================================
func (t *ItPeopleCorp) queryRecord(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var id, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting id of the asset to query")
	}

	id = args[0]
	valAsbytes, err := stub.GetState(id) //get the certificate from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + id + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"certificate does not exist: " + id + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

func (t *ItPeopleCorp) createCar(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//   args[0]      args[1]         args[2]        args[3]
	//ChasisNumber   Manufacturer     Address         Role

	if len(args) != 4 {
		return shim.Error("##Incorrect number of arguments. expecting 4 args")
	}

	if args[3] != MANUFACTURER_ROLE {
		return shim.Error("This Role is not authorised to do this operation")
	}

	// ==== Permissioning and Pre-Existance ====
	mspid, errstr := GetMSPID(stub)
	if errstr != "" {
		return shim.Error(errstr)
	}

	fmt.Println("mspid ::" + mspid)

	//if mspid != MANUFACTURER_MSPID {
	//	return shim.Error("This Role is not authorised to do this operation")
	//}

	txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)
	if errTx != nil {
		return shim.Error("Returning error")
	}

	carId := ComputeHashKey(args[0] + args[1] + (txTimeAsPtr.String()))

	record := &CarAsset{carId, args[0], args[1], "", "", args[2], CREATED, txTimeAsPtr.String(), CAR_ASSET}
	recordJSONasBytes, err := json.Marshal(record)
	if err != nil {
		return shim.Error("unable to marshal car asset JSON")
	}

	err1 := stub.PutState(carId, recordJSONasBytes)
	if err1 != nil {
		return shim.Error("Failed to push car asset record into state")
	}

	fmt.Println("Car asset records is pushed ::" + string(recordJSONasBytes))

	return shim.Success(recordJSONasBytes)
}

func (t *ItPeopleCorp) deliverToDealer(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//   args[0]      args[1]         args[2]        args[3]
	//    carId       Dealer          Address         Role

	if len(args) != 4 {
		return shim.Error("##Incorrect number of arguments. expecting 4 args")
	}

	if args[3] != DEALER_ROLE {
		return shim.Error("This Role is not authorised to do this operation")
	}

	// ==== Permissioning and Pre-Existance ====
	mspid, errstr := GetMSPID(stub)
	if errstr != "" {
		return shim.Error(errstr)
	}
	fmt.Println("mspid ::" + mspid)
	//	if mspid != DEALER_MSPID {
	//		return shim.Error("This Role is not authorised to do this operation")
	//	}

	txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)
	if errTx != nil {
		return shim.Error("Returning error")
	}

	carId = args[0]
	convalAsbytes, conerr := stub.GetState(carId) //get the contractor from chaincode state
	if conerr != nil {
		return shim.Error(conerr.Error())
	} else if convalAsbytes == nil {
		return shim.Error("Car Id with this id does not exist")
	} else {
		carAsset := &CarAsset{}
		err7 := json.Unmarshal([]byte(convalAsbytes), &carAsset)
		if err7 != nil {
			return shim.Error(err7.Error())
		}
		carAsset.Dealer = args[1]
		carAsset.Address = args[2]
		carAsset.Status = READY_FOR_SALE

		convalAsbytes, err8 := json.Marshal(carAsset)
		if err8 != nil {
			return shim.Error(err8.Error())
		}
		err9 := stub.PutState(carId, convalAsbytes)
		if err9 != nil {
			return shim.Error(err9.Error())
		}

		fmt.Println("contractor statistics asset is updated..")
	}

	fmt.Println("Car asset records is pushed ::" + string(convalAsbytes))

	return shim.Success(convalAsbytes)
}

func (t *ItPeopleCorp) saleToCustomer(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//   args[0]      args[1]         args[2]        args[3]
	//    carId       Customer          Address         Role

	if len(args) != 4 {
		return shim.Error("##Incorrect number of arguments. expecting 4 args")
	}

	if args[3] != DEALER_ROLE {
		return shim.Error("This Role is not authorised to do this operation")
	}

	// ==== Permissioning and Pre-Existance ====
	mspid, errstr := GetMSPID(stub)
	if errstr != "" {
		return shim.Error(errstr)
	}
	fmt.Println("mspid ::" + mspid)
	//	if mspid != DEALER_MSPID {
	//		return shim.Error("This Role is not authorised to do this operation")
	//	}

	txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)
	if errTx != nil {
		return shim.Error("Returning error")
	}

	carId = args[0]
	convalAsbytes, conerr := stub.GetState(carId) //get the contractor from chaincode state
	if conerr != nil {
		return shim.Error(conerr.Error())
	} else if convalAsbytes == nil {
		return shim.Error("Car Id with this id does not exist")
	} else {
		carAsset := &CarAsset{}
		err7 := json.Unmarshal([]byte(convalAsbytes), &carAsset)
		if err7 != nil {
			return shim.Error(err7.Error())
		}
		carAsset.Customer = args[1]
		carAsset.Address = args[2]
		carAsset.Status = SOLD
		convalAsbytes, err8 := json.Marshal(carAsset)
		if err8 != nil {
			return shim.Error(err8.Error())
		}
		err9 := stub.PutState(carId, convalAsbytes)
		if err9 != nil {
			return shim.Error(err9.Error())
		}

		fmt.Println("contractor statistics asset is updated..")
	}

	fmt.Println("Car asset records is pushed ::" + string(convalAsbytes))

	return shim.Success(convalAsbytes)
}

// GetTxTimestampChannel Function gets the Transaction time when the chain code was executed it remains same on all the peers where chaincode executes
func (t *ItPeopleCorp) GetTxTimestampChannel(stub shim.ChaincodeStubInterface) (time.Time, error) {
	txTimeAsPtr, err := stub.GetTxTimestamp()
	if err != nil {
		fmt.Printf("Returning error in TimeStamp \n")
		//	return "Error",err
	}
	fmt.Printf("\t returned value from APIstub: %v\n", txTimeAsPtr)
	//timeStr := time.Unix(txTimeAsPtr.Seconds, int64(txTimeAsPtr.Nanos)).String()
	timeInt := time.Unix(txTimeAsPtr.Seconds, int64(txTimeAsPtr.Nanos))

	return timeInt, nil
}

func ComputeHashKey(propertyName string) string {
	h := sha256.New()
	h.Write([]byte(propertyName))
	nameInBytes := h.Sum([]byte(""))
	nameInString := hex.EncodeToString(nameInBytes)
	return nameInString[:62]
	//Creates a key of the form: "621def" + string(sha256(schema_name))
}

// Returns MSPID
func GetMSPID(stub shim.ChaincodeStubInterface) (string, string) {
	id, err := cid.New(stub)
	if err != nil {
		return "", "Unable to return client id"
	}
	mspid, err := id.GetMSPID()
	if err != nil {
		return "", "Unable to return MSP id"
	} else {
		return mspid, ""
	}
}

// Below function provide Role and Group Detail
// func getUserRoleAndGroup(stub shim.ChaincodeStubInterface) UserDetails {

// 	fmt.Println("In getUserRoleAndGroup")

// 	var user UserDetails
// 	// Code for Transient Map TO Know the Rest API Caller ID
// 	if transientMap, err := stub.GetTransient(); err == nil {

// 		keys := reflect.ValueOf(transientMap).MapKeys()
// 		strkeys := make([]string, len(keys))
// 		for i := 0; i < len(keys); i++ {
// 			strkeys[i] = keys[i].String()
// 		}
// 		fmt.Println("All the Keys")
// 		fmt.Println(strings.Join(strkeys, ","))

// 		if userRole, ok := transientMap["UserRole"]; ok {
// 			user.UserRole = string(userROle[:])
// 			fmt.Println("User Role: %v", string(userRole[:]))

// 		} else {
// 			fmt.Println("Does not exists bcsRestClientId")
// 		}
// 		if userGroup, ok1 := transientMap["Group"]; ok1 {
// 			user.UserGroup = string(userGroup[:])
// 			fmt.Println("User Group: %v", string(userGroup[:]))

// 		} else {
// 			fmt.Println("Does not exists Group")
// 		}

// 	}
// 	return user
// }
