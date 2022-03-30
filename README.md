# aws-lambda-loadtester

A multi-threaded AWS Lambda load tester. Uses the AWS SDK for authentication.

## Example

```sh
EXPORT AWS_REGION="us-east-1"
```

```sh
$ ./loadtest \
    --jobs=20 \
    --workers=10 \
    --function-name="example" \
    --payload=./payload.json
```

## Getting started

### Building the source code

```sh
$ go build
```

