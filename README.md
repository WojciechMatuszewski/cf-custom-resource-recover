# Recovering from custom resource lambda failures

This repo contains an example of how one might recover from a custom resource that is backed by a lambda function failing.
This example does not use the `cfn.LambdaWrap` (similar wrappers exist for CDK) exposed by the `aws-lambda-go` package for learning purposes.

Not handling errors within your custom resource handler can be painful. _CloudFormation_ can take ages to start rollback because of your handler being invoked over and over again. Usually these situations are covered by the wrappers I mentioned above, but sometimes you cannot or do not want to use them.

This repo is based on [this blog post on _AWS Compute Blog_](https://aws.amazon.com/blogs/compute/adding-resiliency-to-aws-cloudformation-custom-resource-deployments/)

## Deployment

1. Ensure that `ShouldFail` property on the `AWS::CloudFormation::CustomResource` resource is set to `false`.
1. Run `sam build`
1. Run `sam-deploy --guided`
1. Switch the `ShouldFail` property on the `AWS::CloudFormation::CustomResource` to `true`.
1. Run `make deploy`
1. Observe that the custom resource was modified successfully, read the _CloudWatch logs_ of given functions and observe the _Lambda Destinations_ in action.

## Learnings

- Always do `sam build` before deploying 🤦‍♂️.

- There exist a whole different world of handling custom resources if you are not using the wrappers provided by the frameworks.

- When you screw up and not handle errors inside the custom resource the CloudFormation will take forever to detect that.

- Every property within the `ResourceProperties` the custom resource receives is a string?

  - In most of relevant CF templates, the `true` value is used as boolean but seem to be annotated as string, for example in the `Parameters` section.
  - There does not exist a `Boolean` type one can specify as a `Parameter`. [Checkout the documentation](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/parameters-section-structure.html#parameters-section-structure-properties)
  - There are `Boolean` types listed in the CF spec for various resource properties, for example the [AWS::S3::Bucket resource](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket.html#cfn-s3-bucket-objectlockenabled)
  - All of the above leads me to believe that the `Boolean` type listed in various CF resource properties is just a string that is coerced to a `Boolean` type.

- The payload that you get from _Lambda Destinations_ via SQS is not very friendly. The body is a _stringified_ blob.
  But I was able to parse it without any problems.

- This technique of having the _Lambda Destination_ is a godsend. Really works well if you do not want to use the wrappers other frameworks offer.
  Now the question is, why would not you use them?

  - I'm really not sure about the answer to this one. Maybe some kind of regulations? I'm not sure.

- Do not mistake the _Event Source Mapping_ with the _Event Invoke Config_.

  - The _Event Source Mapping_ is used when lambda reads events from various services, then invokes your function. [Here are the docs](https://docs.aws.amazon.com/lambda/latest/dg/invocation-eventsourcemapping.html).
  - The _Event Invoke Config_ is a CF configuration that relates to the _Lambda Destinations_.

- Remember that if your function is invoked _asynchronously_, by default, _Lambda service_ will retry your function two times.
  Underneath, _Lambda service_ uses _SQS_ for managing the invocation and retries.

- Overriding the default _Log retention_ setting for a lambda log group can be tricky

  - By default, the _Log retention_ is set to never expire. Whenever the lambda is invoked, a log group will be created for that lambda (if it does not yet exists). The log group name is following the `/aws/lambda/LAMBDA_NAME` scheme.

  - If you use _AWS CDK_ lambda construct a custom resource is used to create a log group with the name of your lambda **before** that lambada is created. You can then specify the retention period for that log group.

  - You **do not have to** use a custom resource to change the _Log retention_ period though. All you have to do is to create the log group resource yourself **before** your lambda is created. In _CloudFormation_ / _AWS SAM_ world that would be using the `DependsOn` property.
