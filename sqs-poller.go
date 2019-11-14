// This file is a modified version of code from Amazon which they distributed under an Apache 2.0 license

package main

import (
    "encoding/xml"
    "flag"
    "fmt"
    "os"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/awserr"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sqs"
)

// Receive message from Queue with long polling enabled.
//
// Usage:
//    go run sqs_longpolling_receive_message.go -n queue_name -t timeout
func main() {
    type Uri struct {
        XMLName xml.Name `xml:"uri"`
        Path    string
    }

    type SimpleAsset struct {
        XMLName    xml.Name `xml:"simpleAsset"`
        ActivityId string   `xml:"activityId,attr"` // notice the capitalized field inputName here and the `xml:"app_name,attr"`
        Uri        Uri
    }

    var inputName, badMesgName string
    var timeout int64
    flag.StringVar(&inputName, "n", "", "Input Queue inputName")
    flag.StringVar(&badMesgName, "b", "", "Bad Message Queue inputName")
    flag.Int64Var(&timeout, "t", 20, "(Optional) Timeout in seconds for long polling")
    flag.Parse()

    if len(inputName) == 0 {
        flag.PrintDefaults()
        exitErrorf("Input Queue Name required")
    }
    if len(badMesgName) == 0 {
        flag.PrintDefaults()
        exitErrorf("Bad Message Queue Name required")
    }

    // Initialize a session in eu-west-1 that the SDK will use to load
    svc := setupSession()

    // Need to convert the queue inputName into a URL. Make the GetQueueUrl
    // API call to retrieve the URL. This is needed for receiving messages
    // from the queue.
    inputQueueURL := findQueueUrl(svc, inputName)
    badMesgQueueURL := findQueueUrl(svc, badMesgName)

    inputChannel := make(chan string)
    badMesgChannel := make(chan string)
    go messagePoller(svc, inputQueueURL, inputChannel)
    go messageSender(svc, badMesgQueueURL, badMesgChannel)

    for {
        var message = <-inputChannel
        var asset SimpleAsset
        err := xml.Unmarshal([]byte(message), &asset)
        if err != nil {
            fmt.Printf("Bad Message: %v\n", err)
            badMesgChannel <- message
        } else {
            fmt.Printf("asset ID:: %q\n", asset.ActivityId)
        }
    }
}

func findQueueUrl(svc *sqs.SQS, queueName string) *sqs.GetQueueUrlOutput {
    queueURL, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
        QueueName: aws.String(queueName),
    })
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
            exitErrorf("Unable to find queue %q.", queueName)
        }
        exitErrorf("Unable to queue %q, %v.", queueName, err)
    }
    return queueURL
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
