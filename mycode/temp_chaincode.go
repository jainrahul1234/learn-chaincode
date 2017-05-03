package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"

)

//==============================================================================================================================
//	 Structure Definitions
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type  SimpleChaincode struct {
}

//==============================================================================================================================
//	Package - Defines the structure for a package object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type PackageInfo struct {
  PkgId      string `json:"packageid"`
  Shipper    string `json:"shipper"`
  Insurer    string `json:"insurer"`
  Consignee  string `json:"consignee"`
  Owner      string `json:"owner"`
  TempratureMin int `json:"Tempraturemin"`
  TempratureMax int `json:"Tempraturemax"`
  PackageDes string `json:"packagedes"`
  PkgStatus  string `json:"pkgstatus"`
}

//==============================================================================================================================
//	 PkgStatus types - Asset lifecycle is broken down into 4 statuses, this is part of the business logic to determine what can
//					be done to the package at points in it's lifecycle
//==============================================================================================================================
//  1 - Label_Generated
//  2 - In_Transit
//  3 - Pkg_Damaged
//  4 - Pkg_Delivered


//==============================================================================================================================
//	Package Holder - Defines the structure that holds all the PkgId for Packages that have been created.
//				Used as an index when querying all vehicles.
//==============================================================================================================================
type PKG_Holder struct {
	PkgIds 	[]string `json:"packageids"`
}

//=================================================================================================================================
//	 Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {
err := shim.Start(new(SimpleChaincode))
if err != nil {
  fmt.Printf("Error starting Simple chaincode: %s", err)
  }
}


//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

var jsonResp string
var packageinfo PackageInfo
var err error

//  Validate inpit
if len(args) != 6 {
  jsonResp = "{\"Error\":\"Incorrect number of arguments. Expecting 6 in order of Shipper, Insurer, Consignee, Temprature, PackageDes, Owner\"}"
  return nil, errors.New(jsonResp)
  }


//  Polulating JSON block with input for first block
packageinfo.PkgId = "1Z20170426"
packageinfo.Shipper = args[0]
packageinfo.Insurer  = args[1]
packageinfo.Consignee  = args[2]
packageinfo.Owner = args[3]
packageinfo.TempratureMin, err = strconv.Atoi(args[4])
if err != nil {
  jsonResp = "{\"Error\":\"5th argument must be a numeric string\"}"
  return nil, errors.New(jsonResp)
	}
packageinfo.TempratureMax, err = strconv.Atoi(args[5])
if err != nil {
    jsonResp = "{\"Error\":\"6th argument must be a numeric string\"}"
    return nil, errors.New(jsonResp)
  	}
packageinfo.PackageDes = args[6]
packageinfo.PkgStatus = "Label_Generated"


//  populate package holder
var packageids_array PKG_Holder
packageids_array.PkgIds  = append(packageids_array.PkgIds , packageinfo.PkgId)

bytes, err := json.Marshal(&packageids_array)

//  write to blockchain
err = stub.PutState("PkgIdsKey", bytes)
if err != nil {
  return nil, errors.New("Error writing to blockchain for PKG_Holder")
  }

bytes, err = json.Marshal(&packageinfo)
if err != nil {
          fmt.Println("Could not marshal personal info object", err)
          return nil, errors.New("Could not marshal personal info object")
  }

//  write to blockchain
err = stub.PutState("1Z20170426", bytes)
if err != nil {
  return nil, errors.New("Error writing to blockchain for Package")
  }

return nil, nil
}


//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> create
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
fmt.Println("invoke is running " + function)

// Handle different functions
if function == "create" {
  return t.create(stub,args)
  } else if function == "deliverpkg"{
  return t.deliverpkg(stub,args)
  } else if function == "updatetemp" {
  return t.updatetemp(stub, args)
  }

fmt.Println("invoke did not find func: " + function)
return nil, errors.New("Received unknown function invocation: " + function)

}


