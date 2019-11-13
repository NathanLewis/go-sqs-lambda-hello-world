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
        ActivityId string   `xml:"activityId,attr"` // notice the capitalized field name here and the `xml:"app_name,attr"`
        Uri        Uri
    }

    var name string
    var timeout int64
    flag.StringVar(&name, "n", "", "Queue name")
    flag.Int64Var(&timeout, "t", 20, "(Optional) Timeout in seconds for long polling")
    flag.Parse()

    if len(name) == 0 {
        flag.PrintDefaults()
        exitErrorf("Queue name required")
    }

    // Initialize a session in eu-west-1 that the SDK will use to load
    // credentials from the shared credentials file ~/.aws/credentials.
    err, svc := setupSession()

    // Need to convert the queue name into a URL. Make the GetQueueUrl
    // API call to retrieve the URL. This is needed for receiving messages
    // from the queue.
    resultURL, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
        QueueName: aws.String(name),
    })
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
            exitErrorf("Unable to find queue %q.", name)
        }
        exitErrorf("Unable to queue %q, %v.", name, err)
    }

    inputChannel := make(chan string)
    go messagePoller(svc, resultURL, inputChannel)
    for {
        var message = <-inputChannel
        if 0 == len(message) {
            continue
        }
        var asset SimpleAsset
        err := xml.Unmarshal([]byte(message), &asset)
        if err != nil {
            fmt.Printf("error: %v\n", err)
        } else {
            fmt.Printf("asset ID:: %q\n", asset.ActivityId)
        }
    }
}

func setupSession() (error, *sqs.SQS) {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("eu-west-1")},
    )
    // Create a SQS service client.
    svc := sqs.New(sess)
    return err, svc
}

func exitErrorf(msg string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, msg+"\n", args...)
    os.Exit(1)
}
