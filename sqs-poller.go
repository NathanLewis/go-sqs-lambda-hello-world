// This file is a modified version of code from Amazon which they distributed under an Apache 2.0 license

package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sqs"
    "os"
)



// Receive message from Queue with long polling enabled.
//
// Usage:
//    go run sqs_handler.go -n queue_name -t timeout
func main() {

    var inputName, badMesgName string
    var timeout int64
    flag.StringVar(&inputName, "n", "", "Input Queue inputName")
    flag.StringVar(&badMesgName, "b", "", "Bad Message Queue inputName")
    flag.Int64Var(&timeout, "t", 20, "(Optional) Timeout in seconds for long polling")
    flag.Parse()

    if len(inputName) == 0 {
        flag.PrintDefaults()
        exitErrorf("Input Queue Name or Arn required")
    }
    if len(badMesgName) == 0 {
        flag.PrintDefaults()
        exitErrorf("Bad Message Queue Name or Arn required")
    }

    // Initialize a session in eu-west-1 that the SDK will use to load
    svc := setupSession()

    // Need to convert the queue inputName into a URL. Make the GetQueueUrl
    // API call to retrieve the URL. This is needed for receiving messages
    // from the queue.
    inputQueueURL := findQueueUrl(svc, inputName)
    fmt.Printf("Found %s\n", inputQueueURL)
    badMesgQueueURL := findQueueUrl(svc, badMesgName)
    fmt.Printf("Found %s\n", badMesgQueueURL)

    inputQueue := make(chan string)
    badMesgQueue := make(chan string)
    go messagePoller(svc, inputQueueURL, inputQueue)
    go messageSender(svc, badMesgQueueURL, badMesgQueue)

    for {
        var message = <-inputQueue
        var asset Asset
        err := asset.readFromString(message)
        if err != nil {
            fmt.Printf("Bad Message: %v\n", err)
            badMesgQueue <- message
        } else {
            asset.printFields()
            result, err := json.Marshal(asset.toMap())
            if nil != err {
                fmt.Println("Error marshalling to JSON", err)
            } else {
                fmt.Println(string(result))
            }
        }
    }
}

func setupSession() *sqs.SQS {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("eu-west-1")},
    )
    if err != nil {
        exitErrorf("Unable to setup Session")
    }

    // Create a SQS service client.
    svc := sqs.New(sess)
    return svc
}

func exitErrorf(msg string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, msg+"\n", args...)
    os.Exit(1)
}
