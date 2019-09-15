package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// VoteAPIParams - Caps for field names, because of json.Marshal requirements
type VoteAPIParams struct {
	FacebookUserID      string `json:"fb_id"`
	FacebookAccessToken string `json:"fb_access_token"`
	PlaceID             string `json:"place_id"`
	PlaceAbbr           string `json:"place_abbr"`
}

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("ap-southeast-1"))

// HandleVotePlaceRequest - Lambda function
func HandleVotePlaceRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod == "POST" {
		params := VoteAPIParams{}
		err := json.Unmarshal([]byte(request.Body), &params)
		if err != nil {
			apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
			return apiResponse, err
		}

		responseBody := "{ \"success:\" false }"

		// Consider using memcache to store valid access token with user id and expiry
		ok := VerifyFacebookAccessToken(params.FacebookUserID, params.FacebookAccessToken)
		if ok {
			// Vote logic
			responseBody = "{ \"success:\" true }"
		}

		apiResponse := events.APIGatewayProxyResponse{
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Methods": "OPTIONS,POST",
			},
			Body:       string(responseBody),
			StatusCode: http.StatusOK}
		return apiResponse, nil
	} else {
		err := errors.New("Method not allowed")
		apiResponse := GenerateErrorResponse("Method Not OK", http.StatusBadGateway)
		return apiResponse, err
	}
}

func main() {
	lambda.Start(HandleVotePlaceRequest)
}
