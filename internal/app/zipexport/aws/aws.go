package aws

import (
	"fmt"
	"time"

	"github.com/NathanLewis/go-sqs-lambda-hello-world/internal/app/zipexport/wormhole"
	"github.com/NathanLewis/go-sqs-lambda-hello-world/internal/pkg/zipexport/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"
)

var retryCount = 3

//AWS wrapper around wormhole functionality
type AWS struct {
	WormHole *wormhole.WormHole
	UgcCert  *util.UgcCert
}

//WormHoleCredentialsProvider used as the aws provider
type WormHoleCredentialsProvider struct {
	credentials.Provider
	WormHole *wormhole.WormHole
	Start    time.Time
}

//Retrieve used to retreive aws credentials
func (wm *WormHoleCredentialsProvider) Retrieve() (credentials.Value, error) {
	si := wm.WormHole.SessionInfo()

	log.Printf("--- RETRIEVE accessKeyId=%s, secretAccessKey=%s, sessionToken=%s", si.AccessKeyID, si.SecretAccessKey, si.SessionToken)
	value := credentials.Value{
		si.AccessKeyID,
		si.SecretAccessKey,
		si.SessionToken,
		"wormhole",
	}

	wm.Start = time.Now()
	return value, nil
}

//IsExpired check to see if credentials have expired
func (wm *WormHoleCredentialsProvider) IsExpired() bool {
	elapsed := wm.Start.Sub(time.Now())
	if elapsed.Minutes() >= 55 {
		return true
	}
	return false
}

//AwsSession used to create aws session
func (ugcAws *AWS) AwsSession(region string) *session.Session {

	config := util.Configuration{}
	config = config.Config()
	var awsRegion = config.Region
	if region != "" {
		awsRegion = region
	}
	//Chain Provider: First check ec2 role if not then use wormhole
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&WormHoleCredentialsProvider{WormHole: ugcAws.WormHole},
		})

	sess, err := session.NewSession(&aws.Config{
		Region:     aws.String(awsRegion),
		MaxRetries: aws.Int(retryCount),
		//LogLevel:    aws.LogLevel(aws.LogDebugWithHTTPBody),
		Credentials: creds},
	)

	if err != nil {
		log.WithFields(log.Fields{
			"Error": fmt.Sprintf("%+v", err),
		}).Fatal("WormHole uable to get to create the session")

	}

	return sess
}
