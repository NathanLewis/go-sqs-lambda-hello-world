# go-sqs-lambda-hello-world
An attempt at making a simple sqs polling lambda in go

### Build Lambda

`make build`

### AWS Infrastructure

1. Create python virtual env

   `make package`

   `source ./venv/bin/activate`

2. Create the deployment bucket stack
   `cd scripts/python/src`
   `python -m helloworld.bucketenv aws_account_id aws_region`

3. Upload zip to bucket
   `./deploy.sh`

4. Create the stack for the function
   `cd scripts/python/src`
   `python -m helloworld.funcenv aws_account_id aws_region`



### Development

After making changes to the code and building the zip. Use the following script to update the lambda.

`./update.sh`

### Note

Create EventSourceMapping is not done via cloudformation 