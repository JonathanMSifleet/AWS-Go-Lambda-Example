package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"golang.org/x/crypto/bcrypt"
)

var cfg, err = config.LoadDefaultConfig(context.Background())
var dbClient = dynamodb.NewFromConfig(cfg)

type SignupEvent struct {
	Body string `json:"body"`
}

type SignupBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type HTTPResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

func main() {
	lambda.Start(Signup)
}

func Signup(event SignupEvent) (HTTPResponse, error) {
	var body SignupBody
	err := json.Unmarshal([]byte(event.Body), &body)
	if err != nil {
		return HTTPResponse{}, err
	}

	password := body.Password
	username := strings.TrimSpace(body.Username)
	email := strings.TrimSpace(body.Email)

	errors, err := ValidateUserInputs(username, email)
	if err != nil {
		return HTTPResponse{}, err
	}
	if len(errors) != 0 {
		return CreateAWSResErr(400, errors), nil
	}

	memberSince := GetSignupDate()

	salt, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return HTTPResponse{}, err
	}

	err = InsertUserToDB(username, email, string(salt), memberSince)
	if err != nil {
		return CreateAWSResErr(500, []string{err.Error()}), nil
	}

	log.Println("Signed up successfully")

	return HTTPResponse{
		StatusCode: 201,
		Body:       "Signup successful",
	}, nil
}

func GetSignupDate() int64 {
	return time.Now().Unix() / (24 * 60 * 60) * 24 * 60 * 60
}

func InsertUserToDB(username string, email string, password string, memberSince int64) error {
	item := Item{
		email:       email,
		memberSince: memberSince,
		numRatings:  0,
		password:    password,
		username:    username,
	}

	input := &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{
				Value: email,
			},
			"memberSince": &types.AttributeValueMemberN{
				Value: fmt.Sprint(memberSince),
			},
			"numRatings": &types.AttributeValueMemberN{
				Value: "0",
			},
			"password": &types.AttributeValueMemberS{
				Value: password,
			},
			"username": &types.AttributeValueMemberS{
				Value: username,
			},
		},
		TableName:              aws.String(os.Getenv("USER_TABLE_NAME")),
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}

	_, err := dbClient.PutItem(context.Background(), input)
	return err
}