//=================================================================================================================================
//	create - create new package on a block
//=================================================================================================================================
func (t *SimpleChaincode) create(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
fmt.Println("running create()")

var key, jsonResp string
var err error

if len(args) != 8 {
  jsonResp = "{\"Error\":\"Incorrect number of arguments. Expecting 8 in order of PkgID, Shipper, Insurer, Consignee, TempratureMin, TempratureMax, PackageDes, Owner\"}"
  return nil, errors.New(jsonResp)
  }

var packageinfo PackageInfo

key = args[0]

packageinfo.PkgId  = args[0]
packageinfo.Shipper = args[1]
packageinfo.Insurer = args[2]
packageinfo.Consignee  = args[3]
packageinfo.TempratureMin , err = strconv.Atoi(args[4])
if err != nil {
  jsonResp = "{\"Error\":\"5th argument must be a numeric string\"}"
  return nil, errors.New(jsonResp)
	}
packageinfo.TempratureMax  , err = strconv.Atoi(args[5])
if err != nil {
  jsonResp = "{\"Error\":\"5th argument must be a numeric string\"}"
  return nil, errors.New(jsonResp)
	}
packageinfo.PackageDes = args[6]
packageinfo.Owner = args[7]
packageinfo.PkgStatus = "Label_Generated"   // Label_Generated

bytes, err := json.Marshal(&packageinfo)
if err != nil {
        fmt.Println("Could not marshal personal info object", err)
        return nil, err
  }

//  populate package holder
var packageids_array PKG_Holder
packageids_arrayasbytes, err := stub.GetState("PkgIdsKey")

err = json.Unmarshal(packageids_arrayasbytes, &packageids_array)
if err != nil {
      fmt.Println("Could not marshal pkgid array object", err)
      return nil, err
  }

packageids_array.PkgIds  = append(packageids_array.PkgIds , packageinfo.PkgId)
packageids_arrayasbytes, err = json.Marshal(&packageids_array)

//  write to blockchain
err = stub.PutState("PkgIdsKey", packageids_arrayasbytes)
if err != nil {
    return nil, errors.New("Error writing to blockchain for PKG_Holder")
  }

err = stub.PutState(key, bytes)
if err != nil {
  return nil, err
  }

return nil, nil
}

//=================================================================================================================================
//	deliverpkg - deliver package to cosignee, change owner of package
//=================================================================================================================================
func (t *SimpleChaincode) deliverpkg(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
fmt.Println("running deliverpkg()")
var key , jsonResp string
var err error

if len(args) != 2 {
	jsonResp = "{\"Error\":\"Incorrect number of arguments. Expecting 2 : PkgId and New Owner\"}"
  	return nil, errors.New(jsonResp)
  }

  key = args[0]
  var packageinfo PackageInfo

  valAsbytes, err := stub.GetState(key)

  if err != nil {
    jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
    return nil, errors.New(jsonResp)
    }

  err = json.Unmarshal(valAsbytes, &packageinfo)
  if err != nil {
        fmt.Println("Could not marshal personal info object", err)
        return nil, err
    }

// validate pkd exist or not by checking temprature
  if packageinfo.PkgId != key{
    jsonResp = "{\"Error\":\"Invalid PackageId Passed\"}"
    return nil, errors.New(jsonResp)
    }

  // check wheather the pkg temprature is in acceptable range and package in in valid status
  if packageinfo.PkgStatus == "Pkg_Damaged" {    // Pkg_Damaged
	  jsonResp = "{\"Error\":\"Temprature thershold crossed - Package Damaged\"}"
          return nil, errors.New(jsonResp)
    }

  packageinfo.Owner = args[1]
  packageinfo.PkgStatus = "Pkg_Delivered"

  bytes, err := json.Marshal(&packageinfo)
  if err != nil {
          fmt.Println("Could not marshal personal info object", err)
          return nil, err
    }

  err = stub.PutState(key, bytes)
  if err != nil {
    return nil, err
    }

  return nil, nil
}




