/*
Copyright IBM Corp 2016 All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
		 http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/


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
  Provider      string `json:"provider"`
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
if len(args) != 7 {
  jsonResp = "Error: Incorrect number of arguments. Expecting 6 in order of Shipper, Insurer, Consignee, Temprature, PackageDes, Provider "
  return nil, errors.New(jsonResp)
  }


//  Polulating JSON block with input for first block
packageinfo.PkgId = "1Z20170426"
packageinfo.Shipper = args[0]
packageinfo.Insurer  = args[1]
packageinfo.Consignee  = args[2]
packageinfo.Provider = args[3]
packageinfo.TempratureMin, err = strconv.Atoi(args[4])
if err != nil {
  jsonResp = "Error :5th argument must be a numeric string"
  return nil, errors.New(jsonResp)
	}
packageinfo.TempratureMax, err = strconv.Atoi(args[5])
if err != nil {
    jsonResp = "Error: 6th argument must be a numeric string "
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
  } else if function == "acceptpkg"{
  return t.acceptpkg(stub,args)
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
  jsonResp = " Error: Incorrect number of arguments. Expecting 8 in order of PkgID, Shipper, Insurer, Consignee, TempratureMin, TempratureMax, PackageDes, Provider "
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
  jsonResp = " Error: 5th argument must be a numeric string "
  return nil, errors.New(jsonResp)
	}
packageinfo.TempratureMax  , err = strconv.Atoi(args[5])
if err != nil {
  jsonResp = " Error: 5th argument must be a numeric string "
  return nil, errors.New(jsonResp)
	}
packageinfo.PackageDes = args[6]
packageinfo.Provider = args[7]
packageinfo.PkgStatus = "Label_Generated"   // Label_Generated

bytes, err := json.Marshal(&packageinfo)
if err != nil {
        fmt.Println("Could not marshal personal info object", err)
        return nil, err
  }

// check for duplicate package id
valAsbytes, err := stub.GetState(key)

if valAsbytes != nil {
  jsonResp = " Package already present on blockchain " + key
  return nil, errors.New(jsonResp)
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
//	acceptpkg - Accept Package from Shipper , change status
//=================================================================================================================================
func (t *SimpleChaincode) acceptpkg(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
fmt.Println("running acceptpkg()")
var key , jsonResp string
var err error

if len(args) != 2 {
	jsonResp = " Error:Incorrect number of arguments. Expecting : PkgId and Provider "
  	return nil, errors.New(jsonResp)
  }

  key = args[0]
  var packageinfo PackageInfo

  valAsbytes, err := stub.GetState(key)

  if err != nil {
    jsonResp = " Error Failed to get state for " + key
    return nil, errors.New(jsonResp)
    }

  err = json.Unmarshal(valAsbytes, &packageinfo)
  if err != nil {
        fmt.Println("Could not marshal personal info object", err)
        return nil, err
    }

// validate pkd exist or not by checking temprature
  if packageinfo.PkgId != key{
    jsonResp = "Error: Invalid PackageId Passed "
    return nil, errors.New(jsonResp)
    }

  // check wheather the pkg temprature is in acceptable range and package in in valid status
  if packageinfo.PkgStatus == "Pkg_Damaged" {    // Pkg_Damaged
	  jsonResp = "Error : Temprature thershold crossed - Package Damaged "
          return nil, errors.New(jsonResp)
    }

	if packageinfo.Provider != args[1] {    // Pkg_Damaged
		  jsonResp = "Error : Wrong Provider passed - Can not accept the package "
	          return nil, errors.New(jsonResp)
	    }

  //packageinfo.Provider = args[1]
  packageinfo.PkgStatus = "In_Transit"

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
//	deliverpkg - deliver package to cosignee, change status of the package
//=================================================================================================================================
func (t *SimpleChaincode) deliverpkg(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
fmt.Println("running deliverpkg()")
var key , jsonResp string
var err error

if len(args) != 2 {
	jsonResp = " Error : Incorrect number of arguments. Expecting 2 : PkgId and Provider "
  	return nil, errors.New(jsonResp)
  }

  key = args[0]
  var packageinfo PackageInfo

  valAsbytes, err := stub.GetState(key)

  if err != nil {
    jsonResp = "Error : Failed to get state for " + key
    return nil, errors.New(jsonResp)
    }

  err = json.Unmarshal(valAsbytes, &packageinfo)
  if err != nil {
        fmt.Println("Could not marshal personal info object", err)
        return nil, err
    }

// validate pkd exist or not by checking temprature
  if packageinfo.PkgId != key{
    jsonResp = "Error: Invalid PackageId Passed "
    return nil, errors.New(jsonResp)
    }

  // check wheather the pkg temprature is in acceptable range and package in in valid status
  if packageinfo.PkgStatus == "Pkg_Damaged" {    // Pkg_Damaged
	  jsonResp = " Error: Temprature thershold crossed - Package Damaged"
          return nil, errors.New(jsonResp)
    }

	if packageinfo.PkgStatus == "Pkg_Delivered" {    // Pkg_Damaged
	  jsonResp = " Error: Package Already Delivered"
	  return nil, errors.New(jsonResp)
	  }

 // check wheather the pkg Provider is same as input value
if packageinfo.Provider != args[1] {
	  jsonResp = " Error :Wrong Pkg Provider passrd - Not authorized to deliver this Package"
	  return nil, errors.New(jsonResp)
	  }

//  packageinfo.Owner = args[1]
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
  jsonResp = "Error :Incorrect number of arguments. Expecting 2. name of the key and temprature value to set"
  return nil, errors.New(jsonResp)
  }


key = args[0]
var packageinfo PackageInfo
var temprature_reading int

valAsbytes, err := stub.GetState(key)

if err != nil {
  jsonResp = "Error :Failed to get state for " + key
  return nil, errors.New(jsonResp)
  }

err = json.Unmarshal(valAsbytes, &packageinfo)
if err != nil {
      fmt.Println("Could not marshal  info object", err)
      return nil, err
  }
// validate pkd exist or not by checking temprature
if packageinfo.PkgId != key{
  jsonResp = " Error : Invalid PackageId Passed "
  return nil, errors.New(jsonResp)
  }

// check wheather the pkg temprature is in acceptable range and package in in valid status
if packageinfo.PkgStatus == "Pkg_Damaged" {
  jsonResp = " Error :Temprature thershold crossed - Package Damaged"
  return nil, errors.New(jsonResp)
  }

temprature_reading, err = strconv.Atoi(args[1])
if err != nil {
	jsonResp = " Error : 2nd argument must be a numeric string"
  	return nil, errors.New(jsonResp)
	}


if temprature_reading > packageinfo.TempratureMax  || temprature_reading < packageinfo.TempratureMin  {
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
  } else if function == "querypkgbyprovider" {
  return t.querypkgbyprovider(stub, args)
  } else if function == "querypkgbyshipper" {
  return t.querypkgbyshipper(stub, args)
  } else if function == "querybypkgstatus"  {
  return t.querybypkgstatus(stub, args)
  } else if function == "querybyrole"{
  return t.querybyrole(stub, args)
  } else if function == "querybyrole_status"{
  return t.querybyrole_status(stub, args)
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
  jsonResp = " Error: Incorrect number of arguments. Expecting PkgID to query  " + key
	//return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
  return nil, errors.New(jsonResp)
  }

valAsbytes, err := stub.GetState(key)
if err != nil {
  jsonResp = "Error :Failed to get state for " + key
  return nil, errors.New(jsonResp)
}

if valAsbytes == nil {
  jsonResp = " Error: Invalid PackageId Passed " + key
  return nil, errors.New(jsonResp)
  }

var packageinfo PackageInfo
err = json.Unmarshal(valAsbytes, &packageinfo)
if err != nil {
      fmt.Println("Could not marshal personal info object", err)
      jsonResp = " Error :Could not marshal personal info object"
      return nil, errors.New(jsonResp)
}

// validate pkg exist or not by checking temprature
if packageinfo.PkgId != key{
	  fmt.Println("Invalid PackageId Passed")
	  jsonResp = " Error :Invalid PackageId Passed " + key
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
    jsonResp = " Error: Incorrect number of arguments "
    return nil, errors.New(jsonResp)
    }

valAsbytes, err := stub.GetState("PkgIdsKey")
if err != nil {
    jsonResp = "Error: Failed to get state for PkgIdsKey "
    return nil, errors.New(jsonResp)
    }

if valAsbytes == nil {
    jsonResp = "Error: Invalid PackageId Passed fpr PkgIdsKey "
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
      jsonResp = "Error: Incorrect number of arguments."
      return nil, errors.New(jsonResp)
      }

  valAsbytes, err := stub.GetState("PkgIdsKey")
  if err != nil {
      jsonResp = "Error:Failed to get state for PkgIdsKey "
      return nil, errors.New(jsonResp)
      }

  var package_holder PKG_Holder
  err = json.Unmarshal(valAsbytes, &package_holder)
  if err != nil {
            fmt.Println("Could not marshal personal info object", err)
            jsonResp = "Error: Could not marshal personal info object"
            return nil, errors.New(jsonResp)
      }

  var pkginfo PackageInfo

  result := "["

  var temp []byte

  for _, PkgId := range package_holder.PkgIds  {

    pkginfoasbytes, err := stub.GetState(PkgId)
    if err != nil {
      jsonResp = "Error:Failed to get state for " + PkgId
      return nil, errors.New(jsonResp)
    }

    err = json.Unmarshal(pkginfoasbytes, &pkginfo);
    if err != nil {
              fmt.Println("Could not marshal personal info object", err)
              jsonResp = "Error: Could not marshal personal info object"
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
//	querypkgbyprovider- query function to read key/value pair by given Provider
//=================================================================================================================================
func (t *SimpleChaincode) querypkgbyprovider(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){


    var jsonResp string
    var err error

    if len(args) != 1 {
        jsonResp = "Error:Incorrect number of arguments. Need to pass Provider"
        return nil, errors.New(jsonResp)
        }

    valAsbytes, err := stub.GetState("PkgIdsKey")
    if err != nil {
        jsonResp = "Error:Failed to get state for PkgIdsKey "
        return nil, errors.New(jsonResp)
        }

    var package_holder PKG_Holder
    err = json.Unmarshal(valAsbytes, &package_holder)
    if err != nil {
              fmt.Println("Could not marshal personal info object", err)
              jsonResp = "Error:Could not marshal personal info object"
              return nil, errors.New(jsonResp)
        }

    var pkginfo PackageInfo

    result := "["

    var temp []byte

    for _, PkgId := range package_holder.PkgIds  {

      pkginfoasbytes, err := stub.GetState(PkgId)
      if err != nil {
        jsonResp = " Error:Failed to get state for " + PkgId
        return nil, errors.New(jsonResp)
      }

      err = json.Unmarshal(pkginfoasbytes, &pkginfo);
      if err != nil {
                fmt.Println("Could not marshal personal info object", err)
                jsonResp = "Error: Could not marshal personal info object "
                return nil, errors.New(jsonResp)
      }

  // check for inout owner
      if pkginfo.Provider == args[0] {
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

//=================================================================================================================================
//	querypkgbyshipper - query function to read key/value pair by shipper of package
//=================================================================================================================================
func (t *SimpleChaincode) querypkgbyshipper(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){

  var jsonResp string
  var err error

  if len(args) != 1 {
      jsonResp = "Error: Incorrect number of arguments. Need to pass Shipper "
      return nil, errors.New(jsonResp)
      }

  valAsbytes, err := stub.GetState("PkgIdsKey")
  if err != nil {
      jsonResp = "Error: Failed to get state for PkgIdsKey "
      return nil, errors.New(jsonResp)
      }

  var package_holder PKG_Holder
  err = json.Unmarshal(valAsbytes, &package_holder)
  if err != nil {
            fmt.Println("Could not marshal personal info object", err)
            jsonResp = " Error:Could not marshal personal info object"
            return nil, errors.New(jsonResp)
      }

  var pkginfo PackageInfo

  result := "["

  var temp []byte

  for _, PkgId := range package_holder.PkgIds  {

    pkginfoasbytes, err := stub.GetState(PkgId)
    if err != nil {
      jsonResp = "Error:Failed to get state for " + PkgId
      return nil, errors.New(jsonResp)
    }

    err = json.Unmarshal(pkginfoasbytes, &pkginfo);
    if err != nil {
              fmt.Println("Could not marshal personal info object", err)
              jsonResp = " Error:Could not marshal personal info object"
              return nil, errors.New(jsonResp)
    }

// check for inout Shipper
    if pkginfo.Shipper == args[0] {
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

//=================================================================================================================================
//	querybypkgstatus - query function to read key/value pair by status of package
//=================================================================================================================================
func (t *SimpleChaincode) querybypkgstatus(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){

  var jsonResp string
  var err error

  if len(args) != 1 {
      jsonResp = "Error: Incorrect number of arguments. Need to pass status"
      return nil, errors.New(jsonResp)
      }

  valAsbytes, err := stub.GetState("PkgIdsKey")
  if err != nil {
      jsonResp = " Error:Failed to get state for PkgIdsKey "
      return nil, errors.New(jsonResp)
      }

  var package_holder PKG_Holder
  err = json.Unmarshal(valAsbytes, &package_holder)
  if err != nil {
            fmt.Println("Could not marshal personal info object", err)
            jsonResp = " Error: Could not marshal personal info object"
            return nil, errors.New(jsonResp)
      }

  var pkginfo PackageInfo

  result := "["

  var temp []byte

  for _, PkgId := range package_holder.PkgIds  {

    pkginfoasbytes, err := stub.GetState(PkgId)
    if err != nil {
      jsonResp = "Error:Failed to get state for " + PkgId
      return nil, errors.New(jsonResp)
    }

    err = json.Unmarshal(pkginfoasbytes, &pkginfo);
    if err != nil {
              fmt.Println("Could not marshal personal info object", err)
              jsonResp = "Error: Could not marshal personal info object "
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

//=================================================================================================================================
//	querybyrole - query function to read key/value pair by role
//=================================================================================================================================
func (t *SimpleChaincode) querybyrole(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){

  var jsonResp string
  var err error

  if len(args) != 2 {
      jsonResp = "Error:Incorrect number of arguments. Need to pass Role: Shipper, Provider, Insurer or Consignee & status value to be passed"
      return nil, errors.New(jsonResp)
      }

// validate role
  if args[0] == "Shipper"{
  fmt.Println("Shipper has been passed as Role")
  } else if args[0] == "Provider" {
  fmt.Println("Provider has been passed as Role")
  } else if args[0] == "Insurer" {
  fmt.Println("Insurer has been passed as Role")
  } else if args[0] == "Consignee" {
  fmt.Println("Consignee has been passed as Role")
  } else {
    jsonResp = " Error:Incorrect Role has been passed, should be: Shipper, Provider, Insurer or Consignee"
    return nil, errors.New(jsonResp)
  }

  valAsbytes, err := stub.GetState("PkgIdsKey")
  if err != nil {
      jsonResp = "Error:Failed to get state for PkgIdsKey "
      return nil, errors.New(jsonResp)
      }

  var package_holder PKG_Holder
  err = json.Unmarshal(valAsbytes, &package_holder)
  if err != nil {
            fmt.Println("Could not marshal personal info object", err)
            jsonResp = "Error:Could not marshal personal info object"
            return nil, errors.New(jsonResp)
      }

  var pkginfo PackageInfo

  result := "["

  var temp []byte

  for _, PkgId := range package_holder.PkgIds  {

    pkginfoasbytes, err := stub.GetState(PkgId)
    if err != nil {
      jsonResp = "Error:Failed to get state for " + PkgId
      return nil, errors.New(jsonResp)
    }

    err = json.Unmarshal(pkginfoasbytes, &pkginfo);
    if err != nil {
              fmt.Println("Could not marshal personal info object", err)
              jsonResp = "Error:Could not marshal personal info object"
              return nil, errors.New(jsonResp)
    }

    // check for inout role
    if args[0] == "Provider"{
      if pkginfo.Provider == args[1] {
        temp = pkginfoasbytes
        result += string(temp) + ","
      }
    } else if args[0] == "Shipper" {
      if pkginfo.Shipper == args[1] {
        temp = pkginfoasbytes
        result += string(temp) + ","
      }
    } else if args[0] == "Insurer" {
      if pkginfo.Insurer == args[1] {
        temp = pkginfoasbytes
        result += string(temp) + ","
      }
    } else if args[0] == "Consignee" {
      if pkginfo.Consignee == args[1] {
        temp = pkginfoasbytes
        result += string(temp) + ","
      }
    }


  } // end of for loop

  if len(result) == 1 {
    result = "[]"
  } else {
    result = result[:len(result)-1] + "]"
  }

  return []byte(result), nil

}


//=================================================================================================================================
//	querybyrole_status - query function to read key/value pair by role & status of package
//=================================================================================================================================
func (t *SimpleChaincode) querybyrole_status(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){

  var jsonResp string
  var err error

  if len(args) != 3 {
      jsonResp = "Error: Incorrect number of arguments. Need to pass Role, Value and Status "
      return nil, errors.New(jsonResp)
      }

// validate role
  if args[0] == "Shipper"{
  fmt.Println("Shipper has been passed as Role")
  } else if args[0] == "Provider" {
  fmt.Println("Provider has been passed as Role")
  } else if args[0] == "Insurer" {
  fmt.Println("Insurer has been passed as Role")
  } else if args[0] == "Consignee" {
  fmt.Println("Consignee has been passed as Role")
  } else {
    jsonResp = "Error:Incorrect Role has been passed, should be: Shipper, Provider, Insurer or Consignee"
    return nil, errors.New(jsonResp)
  }

// validate status
    if args[2] == "Label_Generated"{
    fmt.Println("Label_Generated has been passed as status")
    } else if args[2] == "In_Transit" {
    fmt.Println("In_Transit has been passed as status")
    } else if args[2] == "Pkg_Damaged" {
    fmt.Println("Pkg_Damaged has been passed as status")
    } else if args[2] == "Pkg_Delivered" {
    fmt.Println("Pkg_Delivered has been passed as status")
    } else {
      jsonResp = "Error: Incorrect Status has been passed, should be: Label_Generated, In_Transit, Pkg_Damaged or Pkg_Delivered"
      return nil, errors.New(jsonResp)
    }

  valAsbytes, err := stub.GetState("PkgIdsKey")
  if err != nil {
      jsonResp = "Error:Failed to get state for PkgIdsKey "
      return nil, errors.New(jsonResp)
      }

  var package_holder PKG_Holder
  err = json.Unmarshal(valAsbytes, &package_holder)
  if err != nil {
            fmt.Println("Could not marshal personal info object", err)
            jsonResp = "Error:Could not marshal personal info object"
            return nil, errors.New(jsonResp)
      }

  var pkginfo PackageInfo

  result := "["

  var temp []byte

  for _, PkgId := range package_holder.PkgIds  {

    pkginfoasbytes, err := stub.GetState(PkgId)
    if err != nil {
      jsonResp = "Error: Failed to get state for " + PkgId
      return nil, errors.New(jsonResp)
    }

    err = json.Unmarshal(pkginfoasbytes, &pkginfo);
    if err != nil {
              fmt.Println("Could not marshal personal info object", err)
              jsonResp = "Error:Could not marshal personal info object"
              return nil, errors.New(jsonResp)
    }

    // check for inout role & Status - this is crude way to do this - need to find another way
    if args[0] == "Provider"{
      if pkginfo.Provider == args[1] {
        if pkginfo.PkgStatus  == args[2] {
        temp = pkginfoasbytes
        result += string(temp) + ","
        }
      }
    } else if args[0] == "Shipper" {
      if pkginfo.Shipper == args[1] {
        if pkginfo.PkgStatus == args[2] {
        temp = pkginfoasbytes
        result += string(temp) + ","
        }
      }
    } else if args[0] == "Insurer" {
      if pkginfo.Insurer == args[1] {
        if pkginfo.PkgStatus == args[2] {
        temp = pkginfoasbytes
        result += string(temp) + ","
        }
      }
    } else if args[0] == "Consignee" {
      if pkginfo.Consignee == args[1] {
        if pkginfo.PkgStatus == args[2] {
        temp = pkginfoasbytes
        result += string(temp) + ","
        }
      }
    }


  } // end of for loop

  if len(result) == 1 {
    result = "[]"
  } else {
    result = result[:len(result)-1] + "]"
  }

  return []byte(result), nil

}
