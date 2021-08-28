package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/events"
	aws_lambda_go "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambda_types "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	aws_lambda_go.Start(handler)
}

type Body struct {
	RequestPayload cfn.Event `json:"requestPayload"`
}

func handler(ctx context.Context, event events.SQSEvent) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	svc := lambda.NewFromConfig(cfg)

	for _, record := range event.Records {
		var b Body

		err := json.Unmarshal([]byte(record.Body), &b)
		if err != nil {
			panic(err)
		}

		payload := b.RequestPayload
		payload.ResourceProperties["ShouldFail"] = "false"

		fmt.Println("Sending")
		spew.Dump(payload)

		buf, err := json.Marshal(payload)
		if err != nil {
			panic(err)
		}

		_, err = svc.Invoke(ctx, &lambda.InvokeInput{
			FunctionName:   aws.String(os.Getenv("FUNCTION_NAME")),
			InvocationType: lambda_types.InvocationTypeRequestResponse,
			LogType:        lambda_types.LogTypeTail,
			Payload:        buf,
		})
		if err != nil {
			panic(err)
		}
	}

	return nil
}
