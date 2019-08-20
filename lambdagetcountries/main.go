package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("ap-southeast-1"))

// ErrorJson - Json for error reponse
type ErrorJson struct {
	ErrorMsg string `json:"error"`
}

// Country - Caps for field names, because of json.Marshal requirements
type Country struct {
	Abbr  string `json:"abbr"`
	Name  string `json:"name"`
	Xaxis int64  `json:"xaxis,string"`
	Yaxis int64  `json:"yaxis,string"`
}

// GetCountriesWithoutAnyFilters - No filter get
func GetCountriesWithoutAnyFilters(limit int64) ([]Country, error) {
	// Build the scan input parameters
	params := &dynamodb.ScanInput{
		TableName: aws.String("Countries"),
		Limit:     aws.Int64(limit),
	}

	// Make the DynamoDB Query API call
	result, err := db.Scan(params)
	if err != nil {
		return nil, err
	}

	countries := []Country{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &countries)
	if err != nil {
		return nil, err
	}

	return countries, nil
}

// GetCountriesWithFilter - Filter get
func GetCountriesWithFilter(filter string, val string, limit int64) ([]Country, error) {
	filt := expression.Name(filter).Equal(expression.Value(val))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, err
	}

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String("Countries"),
		Limit:                     aws.Int64(limit),
	}

	// Make the DynamoDB Query API call
	result, err := db.Scan(params)
	if err != nil {
		return nil, err
	}

	countries := []Country{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &countries)
	if err != nil {
		return nil, err
	}

	return countries, nil
}

// GetCountries - Get wrapper
func GetCountries(filter string, val string, limit int64) ([]Country, error) {
	switch filter {
	case "any":
		return GetCountriesWithoutAnyFilters(limit)
	default:
		return GetCountriesWithFilter(filter, val, limit)
	}
}

// GetCountriesResponse - Get response
func GetCountriesResponse(filters string, val string, limit int64) (events.APIGatewayProxyResponse, error) {
	countries, err := GetCountries(filters, val, limit)
	if err != nil {
		apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
		return apiResponse, err
	}

	responseBody, err := json.Marshal(countries)
	if err != nil {
		apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
		return apiResponse, err
	}

	apiResponse := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
		Body:       string(responseBody),
		StatusCode: http.StatusOK}
	return apiResponse, nil
}

// GenerateErrorResponse - Create error response
func GenerateErrorResponse(err string, statusCode int) events.APIGatewayProxyResponse {
	errJSON := &ErrorJson{ErrorMsg: err}
	errBody, _ := json.Marshal(errJSON)
	apiResponse := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
		Body:       string(errBody),
		StatusCode: statusCode}
	return apiResponse
}

// HandleGetCountriesRequest - Lambda function
func HandleGetCountriesRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod == "GET" {
		var queryLimit int64 = 10
		if limit, ok := request.QueryStringParameters["limit"]; ok {
			queryLimit, _ = strconv.ParseInt(limit, 10, 64)
		}

		if abbr, ok := request.QueryStringParameters["abbr"]; ok {
			fmt.Print("[GET] Get countries with abbr filter: " + abbr)
			return GetCountriesResponse("abbr", abbr, queryLimit)
		} else if name, ok := request.QueryStringParameters["name"]; ok {
			fmt.Print("[GET] Get countries with name filter: " + name)
			return GetCountriesResponse("name", name, queryLimit)
		} else {
			fmt.Print("[GET] Get countries without filter")
			return GetCountriesResponse("any", "", queryLimit)
		}
	} else {
		err := errors.New("Method not allowed")
		apiResponse := GenerateErrorResponse("Method Not OK", http.StatusBadGateway)
		return apiResponse, err
	}
}

func main() {
	lambda.Start(HandleGetCountriesRequest)
}
