from troposphere import Template,Ref,Parameter,Sub,GetAtt,Join, Output, Export
from troposphere.s3 import Rules as S3Key,Bucket
from troposphere.awslambda import Function, Code, Alias, Permission, EventSourceMapping
from troposphere.sqs import Queue, QueuePolicy, RedrivePolicy
import awacs
from troposphere.iam import Role, PolicyType
from awacs.aws import Action,Allow,Condition,Policy,PolicyDocument,Principal,Statement,Condition
from awacs.sts import AssumeRole
from troposphere import iam, awslambda

class LambdaStack():
    
    def generate_template(self, template):
        
        alias = Parameter("LambdaEnvAlias",Default="int",Description="Alias used to reference the lambda",Type="String" )
        template.add_parameter(aliasP)

        lambda_bucket = template.add_parameter(Parameter(
            "LambdaBucket",
            Type="String",
            Default="go-lambda-hello-world",
            Description="The S3 Bucket that contains the zip to bootstrap your "
                "lambda function"
        ))

        s3_key = template.add_parameter(Parameter(
            "S3Key",
            Type="String",
            Default="main.zip",
            Description="The S3 key that references the zip to bootstrap your "
                "lambda function"
        ))

        handler = template.add_parameter(Parameter(
            "LambdaHandler",
            Type="String",
            Default="event_handler.handler",
            Description="The name of the function (within your source code) "
                "that Lambda calls to start running your code."
        ))

        memory_size = template.add_parameter(Parameter(
            "LambdaMemorySize",
            Type="Number",
            Default="128",
            Description="The amount of memory, in MB, that is allocated to "
                "your Lambda function.",
            MinValue="128"
        ))

        timeout = template.add_parameter(Parameter(
            "LambdaTimeout",
            Type="Number",
            Default="300",
            Description="The function execution time (in seconds) after which "
                "Lambda terminates the function. "
        ))

        lambda_function = template.add_resource(Function(
            "LambdaGoHelloWorld",
            Code=Code(
                S3Bucket="go-lambda-hello-world",
                S3Key=Ref(s3_key)
            ),
            Description="Go function used to demonstate sqs integration",
            Handler=Ref(handler),
            Role=GetAtt("LambdaExecutionRole", "Arn"),
            Runtime="go1.x",
            MemorySize=Ref(memory_size),
            FunctionName="go-lambda-hello-world",
            Timeout=Ref(timeout)
        ))

        alias = template.add_resource(Alias(
            "GolLambdaAlias",
            Description="Alias for the go lambda",
            FunctionName=Ref(lambda_function),
            FunctionVersion="$LATEST",
            Name=Ref(alias)
        ))

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

    
if __name__ == '__main__':
    t = Template(Description="Go lambda Stack")
    t.add_version("2010-09-09")
    lambda_stack = LambdaStack()
    lambda_stack.generate_template(t)
    print(t.to_json())