//=================================================================================================================================
//	updatetemp - update pkg status based on the supplied temprature
//=================================================================================================================================
func (t *SimpleChaincode) updatetemp(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
var key , jsonResp string
var err error
fmt.Println("running updatetemp()")

if len(args) != 2 {
  jsonResp = "{\"Error\":\"Incorrect number of arguments. Expecting 2. name of the key and temprature value to set\"}"
  return nil, errors.New(jsonResp)
  }


key = args[0]
var packageinfo PackageInfo
var temprature_reading int

valAsbytes, err := stub.GetState(key)

if err != nil {
  jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
  return nil, errors.New(jsonResp)
  }

err = json.Unmarshal(valAsbytes, &packageinfo)
if err != nil {
      fmt.Println("Could not marshal  info object", err)
      return nil, err
  }
// validate pkd exist or not by checking temprature
if packageinfo.PkgId != key{
  jsonResp = "{\"Error\":\"Invalid PackageId Passed\"}"
  return nil, errors.New(jsonResp)
  }

// check wheather the pkg temprature is in acceptable range and package in in valid status
if packageinfo.PkgStatus == "Pkg_Damaged" {
  jsonResp = "{\"Error\":\"Temprature thershold crossed - Package Damaged\"}"
  return nil, errors.New(jsonResp)
  }

temprature_reading, err = strconv.Atoi(args[1])
if err != nil {
	jsonResp = "{\"Error\":\"2nd argument must be a numeric string\"}"
  	return nil, errors.New(jsonResp)
	}


if temprature_reading > packageinfo.TempratureMax  || temprature_reading > packageinfo.TempratureMin  {
    packageinfo.PkgStatus = "Pkg_Damaged"
  }

bytes, err := json.Marshal(&packageinfo)
if err != nil {
        fmt.Println("Could not marshal personal info object", err)
        return nil, err
  }

err = stub.PutState(key, bytes)
if err != nil {
  return nil, err
  }

return nil, nil
}

//=================================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
fmt.Println("query is running " + function)

// Handle different functions
if function == "querypkgbyid" {
  return t.querypkgbyid(stub, args)
  } else if function == "queryallpkgids"{
  return t.queryallpkgids(stub, args)
  } else if function == "queryallpkg" {
  return t.queryallpkg(stub, args)
  } else if function == "querypkgbyshipper" {
  return t.querypkgbyshipper(stub, args)
  } else if function == "querypkgbyowner" {
  return t.querypkgbyowner(stub, args)
  } else if function == "querybypkgstatus"  {
  return t.querybypkgstatus(stub, args)
}

fmt.Println("query did not find func: " + function)
return nil, errors.New("Received unknown function query: " + function)
}


//=================================================================================================================================
//	querypkgbyid - query function to read key/value pair
//=================================================================================================================================
func (t *SimpleChaincode) querypkgbyid(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
var key, jsonResp string
var err error

key = args[0]

if len(args) != 1 {
  jsonResp = "{\"Error\":\"Incorrect number of arguments. Expecting PkgID to query " + key + "\"}"
  //return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
  return nil, errors.New(jsonResp)
  }

valAsbytes, err := stub.GetState(key)
if err != nil {
  jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
  return nil, errors.New(jsonResp)
}

if valAsbytes == nil {
  jsonResp = "{\"Error\":\"Invalid PackageId Passed" + key + "\"}"
  return nil, errors.New(jsonResp)
  }

var packageinfo PackageInfo
err = json.Unmarshal(valAsbytes, &packageinfo)
if err != nil {
      fmt.Println("Could not marshal personal info object", err)
      jsonResp = "{\"Error\":\"Could not marshal personal info object\"}"
      return nil, errors.New(jsonResp)
}

// validate pkg exist or not by checking temprature
if packageinfo.PkgId != key{
	  fmt.Println("Invalid PackageId Passed")
	  jsonResp = "{\"Error\":\"Invalid PackageId Passed" + key + "\"}"
          return nil, errors.New(jsonResp)
    }

return valAsbytes, nil
}

//=================================================================================================================================
//	queryallpkgids - query function to read all keys for packages
//=================================================================================================================================
func (t *SimpleChaincode) queryallpkgids(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){

var jsonResp string
var err error

if len(args) != 0 {
    jsonResp = "{\"Error\":\"Incorrect number of arguments.\"}"
    return nil, errors.New(jsonResp)
    }

valAsbytes, err := stub.GetState("PkgIdsKey")
if err != nil {
    jsonResp = "{\"Error\":\"Failed to get state for PkgIdsKey \"}"
    return nil, errors.New(jsonResp)
    }

if valAsbytes == nil {
    jsonResp = "{\"Error\":\"Invalid PackageId Passed fpr PkgIdsKey \"}"
    return nil, errors.New(jsonResp)
    }

return valAsbytes, nil

}

