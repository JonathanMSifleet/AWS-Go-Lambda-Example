package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/badoux/checkmail"
)

func validateUserInputs(username string, email string) []error {
	var errors []error
	errors = append(errors, checkUniqueAttribute("email", email, "email"))
	errors = append(errors, checkUniqueAttribute("username", username, ""))

	validateInput(username, "Username")
	for _, err := range errors {
		errors = append(errors, err)
	}
	validateInput(email, "Email")
	for _, err := range errors {
		errors = append(errors, err)
	}

	return errors
}

func validateInput(value string, valueName string) []error {
	var errors []error

	switch valueName {
	case "Bio":
		errors = append(errors, validateLength(value, valueName, 0, 1000))
	case "Email":
		errors = append(errors,
			validateNotEmpty(value, valueName),
			validateLength(value, valueName, 3, 256),
			validateIsEmail(value))
	case "FirstName", "LastName":
		errors = append(errors,
			validateLength(value, valueName, 0, 32),
			validateAgainstRegex(
				value,
				valueName,
				regexp.MustCompile(`[!"`+"`"+`'#%&,:;<>=@{}~$()*+/\\?[\]^|0-9]`),
				"cannot contain numbers or special characters"))
	case "Password":
		errors = append(errors,
			validateNotEmpty(value, valueName),
			validateLength(value, valueName, 8, 32))
	case "Rating":
		errors = append(errors,
			validateIsWholeNumber(value, valueName))
	case "Review":
		errors = append(errors,
			validateLength(value, valueName, 0, 500))
	case "Username":
		errors = append(errors,
			validateNotEmpty(value, valueName),
			validateLength(value, valueName, 3, 16),
			validateAgainstRegex(
				value,
				valueName,
				regexp.MustCompile(`[^A-Za-z0-9]+`),
				"cannot contain special characters"))
	}

	return errors
}

func validateAgainstRegex(value string, name string, regex *regexp.Regexp, message string) error {
	if regex.MatchString(value) {
		return fmt.Errorf("%s %s", name, message)
	}
	return nil
}

func validateIsEmail(value string) error {
	if err := checkmail.ValidateFormat(value); err != nil {
		return fmt.Errorf("Email must be valid")
	}
	return nil
}

func validateIsWholeNumber(value string, name string) error {
	if _, err := strconv.Atoi(value); err != nil {
		return fmt.Errorf("%s must be a whole number", name)
	}
	return nil
}

func validateLength(value string, valueName string, min int64, max int64) error {
	if int64(len(value)) < min || int64(len(value)) > max {
		return fmt.Errorf("%s must be between %d and %d characters", valueName)
	}
	return nil
}

func validateNotEmpty(value string, name string) error {
	if len(strings.TrimSpace(value)) == 0 {
		return fmt.Errorf("%s must not be empty", name)
	}
	return nil
}

func validateNotValue(value string, name string, match string) error {
	if value == match {
		return fmt.Errorf("%s must not be %s", name)
	}
	return nil
}

func validateValue(value float64, name string, lowerBound float64, upperBound float64) string {
	if value < lowerBound || value > upperBound {
		return fmt.Sprintf("%s must be between %f and %f", name)
	}
	return ""
}

func validateWholeNumber(value float64, name string) error {
	if math.Mod(float64(int64(value)), float64(1)) != 0.0 {
		return fmt.Errorf("%s must be a whole number", name)
	}
	return nil
}

func checkUniqueAttribute(keyName string, keyValue string, indexName string) error {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return err
	}

	client := dynamodb.NewFromConfig(cfg)

	input := &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("USER_TABLE_NAME")),
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :v1", keyName)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v1": &types.AttributeValueMemberS{Value: keyValue},
		},
	}

	if indexName != "" {
		input.IndexName = aws.String(indexName)
	}

	result, err := client.Query(context.Background(), input)
	if err != nil {
		return err
	}

	if len(result.Items) > 0 {
		return fmt.Errorf("%s already in use", alphabeticalizeFirstChar(keyName))
	}

	return nil
}

func alphabeticalizeFirstChar(input string) string {
	return strings.ToUpper(string(input[0])) + input[1:]
}
