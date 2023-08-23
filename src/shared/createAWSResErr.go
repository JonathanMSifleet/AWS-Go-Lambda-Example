package shared

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type IHTTP struct {
	Headers    *Headers `json:"headers,omitempty"`
	StatusCode int      `json:"statusCode"`
	Body       *string  `json:"body,omitempty"`
}

type Headers struct {
	AccessControlAllowOrigin      string `json:"Access-Control-Allow-Origin"`
	AccessControlAllowCredentials bool   `json:"Access-Control-Allow-Credentials"`
}

func createAWSResErr(statusCode int, message interface{}) (IHTTP, error) {
	var msg string
	switch v := message.(type) {
	case string:
		msg = v
	case []string:
		if err := logErrors(v); err != nil {
			return IHTTP{}, err
		}
		msg = strings.Join(v, "\n")
	default:
		return IHTTP{}, fmt.Errorf("invalid message type")
	}

	res := IHTTP{
		StatusCode: statusCode,
		Body:       aws.String(fmt.Sprintf(`{"statusCode": %d, "message": "%s"}`, statusCode, msg)),
	}
	return res, nil
}

func logErrors(errors []string) error {
	fmt.Println("Errors:")
	for i, element := range errors {
		fmt.Printf("%d) %s\n", i, element)
	}
	return nil
}
