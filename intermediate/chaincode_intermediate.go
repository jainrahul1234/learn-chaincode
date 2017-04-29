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

// Import all the necessary libraries
import (
"errors"
"fmt"
"encoding/json"
"github.com/hyperledger/fabric/core/chaincode/shim"
"strconv"
//"github.com/hyperledger/fabric/protos/peer"
)

//custom data models for Package Information
type PackageInfo struct {
  PkgId      string `json:"packageid"`
  Shipper    string `json:"shipper"`
  Insurer    string `json:"insurer"`
  Consignee  string `json:"consignee"`
  Owner      string `json:"owner"`
  Temprature int `json:"temprature"`
  PackageDes string `json:"packagedes"`
  PkgStatus  string `json:"pkgstatus"`
}

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func main() {
err := shim.Start(new(SimpleChaincode))
if err != nil {
  fmt.Printf("Error starting Simple chaincode: %s", err)
  }
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
if len(args) != 6 {
  return nil, errors.New("Incorrect number of arguments. Expecting 6 in order of Shipper, Insurer, Consignee, Temprature, PackageDes, Owner ")
  }

var packageinfo PackageInfo
var err error

packageinfo.PkgId = "1Z20170426"
packageinfo.Shipper = args[0]
packageinfo.Insurer  = args[1]
packageinfo.Consignee  = args[2]
packageinfo.Temprature, err = strconv.Atoi(args[3])
if err != nil {
    return nil, errors.New("2nd argument must be a numeric string")
	}
packageinfo.PackageDes = args[4]
packageinfo.Owner = args[5]
packageinfo.PkgStatus = "In-Valid"
if packageinfo.Temprature < 5 && packageinfo.Temprature > -5 {
    packageinfo.PkgStatus = "Valid"
  }

bytes, err := json.Marshal(&packageinfo)
if err != nil {
        fmt.Println("Could not marshal personal info object", err)
        return nil, err
  }

err = stub.PutState("1Z20170426", bytes)
if err != nil {
  return nil, err
  }

return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
fmt.Println("invoke is running " + function)

// Handle different functions
if function == "init" {
  return t.Init(stub, "init", args)
} else if function == "updatetemp" {
  return t.updatetemp(stub, args)
} else if function == "create" {
  return t.create(stub,args)
} else if function == "deliverpkg"{
  return t.deliverpkg(stub,args)
}

fmt.Println("invoke did not find func: " + function)

return nil, errors.New("Received unknown function invocation: " + function)
}

// create - invoke function to create new asset using given key/value pair
func (t *SimpleChaincode) create(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
fmt.Println("running create()")

var key string
var err error

if len(args) != 6 {
  return nil, errors.New("Incorrect number of arguments. Expecting 7 in order of PkgID, Shipper, Insurer, Consignee, Temprature, PackageDes, Owner")
  }

var packageinfo PackageInfo

key = args[0]

packageinfo.PkgId  = args[0]
packageinfo.Shipper = args[1]
packageinfo.Insurer = args[2]
packageinfo.Consignee  = args[3]
packageinfo.Temprature, err = strconv.Atoi(args[4])
if err != nil {
    return nil, errors.New("5th argument must be a numeric string")
	}
packageinfo.PackageDes = args[5]
packageinfo.Owner = args[6]

packageinfo.PkgStatus = "In-Valid"
if packageinfo.Temprature < 5 && packageinfo.Temprature > -5 {
    packageinfo.PkgStatus = "Valid"
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


// update temprature - invoke function to update the temprature of Package
func (t *SimpleChaincode) updatetemp(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
var key , jsonResp string
var err error
fmt.Println("running updatetemp()")

if len(args) != 2 {
  return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and temprature value to set")
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
      fmt.Println("Could not marshal  info object", err)
      return nil, err
  }
// validate pkd exist or not by checking temprature
if packageinfo.PkgId != key{
  return nil, errors.New("Invalid PackageId Passed")
  }

// check wheather the pkg temprature is in acceptable range and package in in valid status
if packageinfo.PkgStatus != "Valid" {
    return nil, errors.New("Temprature thershold crossed - Package in Invalid state")
  }

packageinfo.Temprature, err = strconv.Atoi(args[1])
if err != nil {
    return nil, errors.New("2nd argument must be a numeric string")
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


// deliverpkg - update owner og package
func (t *SimpleChaincode) deliverpkg(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
fmt.Println("running deliverpkg()")
var key , jsonResp string
var err error

if len(args) != 2 {
  return nil, errors.New("Incorrect number of arguments. Expecting 2. PkgId and New Owner")
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
    return nil, errors.New("Invalid PackageId Passed")
    }

  // check wheather the pkg temprature is in acceptable range and package in in valid status
  if packageinfo.PkgStatus != "Valid" {
      return nil, errors.New("Temprature thershold crossed - Package in Invalid state")
    }

  packageinfo.Owner = args[1]

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



// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
fmt.Println("query is running " + function)

// Handle different functions
if function == "read" { //read a variable
  return t.read(stub, args)
} 
fmt.Println("query did not find func: " + function)

return nil, errors.New("Received unknown function query: " + function)
}



// read - query function to read key/value pair
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
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

var packageinfo PackageInfo
err = json.Unmarshal(valAsbytes, &packageinfo)
if err != nil {
      fmt.Println("Could not marshal personal info object", err)
      return nil, err
}

// validate pkd exist or not by checking temprature
  if packageinfo.PkgId != key{
    return nil, errors.New("Invalid PackageId Passed")
    }

fmt.Println(packageinfo.PkgId)
fmt.Println(packageinfo.Shipper)
fmt.Println(packageinfo.Insurer)
fmt.Println(packageinfo.Consignee)
fmt.Println(packageinfo.Temprature)
fmt.Println(packageinfo.PackageDes)
fmt.Println(packageinfo.Owner)
fmt.Println(packageinfo.PkgStatus)

return valAsbytes, nil
//  return packageinfo, nil
}

