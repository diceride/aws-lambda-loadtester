package main

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/lambda"

    "errors"
    "flag"
    "io/ioutil"
    "log"
    "os"
    "strconv"
    "sync"
)

func worker(queue chan string, wg *sync.WaitGroup, client *lambda.Lambda, request []byte) {
    defer wg.Done()

    for id := range queue {
        log.Printf("[request][%s]: start", id)

        response, err := client.Invoke(&lambda.InvokeInput{
            FunctionName: aws.String("test-ramen-runtime"),
            Payload: request,
        })
        if err != nil {
            log.Printf("client.Invoke: %v", err.Error())
        }

        if response.FunctionError == nil && *response.StatusCode == int64(200) {
            log.Printf("[request][%s]: success", id)
        } else {
            log.Printf("[request][%s]: failed: status code %d\n%s", id, *response.StatusCode, string(response.Payload))
        }
    }
}

func main() {
    var (
        workers int
        jobs int
        functionName string
        payload string
    )

	flag.IntVar(&workers, "workers", 100, "the amount of workers to start")
	flag.IntVar(&jobs, "jobs", 100, "the amount of jobs to run")
	flag.StringVar(&functionName, "function-name", "", "the AWS Lambda function name")
	flag.StringVar(&payload, "payload", "", "the AWS Lambda payload as a JSON filepath")
	flag.Parse()

	if (functionName == "") {
	    panic(errors.New("option --function-name can not be empty"))
	}

	if (payload == "") {
	    panic(errors.New("option --payload can not be empty"))
	}

	request, err := ioutil.ReadFile(payload)
    if err != nil {
      log.Panicf("ioutil.ReadFile: %v", err.Error())
    }

    var test bool
	accessKeyID, test := os.LookupEnv("AWS_ACCESS_KEY_ID")
	secretAccessKey, test := os.LookupEnv("AWS_SECRET_ACCESS_KEY")

	profile, test := os.LookupEnv("AWS_PROFILE")
	if !test {
		profile = "default"
	}

	region, test := os.LookupEnv("AWS_REGION")
	if !test {
		panic(errors.New("environment variable AWS_REGION is not set"))
	}

    // Create a new AWS configuration
    config := aws.Config{
        Region: aws.String(region),
    }

    if (accessKeyID == "" && secretAccessKey == "") {
        config.Credentials = credentials.NewSharedCredentials("", profile)
    } else {
        config.Credentials = credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
    }

    wg := new(sync.WaitGroup)
    queue := make(chan string, jobs)

    log.Printf("starting worker pool with %d workers", workers)

    // Worker pool
    for i := 0; i < workers; i++ {
        wg.Add(1)

        // Create a new AWS session
        session := session.Must(session.NewSessionWithOptions(session.Options{
            SharedConfigState: session.SharedConfigEnable,
        }))

        // Create a new AWS Lambda client
        client := lambda.New(session, &config)

        go worker(queue, wg, client, request)
    }

    // Send jobs to workers
    for i := 0; i < jobs; i++ {
        queue <- strconv.Itoa(i)
    }

    close(queue)
    wg.Wait()
}
