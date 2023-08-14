package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"golang.org/x/crypto/bcrypt"
	"../../shared/functions/createAWSResErr.go"
	"../../shared/functions/validationFunctions.go"
)

var dbClient = dynamodb.New(session.New())

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

func Handler() interface{} {
	return cors.New(cors.Config{
		Next: Signup,
	})
}

func GetSignupDate() int64 {
	return time.Now().Unix() / (24 * 60 * 60) * 24 * 60 * 60
}

func InsertUserToDB(username, email, password string, memberSince int64) error {
	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
			"isVerified": {
				BOOL: aws.Bool(false),
			},
			"memberSince": {
				N: aws.String(fmt.Sprint(memberSince)),
			},
			"numRatings": {
				N: aws.String("0"),
			},
			"password": {
				S: aws.String(password),
			},
			"username": {
				S: aws.String(username),
			},
        },
        TableName: aws.String(os.Getenv("USER_TABLE_NAME")),
        ReturnConsumedCapacity: aws.String("TOTAL"),
    }

    _, err := dbClient.PutItem(input)
    return err
}

