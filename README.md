# go-sqs-lambda-hello-world
A simple sqs lambda in go

### Build for ec2 
GOOS=linux GOARCH=amd64 go build

### Build Lambda

`$ > make build`

### AWS Infrastructure

1. Create python virtual env

   `$ > make package`

   `$ > source ./venv/bin/activate`

2. Create the deployment bucket stack
   `$ > cd scripts/python/src`
   `$ > python -m helloworld.bucketenv aws_account_id aws_region`

3. Upload zip to bucket
   `$ > ./deploy.sh`

4. Create the stack for the function
   `$ > cd scripts/python/src`
   `$ > python -m helloworld.funcenv aws_account_id aws_region`



### Development

After making changes to the code and building the zip. Use the following script to update the lambda.

`$ > ./update.sh`



### Publishing To Queue

The following script can be used to test out the functionality.

`$ > cd scripts/python/src`

`$ > python -m helloworld.publish.py aws_account_d aws_region`

### Note

aws cli is used to `CreateEventSourceMapping` instead of doing it via cloudformation 
