// This file is a modified version of code from Amazon which they distributed under an Apache 2.0 license

package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "time"
    "os"
)



// Receive message from Queue with long polling enabled.
//
// Usage:
//    go run sqs_handler.go -n queue_name -t timeout
func main() {

    var inputName, badMesgName, snsTopicArn string
    var timeout int64
    flag.StringVar(&inputName, "n", "", "Input Queue inputName")
    flag.StringVar(&badMesgName, "b", "", "Bad Message Queue inputName")
    flag.StringVar(&snsTopicArn, "i", "", "Ispy Topic Arn")
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
    fmt.Printf("Found %s\n", inputQueueURL)
    badMesgQueueURL := findQueueUrl(sqs, badMesgName)
    fmt.Printf("Found %s\n", badMesgQueueURL)

    inputQueueChannel := make(chan string)
    badMesgQueueChannel := make(chan string)
    snsChannel := make(chan string)
    go messagePoller(sqs, inputQueueURL, inputQueueChannel)
    go sqsSender(sqs, badMesgQueueURL, badMesgQueueChannel)
    go snsSender(sess, snsTopicArn, snsChannel)

    for {
        var message = <-inputQueueChannel
        var asset Asset
        err := asset.readFromString(message)
        if err != nil {
            fmt.Printf("Bad Message: %v\n", err)
            badMesgQueueChannel <- message
        } else {
            asset.printFields()
            snsMesgMap := asset.toMap()
            // Add whatever fields you want / need to the sns message
            snsMesgMap["event_name"] = "go.sqs.message.status.ok"
            snsMesgMap["event_timestamp"] = string(int64(time.Now().Unix()))
            result, err := json.Marshal(snsMesgMap)
            if nil != err {
                fmt.Println("Error marshalling to JSON", err)
                fmt.Println(message)
            } else {
                sresult := string(result)
                fmt.Println(sresult)
                snsChannel <- sresult
            }
        }
    }
}

func exitErrorf(msg string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, msg+"\n", args...)
    os.Exit(1)
}
