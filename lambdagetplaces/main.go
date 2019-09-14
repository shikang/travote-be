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
)

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("ap-southeast-1"))

// GetPlacesWithoutAnyFilters - No filter get
func GetPlacesWithoutAnyFilters(abbr string, limit int64) ([]Place, error) {
	// Build the query input parameters
	params := &dynamodb.QueryInput{
		TableName: aws.String("Places"),
		KeyConditions: map[string]*dynamodb.Condition{
			"abbr": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(abbr),
					},
				},
			},
		},
		Limit: aws.Int64(limit),
	}

	// Make the DynamoDB Query API call
	result, err := db.Query(params)
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
	// Build the query input parameters
	params := &dynamodb.QueryInput{
		TableName: aws.String("Places"),
		IndexName: aws.String(filter + "-index"),
		KeyConditions: map[string]*dynamodb.Condition{
			filter: {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(val),
					},
				},
			},
		},
		Limit: aws.Int64(limit),
	}

	// Make the DynamoDB Query API call
	result, err := db.Query(params)
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
	case "abbr":
		return GetPlacesWithoutAnyFilters(val, limit)
	default:
		return GetPlacesWithFilter(filter, val, limit)
	}
}

// GetPlacesResponse - Get response
func GetPlacesResponse(filter string, val string, limit int64) (events.APIGatewayProxyResponse, error) {
	places, err := GetPlaces(filter, val, limit)
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

// HandleGetPlacesRequest - Lambda function
func HandleGetPlacesRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod == "GET" {
		var queryLimit int64 = 50
		if limit, ok := request.QueryStringParameters["limit"]; ok {
			queryLimit, _ = strconv.ParseInt(limit, 10, 64)
		}

		if abbr, ok := request.QueryStringParameters["abbr"]; ok {
			fmt.Print("[GET] Get places with abbr filter: " + abbr)
			return GetPlacesResponse("abbr", abbr, queryLimit)
		} else if category, ok := request.QueryStringParameters["category"]; ok {
			fmt.Print("[GET] Get places with category filter: " + category)
			return GetPlacesResponse("category", category, queryLimit)
		} else {
			err := errors.New("Please specify abbr or category")
			apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
			return apiResponse, err
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
