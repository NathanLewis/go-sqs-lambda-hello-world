package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
)

type Parameter struct {
		Type			string
		Default			string
		AllowedValues	string
		Description		string
		MinValue        string
}

func main() {
	// Create a new CloudFormation template
	template := cloudformation.NewTemplate()

	
	aliasP := Parameter{Type: "String", Default: "int", Description: "Alias  used to generate the lambda"}
	template.Parameters["LambdaAlias"] = &aliasP
	
	lambdaBucketP := Parameter{Type: "String", Default: "go-lambda-hello-world", Description: "The S3 bucket that contains the zip to bootstap lambda function"}
	template.Parameters["LambdaBucket"] = &lambdaBucketP
	
	s3KeyP := Parameter{Type: "String", Default: "main.zip", Description: "The S3 key that references the zip to bootstrap your lambda function"}
	template.Parameters["S3Key"] = &s3KeyP
	
	handlerP := Parameter{Type: "String", Default: "main", Description: "The lambda function entry point"}
	template.Parameters["LambdaHandler"] = &handlerP

	memorySizeP := Parameter{Type: "Number", Default: "128", Description: "Amount of memory allocated for the lambda", MinValue: "128"}
	template.Parameters["LambdaMemorySize"] = &memorySizeP

	timeoutP := Parameter{Type: "Number", Default: "300", Description: "The function execution time"}
	template.Parameters["LambdaTimeout"] = &timeoutP
	

	template.Resources["GoLambdaAlias"] = &resources.AWSLambdaAlias{Description:"Alias for the go lambda", FunctionName: "", FunctionVersion: "$LATEST", Name: cloudformation.Ref("LambdaAlias")}
	template.Resources["LambdaGoHelloWorld"] = &resources.AWSServerlessFunction{
		Description: "Go function used to demonstrate sqs integration",
		/*
		Code: &resources.AWSServerlessFunction_CodeUri{
			S3Location: &resources.AWSServerlessFunction_S3Location{
					Bucket: "go-lambda-hellow-world",
					Key: "cloudformation.Ref(\"S3Key\")"
			}
		},
		*/
		Handler: cloudformation.Ref("LambdaParameter"),
		Role: cloudformation.GetAtt("LambdaExecutionRole","Arn"),
		Runtime: "go1.x",
		//MemorySize: cloudformation.Ref("LambdaMemorySize"),
		FunctionName: "go-lambda-hello-world",
		Timeout: 1000}

		/*

	teplate.Resources["LambdaExecutionRole"] = &resources.AWSIAMUser_Policy{Guide/aws-properties-iam-policy.html#cfn-iam-policies-policydocument
			PolicyDocument: &resources.AWSIAMPolicy, 
			PolicyName: "GoFunctionRolePolicy"}

		

		cacheContent := map[string]interface{}{
			"Statement": map[string]interface{}{
				"Effect": "Allow",
				"Action": [ "logs": "CreateLogStream",
					      "logs": "PutLogEvents"
				],
				"Resource": ["arn:aws:logs:*:*:*"]
			},
		}

		"Statement": [
			{
				"Action": [
					"logs:CreateLogGroup",
					"logs:CreateLogStream",
					"logs:PutLogEvents"
				],
				"Effect": "Allow",
				"Resource": [
					"arn:aws:logs:*:*:*"
				]
			},
			{
				"Action": [
					"sqs:ChangeMessageVisibility",
					"sqs:DeleteMessage",
					"sqs:GetQueueAttributes",
					"sqs:ReceiveMessage"
				],
				"Effect": "Allow",
				"Resource": [
					{
						"Fn::GetAtt": [
							"GoLambdaQueue",
							"Arn"
						]
					}
				]
			}
		]
	},
		
		Statement = [Statement(
			Effect=Allow,
			Action = [ Action("logs", "CreateLogGroup"),
				Action("logs", "CreateLogStream"),
				Action("logs", "PutLogEvents")
			],
			Resource = ["arn:aws:logs:*:*:*"]
		),
		Statement(
			Effect= Allow,
			Action = [ Action("sqs","ChangeMessageVisibility"),
				Action("sqs","DeleteMessage"),
				Action("sqs","GetQueueAttributes"),
				Action("sqs","ReceiveMessage")
			],
			Resource = [GetAtt(go_helloworld_queue,"Arn")]
	)])

		{
			"Version": "2012-10-17",
			"Statement": [
			  {
				"Effect": "Allow",
				"Action": [
				  "s3:ListAllMyBuckets",
				  "s3:GetBucketLocation"
				],
				"Resource": "arn:aws:s3:::*"
			  },
			  {
				"Effect": "Allow",
				"Action": "s3:ListBucket",
				"Resource": "arn:aws:s3:::BUCKET-NAME",
				"Condition": {"StringLike": {"s3:prefix": [
				  "",
				  "home/",
				  "home/${aws:username}/"
				]}}
			  },
			  {
				"Effect": "Allow",
				"Action": "s3:*",
				"Resource": [
				  "arn:aws:s3:::BUCKET-NAME/home/${aws:username}",
				  "arn:aws:s3:::BUCKET-NAME/home/${aws:username}/*"
				]
			  }
			]
		  }
	template.Resources["LambdaExecutionRole"] = &resources.AWSIAMRole {

		Policies: [&resources.AWSIAMRole_Policy {
			PolicyName: "GoFunctionRolePolicy",
			PolicyDocument: &resources.

		}]

		AssumeRolePolicyDocument:
	}

	/*
	
	Environment: &cloudformation.AWSLambdaFunction_Environment{
      Variables: map[string]string{
          "CLUSTER": cluster.Cluster.Name,
          "ASGPREFIX": Sub("${ClusterName}-asg-${Version}"),
          "REGION": Ref("AWS::Region"),
  },
},

	dead_letter_queue = template.add_resource(Queue(
		"GoLambdaDeadLetterQueue",
		QueueName=("golambdaqueue-dlq"),
		VisibilityTimeout=30,
		MessageRetentionPeriod=1209600,
		MaximumMessageSize=262144,
		DelaySeconds=0,
		ReceiveMessageWaitTimeSeconds=0
	))

	go_helloworld_queue = template.add_resource(Queue(
		"GoLambdaQueue",
		QueueName=("golambdaqueue"),
		VisibilityTimeout=1800,
		RedrivePolicy=RedrivePolicy(
			deadLetterTargetArn=GetAtt(dead_letter_queue,"Arn"),
			maxReceiveCount=3
			)
	))

	lambda_execution_role = template.add_resource(
		Role(
			"LambdaExecutionRole",
			Policies=[iam.Policy(
						PolicyName="GoFunctionRolePolicy",
						PolicyDocument= Policy(
							Statement = [Statement(
									Effect=Allow,
									Action = [ Action("logs", "CreateLogGroup"),
										Action("logs", "CreateLogStream"),
										Action("logs", "PutLogEvents")
									],
									Resource = ["arn:aws:logs:*:*:*"]
								),
								Statement(
									Effect= Allow,
									Action = [ Action("sqs","ChangeMessageVisibility"),
										Action("sqs","DeleteMessage"),
										Action("sqs","GetQueueAttributes"),
										Action("sqs","ReceiveMessage")
									],
									Resource = [GetAtt(go_helloworld_queue,"Arn")]
							)])
			)],
			AssumeRolePolicyDocument=Policy(
				Statement = [
					Statement(
						Effect = Allow,
						Action = [AssumeRole],
						Principal=Principal("Service", ["lambda.amazonaws.com"])
					)
				]
			)
		)
	)  

	template.add_resource(awslambda.Permission(
		'QueueInvokePermission',
		FunctionName=GetAtt(lambda_function, 'Arn'),
		Action="lambda:InvokeFunction",
		Principal= "sqs.amazonaws.com",
		SourceArn=GetAtt(go_helloworld_queue, 'Arn'),
	))

	template.add_output(Output(
		"LambdaHelloWorldQueue",
		Value=GetAtt(go_helloworld_queue,"Arn"),
		Export=Export("lambda-go-hello-world-queue"),
		Description="Arn of the queue"
	))

	template.add_output(Output(
		"LambdaHelloWorlFunction",
		Value=Ref(alias),
		Export=Export("lambda-go-hello-world-function"),
		Description="Arn of the function"
	))
	*/
		// Create an Amazon SNS topic, with a unique name based off the current timestamp
		template.Resources["MyTopic"] = &resources.AWSSNSTopic{
			TopicName: "my-topic-" + strconv.FormatInt(time.Now().Unix(), 10),
		}
	
		// Create a subscription, connected to our topic, that forwards notifications to an email address
		template.Resources["MyTopicSubscription"] = &resources.AWSSNSSubscription{
			TopicArn: cloudformation.Ref("MyTopic"),
			Protocol: "email",
			Endpoint: "some.email@example.com",
		}
	

		// Let's see the JSON AWS CloudFormation template
		j, err := template.JSON()
		if err != nil {
			fmt.Printf("Failed to generate JSON: %s\n", err)
		} else {
			fmt.Printf("%s\n", string(j))
		}
	
		// and also the YAML AWS CloudFormation template
		//y, err := template.YAML()
		//if err != nil {
		//	fmt.Printf("Failed to generate YAML: %s\n", err)
		//} else {
	//		fmt.Printf("%s\n", string(y))
	//	}
	
}