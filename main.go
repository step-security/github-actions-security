package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Handler struct {
}

func (h Handler) Invoke(ctx context.Context, req []byte) ([]byte, error, bool) {

	httpRequest := &events.APIGatewayV2HTTPRequest{}

	err := json.Unmarshal([]byte(req), &httpRequest)

	if err == nil && httpRequest.RawPath != "" {

		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		dynamoDbSvc := dynamodb.New(sess)
		var response events.APIGatewayProxyResponse

		if httpRequest.RequestContext.HTTP.Method == "OPTIONS" {
			response = events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
			}
			returnValue, _ := json.Marshal(&response)
			return returnValue, nil, false
		}

		if strings.Contains(httpRequest.RawPath, "/secure-workflow") {

			fixResponse, err, metricsVar := SecureWorkflow(httpRequest.QueryStringParameters, httpRequest.Body, dynamoDbSvc)

			if err != nil {
				response = events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       err.Error(),
				}
			} else {

				output, _ := json.Marshal(fixResponse)
				response = events.APIGatewayProxyResponse{
					StatusCode: http.StatusOK,
					Body:       string(output),
				}
			}

		}

		returnValue, _ := json.Marshal(&response)
		return returnValue, nil, metricsVar

	}

	return nil, fmt.Errorf("request was neither APIGatewayV2HTTPRequest nor SQSEvent"), false
}

func main() {
	lambda.StartHandler(Handler{})
}
