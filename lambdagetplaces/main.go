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

type ErrorJson struct {
	ErrorMsg string `json:"error"`
}

// Place - Caps for field names, because of json.Marshal requirements
type Place struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Master   string `json:"master"`
	Category string `json:"category"`
	Desc     string `json:"desc"`
	Lat      string `json:"lat"`
	Long     string `json:"long"`
	Address  string `json:"address"`
	Postal   string `json:"postal"`
	Contact  string `json:"contact"`
	Hours    string `json:"hours"`
	Website  string `json:"website"`
	Email    string `json:"email"`
	Zone     string `json:"zone"`
	Ext1     string `json:"ext_1"`
}

// GetPlacesWithoutAnyFilters - No filter get
func GetPlacesWithoutAnyFilters(limit int64) ([]Place, error) {
	// Build the scan input parameters
	params := &dynamodb.ScanInput{
		TableName: aws.String("Places"),
		Limit:     aws.Int64(limit),
	}

	// Make the DynamoDB Query API call
	result, err := db.Scan(params)
	if err != nil {
		return nil, err
	}

	places := []Place{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &places)
	if err != nil {
		return nil, err
	}

	return places, nil
}

// GetPlacesWithFilter - Filter get
func GetPlacesWithFilter(filter string, val string, limit int64) ([]Place, error) {
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
		TableName:                 aws.String("Places"),
		Limit:                     aws.Int64(limit),
	}

	// Make the DynamoDB Query API call
	result, err := db.Scan(params)
	if err != nil {
		return nil, err
	}

	places := []Place{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &places)
	if err != nil {
		return nil, err
	}

	return places, nil
}

// GetPlaces - Get wrapper
func GetPlaces(filter string, val string, limit int64) ([]Place, error) {
	switch filter {
	case "any":
		return GetPlacesWithoutAnyFilters(limit)
	default:
		return GetPlacesWithFilter(filter, val, limit)
	}
}

// GetPlacesResponse - Get response
func GetPlacesResponse(filters string, val string, limit int64) (events.APIGatewayProxyResponse, error) {
	places, err := GetPlaces(filters, val, limit)
	if err != nil {
		apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
		return apiResponse, err
	}

	responseBody, err := json.Marshal(places)
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

// HandleGetPlacesRequest - Lambda function
func HandleGetPlacesRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod == "GET" {
		var queryLimit int64 = 10
		if limit, ok := request.QueryStringParameters["limit"]; ok {
			queryLimit, _ = strconv.ParseInt(limit, 10, 64)
		}

		if category, ok := request.QueryStringParameters["category"]; ok {
			fmt.Print("[GET] Get places with category filter: " + category)
			return GetPlacesResponse("category", category, queryLimit)
		} else {
			fmt.Print("[GET] Get places without filter")
			return GetPlacesResponse("any", "", queryLimit)
		}
	} else {
		err := errors.New("Method not allowed")
		apiResponse := GenerateErrorResponse("Method Not OK", http.StatusBadGateway)
		return apiResponse, err
	}
}

func main() {
	lambda.Start(HandleGetPlacesRequest)
}
