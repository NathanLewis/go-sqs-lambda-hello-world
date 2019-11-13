// snippet-comment:[These are tags for the AWS doc team's sample catalog. Do not remove.]
// snippet-sourceauthor:[Doug-AWS]
// snippet-sourcedescription:[Enables long polling on message receipt.]
// snippet-keyword:[Amazon Simple Queue Service]
// snippet-keyword:[Amazon SQS]
// snippet-keyword:[GetQueueUrl function]
// snippet-keyword:[ReceiveMessage function]
// snippet-keyword:[Go]
// snippet-sourcesyntax:[go]
// snippet-service:[sqs]
// snippet-keyword:[Code Sample]
// snippet-sourcetype:[full-example]
// snippet-sourcedate:[2018-03-16]
/*
   Copyright 2010-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.

   This file is licensed under the Apache License, Version 2.0 (the "License").
   You may not use this file except in compliance with the License. A copy of
   the License is located at

    http://aws.amazon.com/apache2.0/

   This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
   CONDITIONS OF ANY KIND, either express or implied. See the License for the
   specific language governing permissions and limitations under the License.
*/

package main

import (
    "flag"
    "fmt"
    "os"
    "encoding/xml"

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
        Path string
    }

    type SimpleAsset struct {
        XMLName xml.Name `xml:"simpleAsset"`
        ActivityId string  `xml:"activityId,attr"`  // notice the capitalized field name here and the `xml:"app_name,attr"`
        Uri Uri
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
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("eu-west-1")},
    )

    // Create a SQS service client.
    svc := sqs.New(sess)

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

    for {
        // Receive a message from the SQS queue with long polling enabled.
        result, _ := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
            QueueUrl: resultURL.QueueUrl,
            AttributeNames: aws.StringSlice([]string{
                "SentTimestamp",
            }),
            MaxNumberOfMessages: aws.Int64(1),
            MessageAttributeNames: aws.StringSlice([]string{
                "All",
            }),
            VisibilityTimeout:   aws.Int64(60),
            WaitTimeSeconds:     aws.Int64(0),
        })
        if len(result.Messages) > 0 {
            fmt.Printf("Received %d messages.\n", len(result.Messages))
            fmt.Printf("%T\n", result.Messages)
            fmt.Println(result.Messages)
            fmt.Printf("%T\n", result.Messages[0])
            fmt.Println(result.Messages[0])
            fmt.Printf("%s\n", *(result.Messages[0]).Body)
    /*
            var message = result.Message[0]
            fmt.Println(result.Message[0])
            var asset SimpleAsset
            err := xml.Unmarshal([]byte(message.Body), &asset)
            if err != nil {
                fmt.Printf("error: %v", err)
            } else {
                fmt.Printf("asset ID:: %q\n", asset.ActivityId)
            }
     */
        }
    }

}

func exitErrorf(msg string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, msg+"\n", args...)
    os.Exit(1)
}
