package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
)

type SmartContract struct {
	contractapi.Contract
}

type Person struct {
	Name       string `json:"name"`
	Age        int    `json:"age"`
	IdType     string `json:"idType"`
	IdNo       int    `json:"idNo"`
	Address    string `json:"address"`
	IsEmployed bool   `json:"isEmployed"`
	IsMarried  bool   `json:"isMarried"`
}

type QueryResult struct {
	Key    string `json:"Key"`
	Record *Person
}

// Create Person
func (s *SmartContract) CreatePerson(ctx contractapi.TransactionContextInterface,
	name string, age int, idType string, idNo int, address string) error {

	logger := logging.NewLogger("sample")
	logger.Infoln("Start: Calling CreatePerson function.")

	key := idNo

	person := Person{
		Name:       name,
		Age:        age,
		IdType:     idType,
		IdNo:       idNo,
		Address:    address,
		IsEmployed: false,
		IsMarried:  false,
	}

	personAsBytes, err := json.Marshal(person)
	if err != nil {
		return fmt.Errorf("CreatePerson: unable to Marshal %s ", personAsBytes)
	}

	return ctx.GetStub().PutState(strconv.Itoa(key), personAsBytes)

}

// Update Details
func (s *SmartContract) UpdateDetails(ctx contractapi.TransactionContextInterface,
	idNo int, age int, address string, isEmployed bool, isMarried bool) error {

	logger := logging.NewLogger("sample")
	logger.Infoln("Start: Calling UpdateDetails function.")

	queryResult, err := s.GetById(ctx, idNo)
	if err != nil {
		return err
	}

	queryResult.Age = age
	queryResult.Address = address
	queryResult.IsEmployed = isEmployed
	queryResult.IsMarried = isMarried

	personAsBytes, err := json.Marshal(queryResult)
	if err != nil {
		return fmt.Errorf("UpdateDetails: unable to Marshal %s ", personAsBytes)
	}

	return ctx.GetStub().PutState(strconv.Itoa(idNo), personAsBytes)

}

//Get by ID
func (s *SmartContract) GetById(ctx contractapi.TransactionContextInterface,
	idNo int) (*Person, error) {

	logger := logging.NewLogger("exercise")
	logger.Infoln("Start: Calling GetById function.")

	if strconv.Itoa(idNo) == "" {
		return nil, fmt.Errorf("GetById: input parameters must not be empty")
	}

	queryResult, err := ctx.GetStub().GetState(strconv.Itoa(idNo))
	if err != nil {
		return nil, fmt.Errorf("GetById: failed to read from world state: %v", err)
	}

	if queryResult == nil {
		return nil, fmt.Errorf("GetById: the person %s does not exist", strconv.Itoa(idNo))
	}

	var person Person
	err = json.Unmarshal(queryResult, &person)
	if err != nil {
		return nil, err
	}

	logger.Infoln("End: GetById called with key value of: ", idNo)
	return &person, nil
}

//Get Employed, return data of employed but NOT MARRIED
func (s *SmartContract) GetEmployed(ctx contractapi.TransactionContextInterface,
	isEmployed bool) ([]Person, error) {

	logger := logging.NewLogger("exercise")
	logger.Infoln("Start: Calling GetEmployed function.")

	// if strconv.FormatBool(isEmployed) == "" || strconv.FormatBool(isMarried) == "" {
	// 	return nil, fmt.Errorf("GetEmployed: input parameter must not be empty")
	// }

	queryResult := fmt.Sprintf("{\"selector\":{\"isEmployed\":%v}}", isEmployed)

	queryResultsIterator, err := ctx.GetStub().GetQueryResult(queryResult)
	if err != nil {
		return nil, err
	}
	defer queryResultsIterator.Close()

	var people []Person
	for queryResultsIterator.HasNext() {
		responseRange, err := queryResultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var person Person
		err = json.Unmarshal(responseRange.Value, &person)
		if err != nil {
			return nil, err
		}
		if !person.IsMarried {
			people = append(people, person)
		}

	}

	if people == nil {
		return nil, fmt.Errorf("GetEmployed: the person %s does not exist", strconv.FormatBool(isEmployed))
	}

	return people, nil

}

//Get People, return all the data in the ledger
func (s *SmartContract) GetPeople(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {

	logger := logging.NewLogger("exercise")
	logger.Infoln("Start: Calling GetPeople function.")

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var people []QueryResult
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var person Person
		err = json.Unmarshal(queryResponse.Value, &person)
		if err != nil {
			return nil, err
		}

		queryResult := QueryResult{Key: queryResponse.Key, Record: &person}
		people = append(people, queryResult)
	}

	if people == nil {
		return nil, fmt.Errorf("no records available")
	}

	logger.Infoln("End: GetPeople called.")
	return people, nil
}

// Delete All, this will delete all record on the ledger
func (s *SmartContract) DeleteAll(ctx contractapi.TransactionContextInterface,
	key int) error {

	logger := logging.NewLogger("transactionchaincode")
	logger.Infoln("Start: Calling DeleteFinanceRequest function.")

	data, err := s.GetById(ctx, key)
	if err != nil {
		return err
	}

	return ctx.GetStub().DelState(strconv.Itoa(data.IdNo))

}

func main() {
	cc, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create chaincode: %s", err.Error())
		return
	}

	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %s", err.Error())
	}
}
