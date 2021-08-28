package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event cfn.Event) (string, error) {
	spew.Dump(event)

	shouldFail, found := event.ResourceProperties["ShouldFail"].(string)
	if !found {
		panic("ShouldFail property not found or is of wrong type (asserted string)")
	}

	fmt.Printf("%v", shouldFail)

	if shouldFail == "true" {
		panic("Failing because of the ShouldFail")
	}

	r := cfn.NewResponse(&event)
	r.Status = cfn.StatusSuccess
	// Uses `init` to populate these variables from environment variables
	r.PhysicalResourceID = lambdacontext.LogStreamName

	err := r.Send()
	if err != nil {
		return err.Error(), err
	}

	return "", nil
}
