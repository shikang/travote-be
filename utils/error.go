package main

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

// ErrorJson - Json for error reponse
type ErrorJson struct {
	ErrorMsg string `json:"error"`
}

// GenerateErrorResponse - Create error response
func GenerateErrorResponse(err string, statusCode int) events.APIGatewayProxyResponse {
	errJSON := &ErrorJson{ErrorMsg: err}
	errBody, _ := json.Marshal(errJSON)
	apiResponse := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Methods": "OPTIONS,GET,POST,PUT",
		},
		Body:       string(errBody),
		StatusCode: statusCode}
	return apiResponse
}
