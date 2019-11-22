// This file is a modified version of code from Amazon which they distributed under an Apache 2.0 license

package main

import (
    "flag"
    "fmt"
    "os"
    "strconv"
    "time"
)



// Receive messages from Queue.
//
// Usage:
//    go run sqs_handler.go -n queue_name -t timeout
func main() {

    var inputName, badMesgName, snsTopicArn string
    var timeout int64
    flag.StringVar(&inputName, "n", "", "Input Queue inputName")
    flag.StringVar(&badMesgName, "b", "", "Bad Message Queue inputName")
    flag.StringVar(&snsTopicArn, "i", "", "Sns Topic Arn")
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

    if len(snsTopicArn) == 0 {
        flag.PrintDefaults()
        exitErrorf("Sns Topic Arn required")
    }

    // Initialize a session in eu-west-1 that the SDK will use to load
    sqs, sess := setupSession()

    // Need to convert the queue inputName into a URL. Make the GetQueueUrl
    // API call to retrieve the URL. This is needed for receiving messages
    // from the queue.
    inputQueueURL := findQueueUrl(sqs, inputName)
    //fmt.Printf("Found %s\n", inputQueueURL)
    badMesgQueueURL := findQueueUrl(sqs, badMesgName)
    //fmt.Printf("Found %s\n", badMesgQueueURL)

    inputQueueChannel := make(chan string)
    badMesgQueueChannel := make(chan string)
    snsChannel := make(chan map[string]interface{})
    go messagePoller(sqs, inputQueueURL, inputQueueChannel)
    go sqsSender(sqs, badMesgQueueURL, badMesgQueueChannel)
    go snsSender(sess, snsTopicArn, snsChannel)

    for {
        var message = <-inputQueueChannel
        var asset Asset
        var snsMesgMap map[string]interface{}
        err := asset.readFromString(message)
        if err == nil {
            snsMesgMap = asset.toMap()
            snsMesgMap["event_name"] = "go.sqs.message.status.ok"
        } else {
            fmt.Printf("Bad Message: %s\n", message)
            badMesgQueueChannel <- message
            snsMesgMap = make(map[string]interface{})
            snsMesgMap["event_name"] = "go.sqs.message.status.badmessage"
            snsMesgMap["bad_message"] = message
        }
        // Add whatever fields you want / need to the sns message
        snsMesgMap["event_timestamp"] = strconv.FormatInt(int64(time.Now().Unix()), 10)

        snsChannel <- snsMesgMap
    }
}

func exitErrorf(msg string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, msg+"\n", args...)
    os.Exit(1)
}