//=================================================================================================================================
//	queryallpkg - query function to read all key/value pair
//=================================================================================================================================
func (t *SimpleChaincode) queryallpkg(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){

  var jsonResp string
  var err error

  if len(args) != 0 {
      jsonResp = "{\"Error\":\"Incorrect number of arguments.\"}"
      return nil, errors.New(jsonResp)
      }

  valAsbytes, err := stub.GetState("PkgIdsKey")
  if err != nil {
      jsonResp = "{\"Error\":\"Failed to get state for PkgIdsKey \"}"
      return nil, errors.New(jsonResp)
      }

  var package_holder PKG_Holder
  err = json.Unmarshal(valAsbytes, &package_holder)
  if err != nil {
            fmt.Println("Could not marshal personal info object", err)
            jsonResp = "{\"Error\":\"Could not marshal personal info object\"}"
            return nil, errors.New(jsonResp)
      }

  var pkginfo PackageInfo

  result := "["

  var temp []byte

  for _, PkgId := range package_holder.PkgIds  {

    pkginfoasbytes, err := stub.GetState(PkgId)
    if err != nil {
      jsonResp = "{\"Error\":\"Failed to get state for " + PkgId + "\"}"
      return nil, errors.New(jsonResp)
    }

    err = json.Unmarshal(pkginfoasbytes, &pkginfo);
    if err != nil {
              fmt.Println("Could not marshal personal info object", err)
              jsonResp = "{\"Error\":\"Could not marshal personal info object\"}"
              return nil, errors.New(jsonResp)
    }

    temp = pkginfoasbytes
    result += string(temp) + ","

  }

  if len(result) == 1 {
    result = "[]"
  } else {
    result = result[:len(result)-1] + "]"
  }

  return []byte(result), nil

}
//=================================================================================================================================
//	querypkgbyshipper - query function to read key/value pair by given shipper
//=================================================================================================================================
func (t *SimpleChaincode) querypkgbyshipper(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){

return nil, nil
}
//=================================================================================================================================
//	querypkgbyowner - query function to read key/value pair by owner of package
//=================================================================================================================================
func (t *SimpleChaincode) querypkgbyowner(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){

return nil, nil
}
//=================================================================================================================================
//	querybypkgstatus - query function to read key/value pair by status of package
//=================================================================================================================================
func (t *SimpleChaincode) querybypkgstatus(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){

  var jsonResp string
  var err error

  if len(args) != 1 {
      jsonResp = "{\"Error\":\"Incorrect number of arguments. Need to pass status\"}"
      return nil, errors.New(jsonResp)
      }

  valAsbytes, err := stub.GetState("PkgIdsKey")
  if err != nil {
      jsonResp = "{\"Error\":\"Failed to get state for PkgIdsKey \"}"
      return nil, errors.New(jsonResp)
      }

  var package_holder PKG_Holder
  err = json.Unmarshal(valAsbytes, &package_holder)
  if err != nil {
            fmt.Println("Could not marshal personal info object", err)
            jsonResp = "{\"Error\":\"Could not marshal personal info object\"}"
            return nil, errors.New(jsonResp)
      }

  var pkginfo PackageInfo

  result := "["

  var temp []byte

  for _, PkgId := range package_holder.PkgIds  {

    pkginfoasbytes, err := stub.GetState(PkgId)
    if err != nil {
      jsonResp = "{\"Error\":\"Failed to get state for " + PkgId + "\"}"
      return nil, errors.New(jsonResp)
    }

    err = json.Unmarshal(pkginfoasbytes, &pkginfo);
    if err != nil {
              fmt.Println("Could not marshal personal info object", err)
              jsonResp = "{\"Error\":\"Could not marshal personal info object\"}"
              return nil, errors.New(jsonResp)
    }

// check for inout status
    if pkginfo.PkgStatus  == args[0] {
      temp = pkginfoasbytes
      result += string(temp) + ","
    }

  }

  if len(result) == 1 {
    result = "[]"
  } else {
    result = result[:len(result)-1] + "]"
  }

  return []byte(result), nil

}