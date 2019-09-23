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

// GetPlacesWithoutAnyFilters - No filter get
func GetPlacesByLongLat(filter string, val string, long float64, lat float64, distance float64, limit int64) ([]Place, error) {
	// Lat range -90 to 90
	minLat := lat - distance
	maxLat := lat + distance

	// Long range -180 to 180
	minLong := long - distance
	maxLong := long + distance

	filt := expression.Name(filter).Equal(expression.Value(val))

	latFilt := expression.Name("lat").Between(expression.Value(minLat), expression.Value(maxLat))
	if minLat < -90 {
		latFilt = expression.Name("lat").Between(expression.Value(-90), expression.Value(maxLat))
	} else if maxLat > 90 {
		latFilt = expression.Name("lat").Between(expression.Value(minLat), expression.Value(90))
	}

	longFilt := expression.Name("long").Between(expression.Value(minLong), expression.Value(maxLong))
	if minLong < -180 {
		newMinLong := 360 + minLong
		longFiltEx := expression.Name("long").Between(expression.Value(newMinLong), expression.Value(180))
		longFilt = expression.Name("long").Between(expression.Value(-180), expression.Value(maxLong)).And(longFiltEx)
	} else if maxLong > 180 {
		newMaxLong := maxLong - 360
		longFiltEx := expression.Name("long").Between(expression.Value(-180), expression.Value(newMaxLong))
		longFilt = expression.Name("long").Between(expression.Value(minLong), expression.Value(180)).And(longFiltEx)
	}

	expr, err := expression.NewBuilder().WithFilter(filt.And(latFilt.And(longFilt))).Build()
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

// GetPlacesWithFilter - Filter query by GSI
/*
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
*/

// GetPlaces - Get wrapper
func GetPlaces(filter string, val string, limit int64) ([]Place, error) {
	switch filter {
	case "abbr":
		return GetPlacesWithoutAnyFilters(val, limit)
	default:
		//return GetPlacesWithFilter(filter, val, limit)
		fmt.Print("Unsupported filter: " + filter)
		err := errors.New("Unsupported filter")
		return nil, err
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

// GetPlacesByLongLatResponse - Get response
func GetPlacesByLongLatResponse(filter string, val string, long float64, lat float64, distance float64, limit int64) (events.APIGatewayProxyResponse, error) {
	places, err := GetPlacesByLongLat(filter, val, long, lat, distance, limit)
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
			long, longOk := request.QueryStringParameters["long"]
			lat, latOk := request.QueryStringParameters["lat"]
			distance, distanceOk := request.QueryStringParameters["distance"]

			if longOk && latOk && distanceOk {
				longF, err := strconv.ParseFloat(long, 64)
				if err != nil {
					fmt.Println("Error parsing float for long: " + err.Error())
					apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
					return apiResponse, err
				}

				latF, err := strconv.ParseFloat(lat, 64)
				if err != nil {
					fmt.Println("Error parsing float for lat: " + err.Error())
					apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
					return apiResponse, err
				}

				distanceF, err := strconv.ParseFloat(distance, 64)
				if err != nil {
					fmt.Println("Error parsing float for distance: " + err.Error())
					apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
					return apiResponse, err
				}

				fmt.Print("[GET] Get places with abbr filter: " + abbr + " | long: " + long + " | lat: " + lat + " | distance: " + distance)
				return GetPlacesByLongLatResponse("abbr", abbr, longF, latF, distanceF, queryLimit)
			} else {
				fmt.Print("[GET] Get places with abbr filter only: " + abbr)
				return GetPlacesResponse("abbr", abbr, queryLimit)
			}
		} else {
			err := errors.New("Please specify abbr")
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
