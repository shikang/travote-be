package main

// Use this code snippet in your app.
// If you need more information about configurations or implementing the sample code, visit the AWS docs:
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// FacebookInfo - Caps for field names, because of json.Marshal requirements
type FacebookInfo struct {
	AppID     string `json:"travote_fb_app_id"`
	AppSecret string `json:"travote_fb_app_secret"`
}

// FacebookAppAccessTokenGraphAPIResponse - Caps for field names, because of json.Marshal requirements
type FacebookAppAccessTokenGraphAPIResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// FacebookCheckAccessTokenGraphAPIResponse - Caps for field names, because of json.Marshal requirements
type FacebookDebugAccessTokenGraphAPIResponse struct {
	Data struct {
		IsValid bool   `json:"is_valid"`
		UserID  string `json:"user_id"`
	} `json:"data"`
}

func GetFacebookInfoFromAWS() (FacebookInfo, error) {
	secretName := "TravoteFacebookAppInfo"
	//region := "ap-southeast-1"

	//Create a Secrets Manager client
	svc := secretsmanager.New(session.New())
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	// In this sample we only handle the specific exceptions for the 'GetSecretValue' API.
	// See https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html

	result, err := svc.GetSecretValue(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeDecryptionFailure:
				// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
				fmt.Println(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())

			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				fmt.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())

			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				fmt.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())

			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				fmt.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())

			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				fmt.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return FacebookInfo{}, err
	}

	// Decrypts secret using the associated KMS CMK.
	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	var secretString string
	if result.SecretString != nil {
		secretString = *result.SecretString
	} else {
		err := errors.New("Expected Secret to be in string")
		fmt.Println("Expected Secret to be in string")

		return FacebookInfo{}, err
	}

	fbInfo := FacebookInfo{}
	unmarshalErr := json.Unmarshal([]byte(secretString), &fbInfo)

	if unmarshalErr != nil {
		fmt.Println("Encountered error while unmarshalling info: " + unmarshalErr.Error())
	}

	return fbInfo, unmarshalErr
}

func VerifyFacebookAccessToken(userID string, accessToken string) bool {
	// Get App ID and App Secret from AWS secret manager
	fbInfo, err := GetFacebookInfoFromAWS()
	if err != nil {
		return false
	}

	// Get App Access Token
	url := "https://graph.facebook.com/oauth/access_token?client_id=" + fbInfo.AppID + "&client_secret=" + fbInfo.AppSecret + "&grant_type=client_credentials"
	//fmt.Println("Getting Facebook App Access Token - " + url)
	appAccessTokenResp, err := http.Get(url)
	if err != nil {
		fmt.Println("Facebook Get App Access Token Graph API fail! - " + err.Error())
		return false
	}

	appAccessTokenResponse := FacebookAppAccessTokenGraphAPIResponse{}
	defer appAccessTokenResp.Body.Close()
	err = json.NewDecoder(appAccessTokenResp.Body).Decode(&appAccessTokenResponse)
	if err != nil {
		fmt.Println("Error in decoding Facebook App Access Token Graph API Response - " + err.Error())
		return false
	}

	//fmt.Println("Facebook App Access Token = " + appAccessTokenResponse.AccessToken)

	// Check User Access Token
	url = "https://graph.facebook.com/debug_token?input_token=" + accessToken + "&access_token=" + appAccessTokenResponse.AccessToken
	//fmt.Println("Checking User Access Token - " + url)
	debugAccessTokenResp, err := http.Get(url)
	if err != nil {
		fmt.Println("Facebook Debug Access Token Graph API fail! - " + err.Error())
		return false
	}

	debugAccessTokenResponse := FacebookDebugAccessTokenGraphAPIResponse{}
	defer debugAccessTokenResp.Body.Close()
	err = json.NewDecoder(debugAccessTokenResp.Body).Decode(&debugAccessTokenResponse)
	if err != nil {
		fmt.Println("Error in decoding Facebook Debug Access Token Graph API Response - " + err.Error())
		return false
	}

	//fmt.Println("Facebook Debug Access Token Is Valid = " + strconv.FormatBool(debugAccessTokenResponse.Data.IsValid))
	//fmt.Println("Facebook Debug Access Token User ID = " + debugAccessTokenResponse.Data.UserID)
	return debugAccessTokenResponse.Data.IsValid && debugAccessTokenResponse.Data.UserID == userID
}
